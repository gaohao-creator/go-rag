package service

import (
	"context"
	"errors"
	"testing"

	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
)

type fakeChunkRepository struct{}

func (f *fakeChunkRepository) BatchCreate(_ context.Context, chunks []domainmodel.Chunk) error {
	if len(chunks) == 0 {
		return errors.New("unexpected chunk")
	}
	return nil
}

func (f *fakeChunkRepository) ListByDocumentID(_ context.Context, documentID int64, _ int, _ int) ([]domainmodel.Chunk, int64, error) {
	return []domainmodel.Chunk{{KnowledgeDocID: documentID, ChunkID: "chunk-1", Content: "hello", Status: 1}}, 1, nil
}

func (f *fakeChunkRepository) DeleteByID(_ context.Context, _ int64) error         { return nil }
func (f *fakeChunkRepository) DeleteByDocumentID(_ context.Context, _ int64) error { return nil }
func (f *fakeChunkRepository) UpdateStatusByIDs(_ context.Context, _ []int64, _ int) error {
	return nil
}
func (f *fakeChunkRepository) UpdateContentByID(_ context.Context, _ int64, _ string) error {
	return nil
}

func TestChunkService_BatchCreate(t *testing.T) {
	svc := NewChunkService(&fakeChunkRepository{})
	err := svc.BatchCreate(context.Background(), []domainmodel.Chunk{{
		KnowledgeDocID: 1,
		ChunkID:        "chunk-1",
		Content:        "hello",
		Status:         1,
	}})
	if err != nil {
		t.Fatalf("BatchCreate returned error: %v", err)
	}
}
