package service

import (
	"context"
	"sort"
	"strings"
	"unicode/utf8"

	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
	domainrepo "github.com/gaohao-creator/go-rag/internal/domain/repository"
	domainservice "github.com/gaohao-creator/go-rag/internal/domain/service"
)

type databaseRetrieverEngine struct {
	documentRepo domainrepo.DocumentRepository
	chunkRepo    domainrepo.ChunkRepository
}

func NewDatabaseRetrieverEngine(documentRepo domainrepo.DocumentRepository, chunkRepo domainrepo.ChunkRepository) domainservice.Retriever {
	return &databaseRetrieverEngine{
		documentRepo: documentRepo,
		chunkRepo:    chunkRepo,
	}
}

func (e *databaseRetrieverEngine) Retrieve(ctx context.Context, req domainservice.RetrieveRequest) ([]domainmodel.RetrievedChunk, error) {
	documents, _, err := e.documentRepo.ListByKnowledgeBase(ctx, req.KnowledgeName, 1, 1000)
	if err != nil {
		return nil, err
	}
	results := make([]domainmodel.RetrievedChunk, 0)
	for _, document := range documents {
		chunks, _, err := e.chunkRepo.ListByDocumentID(ctx, document.ID, 1, 1000)
		if err != nil {
			return nil, err
		}
		for _, chunk := range chunks {
			score := calculateRetrieveScore(req.Question, chunk.Content)
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
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	if req.TopK > 0 && len(results) > req.TopK {
		results = results[:req.TopK]
	}
	return results, nil
}

func calculateRetrieveScore(question string, content string) float64 {
	question = strings.TrimSpace(question)
	content = strings.TrimSpace(content)
	if question == "" || content == "" {
		return 0
	}
	if strings.Contains(content, question) {
		return 1
	}
	questionTerms := splitRetrieveTerms(question)
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

func splitRetrieveTerms(question string) []string {
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
