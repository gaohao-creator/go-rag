package service

import (
	"context"
	"errors"
	"testing"

	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
	domainservice "github.com/gaohao-creator/go-rag/internal/domain/service"
)

type fakeIndexerEngine struct {
	chunks []domainservice.IndexedChunk
	err    error
}

func (f *fakeIndexerEngine) Index(_ context.Context, _ domainservice.IndexRequest) ([]domainservice.IndexedChunk, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.chunks, nil
}

type fakeIndexerDocumentRepository struct {
	nextID        int64
	created       []domainmodel.Document
	statusUpdates []documentStatusUpdate
}

type documentStatusUpdate struct {
	id     int64
	status int
}

func (f *fakeIndexerDocumentRepository) Create(_ context.Context, document domainmodel.Document) (int64, error) {
	f.created = append(f.created, document)
	if f.nextID == 0 {
		f.nextID = 1
	}
	return f.nextID, nil
}

func (f *fakeIndexerDocumentRepository) ListByKnowledgeBase(_ context.Context, _ string, _ int, _ int) ([]domainmodel.Document, int64, error) {
	return nil, 0, nil
}

func (f *fakeIndexerDocumentRepository) UpdateStatus(_ context.Context, id int64, status int) error {
	f.statusUpdates = append(f.statusUpdates, documentStatusUpdate{id: id, status: status})
	return nil
}

func (f *fakeIndexerDocumentRepository) Delete(_ context.Context, _ int64) error { return nil }

type fakeIndexerChunkRepository struct {
	created [][]domainmodel.Chunk
}

func (f *fakeIndexerChunkRepository) BatchCreate(_ context.Context, chunks []domainmodel.Chunk) error {
	f.created = append(f.created, chunks)
	return nil
}

func (f *fakeIndexerChunkRepository) ListByDocumentID(_ context.Context, _ int64, _ int, _ int) ([]domainmodel.Chunk, int64, error) {
	return nil, 0, nil
}

func (f *fakeIndexerChunkRepository) DeleteByID(_ context.Context, _ int64) error         { return nil }
func (f *fakeIndexerChunkRepository) DeleteByDocumentID(_ context.Context, _ int64) error { return nil }
func (f *fakeIndexerChunkRepository) UpdateStatusByIDs(_ context.Context, _ []int64, _ int) error {
	return nil
}
func (f *fakeIndexerChunkRepository) UpdateContentByID(_ context.Context, _ int64, _ string) error {
	return nil
}

func TestIndexerService_IndexSuccess(t *testing.T) {
	documentRepo := &fakeIndexerDocumentRepository{nextID: 42}
	chunkRepo := &fakeIndexerChunkRepository{}
	engine := &fakeIndexerEngine{chunks: []domainservice.IndexedChunk{{ChunkID: "chunk-1", Content: "hello", Ext: "{}"}, {ChunkID: "chunk-2", Content: "world", Ext: "{}"}}}

	svc := NewIndexerService(documentRepo, chunkRepo, engine)
	ids, err := svc.Index(context.Background(), IndexInput{URI: "./testdata/index.txt", KnowledgeName: "demo", FileName: "index.txt"})
	if err != nil {
		t.Fatalf("Index returned error: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("expected 2 ids, got %d", len(ids))
	}
	if len(documentRepo.created) != 1 {
		t.Fatalf("expected 1 created document, got %d", len(documentRepo.created))
	}
	if documentRepo.created[0].Status != DocumentStatusPending {
		t.Fatalf("expected pending status, got %d", documentRepo.created[0].Status)
	}
	if len(documentRepo.statusUpdates) != 1 || documentRepo.statusUpdates[0].status != DocumentStatusActive {
		t.Fatalf("expected active status update, got %+v", documentRepo.statusUpdates)
	}
	if len(chunkRepo.created) != 1 || len(chunkRepo.created[0]) != 2 {
		t.Fatalf("expected one chunk batch with 2 chunks, got %+v", chunkRepo.created)
	}
	if chunkRepo.created[0][0].KnowledgeDocID != 42 {
		t.Fatalf("expected knowledge doc id 42, got %d", chunkRepo.created[0][0].KnowledgeDocID)
	}
}

func TestIndexerService_IndexFailureMarksDocumentFailed(t *testing.T) {
	documentRepo := &fakeIndexerDocumentRepository{nextID: 7}
	chunkRepo := &fakeIndexerChunkRepository{}
	engine := &fakeIndexerEngine{err: errors.New("boom")}

	svc := NewIndexerService(documentRepo, chunkRepo, engine)
	_, err := svc.Index(context.Background(), IndexInput{URI: "./testdata/index.txt", KnowledgeName: "demo", FileName: "index.txt"})
	if err == nil {
		t.Fatal("expected error")
	}
	if len(documentRepo.statusUpdates) != 1 || documentRepo.statusUpdates[0].status != DocumentStatusFailed {
		t.Fatalf("expected failed status update, got %+v", documentRepo.statusUpdates)
	}
	if len(chunkRepo.created) != 0 {
		t.Fatalf("expected no chunks created, got %+v", chunkRepo.created)
	}
}
