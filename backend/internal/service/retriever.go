package service

import (
	"context"
	"errors"
	"sort"
	"strings"

	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
	domainservice "github.com/gaohao-creator/go-rag/internal/domain/service"
)

type RetrieveInput struct {
	Question      string
	TopK          int
	Score         float64
	KnowledgeName string
}

type RetrieverService struct {
	engine domainservice.Retriever
}

func NewRetrieverService(engine domainservice.Retriever) *RetrieverService {
	return &RetrieverService{engine: engine}
}

func (s *RetrieverService) Retrieve(ctx context.Context, in RetrieveInput) ([]domainmodel.RetrievedChunk, error) {
	if strings.TrimSpace(in.Question) == "" {
		return nil, errors.New("问题不能为空")
	}
	if strings.TrimSpace(in.KnowledgeName) == "" {
		return nil, errors.New("知识库名称不能为空")
	}
	if in.TopK == 0 {
		in.TopK = 5
	}
	if in.Score == 0 {
		in.Score = 0.2
	}
	chunks, err := s.engine.Retrieve(ctx, domainservice.RetrieveRequest{
		Question:      in.Question,
		TopK:          in.TopK,
		Score:         in.Score,
		KnowledgeName: in.KnowledgeName,
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(chunks, func(i, j int) bool {
		return chunks[i].Score > chunks[j].Score
	})
	if len(chunks) > in.TopK {
		chunks = chunks[:in.TopK]
	}
	return chunks, nil
}
