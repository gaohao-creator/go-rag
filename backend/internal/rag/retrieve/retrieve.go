package retrieve

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"unicode/utf8"

	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
	domainrepo "github.com/gaohao-creator/go-rag/internal/domain/repository"
	domainservice "github.com/gaohao-creator/go-rag/internal/domain/service"
)

// Retrieve 负责内容检索、降级检索、QA 检索合并和 rerank。
type Retrieve struct {
	contentRetriever  domainservice.Retriever
	fallbackRetriever domainservice.Retriever
	qaRetriever       domainservice.Retriever
	resultReranker    domainservice.Reranker
}

// NewRetrieve 创建统一检索门面。
func NewRetrieve(
	documentRepo domainrepo.DocumentRepository,
	chunkRepo domainrepo.ChunkRepository,
	content domainservice.Retriever,
	qa domainservice.Retriever,
	reranker domainservice.Reranker,
) *Retrieve {
	var fallbackRetriever domainservice.Retriever
	if documentRepo != nil && chunkRepo != nil {
		fallbackRetriever = &databaseRetriever{
			documentRepo: documentRepo,
			chunkRepo:    chunkRepo,
		}
	}
	return &Retrieve{
		contentRetriever:  content,
		fallbackRetriever: fallbackRetriever,
		qaRetriever:       qa,
		resultReranker:    reranker,
	}
}

// Retrieve 执行主检索、QA 检索、合并去重和重排流程。
func (r *Retrieve) Retrieve(ctx context.Context, req domainservice.RetrieveRequest) ([]domainmodel.RetrievedChunk, error) {
	chunks, qaChunks, err := r.retrieveChunks(ctx, req)
	if err != nil {
		return nil, err
	}
	chunks = mergeRetrievedChunks(chunks, qaChunks)
	chunks = r.applyRerank(ctx, req, chunks)
	sortRetrievedChunks(chunks)
	if req.TopK > 0 && len(chunks) > req.TopK {
		chunks = chunks[:req.TopK]
	}
	return chunks, nil
}

func (r *Retrieve) retrieveChunks(
	ctx context.Context,
	req domainservice.RetrieveRequest,
) ([]domainmodel.RetrievedChunk, []domainmodel.RetrievedChunk, error) {
	if r.qaRetriever == nil {
		chunks, err := r.retrieveContentChunks(ctx, req)
		return chunks, nil, err
	}

	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		wg         sync.WaitGroup
		chunks     []domainmodel.RetrievedChunk
		qaChunks   []domainmodel.RetrievedChunk
		contentErr error
	)

	wg.Add(1)
	go func() {
		defer wg.Done()
		chunks, contentErr = r.retrieveContentChunks(childCtx, req)
		if contentErr != nil {
			cancel()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		qaResult, err := r.qaRetriever.Retrieve(childCtx, req)
		if err != nil {
			return
		}
		qaChunks = qaResult
	}()

	wg.Wait()
	if contentErr != nil {
		return nil, nil, contentErr
	}
	return chunks, qaChunks, nil
}

func (r *Retrieve) retrieveContentChunks(
	ctx context.Context,
	req domainservice.RetrieveRequest,
) ([]domainmodel.RetrievedChunk, error) {
	primary := r.contentRetriever
	fallback := r.fallbackRetriever
	if primary == nil {
		primary = fallback
		fallback = nil
	}
	if primary == nil {
		return nil, fmt.Errorf("检索器未配置")
	}

	chunks, err := primary.Retrieve(ctx, req)
	if fallback == nil {
		return chunks, err
	}
	if err == nil && len(chunks) > 0 {
		return chunks, nil
	}
	return fallback.Retrieve(ctx, req)
}

func (r *Retrieve) applyRerank(
	ctx context.Context,
	req domainservice.RetrieveRequest,
	chunks []domainmodel.RetrievedChunk,
) []domainmodel.RetrievedChunk {
	if r.resultReranker == nil || len(chunks) == 0 {
		return chunks
	}
	reranked, err := r.resultReranker.Rerank(ctx, domainservice.RerankInput{
		Question: req.Question,
		TopK:     req.TopK,
		Chunks:   chunks,
	})
	if err != nil {
		return chunks
	}
	return reranked
}

type databaseRetriever struct {
	documentRepo domainrepo.DocumentRepository
	chunkRepo    domainrepo.ChunkRepository
}

func (r *databaseRetriever) Retrieve(ctx context.Context, req domainservice.RetrieveRequest) ([]domainmodel.RetrievedChunk, error) {
	documents, _, err := r.documentRepo.ListByKnowledgeBase(ctx, req.KnowledgeName, 1, 1000)
	if err != nil {
		return nil, err
	}
	results := make([]domainmodel.RetrievedChunk, 0)
	for _, document := range documents {
		chunks, _, err := r.chunkRepo.ListByDocumentID(ctx, document.ID, 1, 1000)
		if err != nil {
			return nil, err
		}
		for _, chunk := range chunks {
			score := scoreContentMatch(req.Question, chunk.Content)
			if score < req.Score {
				continue
			}
			results = append(results, domainmodel.RetrievedChunk{
				KnowledgeDocID: chunk.KnowledgeDocID,
				ChunkID:        chunk.ChunkID,
				Content:        chunk.Content,
				Ext:            chunk.Ext,
				Score:          score,
			})
		}
	}
	sortRetrievedChunks(results)
	if req.TopK > 0 && len(results) > req.TopK {
		results = results[:req.TopK]
	}
	return results, nil
}

func mergeRetrievedChunks(groups ...[]domainmodel.RetrievedChunk) []domainmodel.RetrievedChunk {
	if len(groups) == 0 {
		return nil
	}
	merged := make(map[string]domainmodel.RetrievedChunk)
	order := make([]string, 0)
	for _, group := range groups {
		for _, chunk := range group {
			existing, ok := merged[chunk.ChunkID]
			if !ok {
				merged[chunk.ChunkID] = chunk
				order = append(order, chunk.ChunkID)
				continue
			}
			if chunk.Score > existing.Score {
				merged[chunk.ChunkID] = chunk
			}
		}
	}
	result := make([]domainmodel.RetrievedChunk, 0, len(order))
	for _, chunkID := range order {
		result = append(result, merged[chunkID])
	}
	return result
}

func sortRetrievedChunks(chunks []domainmodel.RetrievedChunk) {
	sort.Slice(chunks, func(i, j int) bool {
		return chunks[i].Score > chunks[j].Score
	})
}

func scoreContentMatch(question string, content string) float64 {
	question = strings.TrimSpace(question)
	content = strings.TrimSpace(content)
	if question == "" || content == "" {
		return 0
	}
	if strings.Contains(content, question) {
		return 1
	}
	questionTerms := splitQuestionTerms(question)
	if len(questionTerms) == 0 {
		return 0
	}
	matched := 0
	for _, term := range questionTerms {
		if strings.Contains(content, term) {
			matched++
		}
	}
	return float64(matched) / float64(len(questionTerms))
}

func splitQuestionTerms(question string) []string {
	fields := strings.Fields(question)
	if len(fields) > 1 {
		return fields
	}
	terms := make([]string, 0, utf8.RuneCountInString(question))
	seen := make(map[string]struct{})
	for _, r := range question {
		term := strings.TrimSpace(string(r))
		if term == "" {
			continue
		}
		if _, ok := seen[term]; ok {
			continue
		}
		seen[term] = struct{}{}
		terms = append(terms, term)
	}
	return terms
}
