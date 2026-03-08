package service

import (
	"context"

	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
)

type RetrieveRequest struct {
	Question      string
	TopK          int
	Score         float64
	KnowledgeName string
}

type Retriever interface {
	Retrieve(ctx context.Context, req RetrieveRequest) ([]domainmodel.RetrievedChunk, error)
}
