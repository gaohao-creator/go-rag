package repository

import (
	"context"

	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
)

type KnowledgeBaseRepository interface {
	Create(ctx context.Context, kb domainmodel.KnowledgeBase) (int64, error)
	GetByID(ctx context.Context, id int64) (*domainmodel.KnowledgeBase, error)
	List(ctx context.Context, filter domainmodel.KnowledgeBaseFilter) ([]domainmodel.KnowledgeBase, error)
	Update(ctx context.Context, id int64, patch domainmodel.KnowledgeBasePatch) error
	Delete(ctx context.Context, id int64) error
}
