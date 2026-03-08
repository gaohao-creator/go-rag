package repository

import (
	"context"

	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
)

type ChunkRepository interface {
	BatchCreate(ctx context.Context, chunks []domainmodel.Chunk) error
	ListByDocumentID(ctx context.Context, documentID int64, page int, size int) ([]domainmodel.Chunk, int64, error)
	DeleteByID(ctx context.Context, id int64) error
	DeleteByDocumentID(ctx context.Context, documentID int64) error
	UpdateStatusByIDs(ctx context.Context, ids []int64, status int) error
	UpdateContentByID(ctx context.Context, id int64, content string) error
}
