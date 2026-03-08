package repository

import (
	"context"

	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
)

type DocumentRepository interface {
	Create(ctx context.Context, document domainmodel.Document) (int64, error)
	ListByKnowledgeBase(ctx context.Context, knowledgeBaseName string, page int, size int) ([]domainmodel.Document, int64, error)
	UpdateStatus(ctx context.Context, id int64, status int) error
	Delete(ctx context.Context, id int64) error
}
