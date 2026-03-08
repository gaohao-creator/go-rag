package dao

import (
	"context"
	"testing"

	daoentity "github.com/gaohao-creator/go-rag/internal/repository/dao/entity"
)

func TestChunkDAO_BatchCreateAndListByDocumentID(t *testing.T) {
	db := newTestDB(t)
	dao := NewChunkDAO(db)
	err := dao.BatchCreate(context.Background(), []*daoentity.Chunk{
		{
			KnowledgeDocID: 1,
			ChunkID:        "chunk-1",
			Content:        "hello",
			Ext:            "{}",
			Status:         1,
		},
		{
			KnowledgeDocID: 1,
			ChunkID:        "chunk-2",
			Content:        "world",
			Ext:            "{}",
			Status:         1,
		},
	})
	if err != nil {
		t.Fatalf("BatchCreate returned error: %v", err)
	}

	chunks, total, err := dao.ListByDocumentID(context.Background(), 1, 1, 10)
	if err != nil {
		t.Fatalf("ListByDocumentID returned error: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected total 2, got %d", total)
	}
	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}
}
