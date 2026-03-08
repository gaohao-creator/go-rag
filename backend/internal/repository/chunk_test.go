package repository

import (
	"context"
	"testing"

	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
	internaldao "github.com/gaohao-creator/go-rag/internal/repository/dao"
)

func newTestChunkRepository(t *testing.T) *ChunkRepository {
	t.Helper()
	db := newTestDB(t)
	return NewChunkRepository(internaldao.NewChunkDAO(db))
}

func TestChunkRepository_BatchCreate(t *testing.T) {
	repo := newTestChunkRepository(t)
	err := repo.BatchCreate(context.Background(), []domainmodel.Chunk{
		{
			KnowledgeDocID: 1,
			ChunkID:        "chunk-1",
			Content:        "hello",
			Ext:            "{}",
			Status:         1,
		},
	})
	if err != nil {
		t.Fatalf("BatchCreate returned error: %v", err)
	}
}
