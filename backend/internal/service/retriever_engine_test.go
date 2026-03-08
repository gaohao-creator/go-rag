package service

import (
	"context"
	"testing"

	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
	domainservice "github.com/gaohao-creator/go-rag/internal/domain/service"
)

type fakeRetrieverDocumentRepository struct {
	documents []domainmodel.Document
}

func (f *fakeRetrieverDocumentRepository) Create(_ context.Context, _ domainmodel.Document) (int64, error) {
	return 0, nil
}
func (f *fakeRetrieverDocumentRepository) ListByKnowledgeBase(_ context.Context, knowledgeBaseName string, _ int, _ int) ([]domainmodel.Document, int64, error) {
	result := make([]domainmodel.Document, 0)
	for _, document := range f.documents {
		if document.KnowledgeBaseName == knowledgeBaseName {
			result = append(result, document)
		}
	}
	return result, int64(len(result)), nil
}
func (f *fakeRetrieverDocumentRepository) UpdateStatus(_ context.Context, _ int64, _ int) error {
	return nil
}
func (f *fakeRetrieverDocumentRepository) Delete(_ context.Context, _ int64) error { return nil }

type fakeRetrieverChunkRepository struct{ byDocumentID map[int64][]domainmodel.Chunk }

func (f *fakeRetrieverChunkRepository) BatchCreate(_ context.Context, _ []domainmodel.Chunk) error {
	return nil
}
func (f *fakeRetrieverChunkRepository) ListByDocumentID(_ context.Context, documentID int64, _ int, _ int) ([]domainmodel.Chunk, int64, error) {
	result := f.byDocumentID[documentID]
	return result, int64(len(result)), nil
}
func (f *fakeRetrieverChunkRepository) DeleteByID(_ context.Context, _ int64) error { return nil }
func (f *fakeRetrieverChunkRepository) DeleteByDocumentID(_ context.Context, _ int64) error {
	return nil
}
func (f *fakeRetrieverChunkRepository) UpdateStatusByIDs(_ context.Context, _ []int64, _ int) error {
	return nil
}
func (f *fakeRetrieverChunkRepository) UpdateContentByID(_ context.Context, _ int64, _ string) error {
	return nil
}

func TestDatabaseRetrieverEngine_RetrieveFiltersByKnowledgeBaseAndSorts(t *testing.T) {
	documentRepo := &fakeRetrieverDocumentRepository{documents: []domainmodel.Document{{ID: 1, KnowledgeBaseName: "demo", FileName: "a.md"}, {ID: 2, KnowledgeBaseName: "other", FileName: "b.md"}}}
	chunkRepo := &fakeRetrieverChunkRepository{byDocumentID: map[int64][]domainmodel.Chunk{1: {{KnowledgeDocID: 1, ChunkID: "chunk-1", Content: "这是测试索引器的内容"}, {KnowledgeDocID: 1, ChunkID: "chunk-2", Content: "RAG 检索流程说明"}}, 2: {{KnowledgeDocID: 2, ChunkID: "chunk-3", Content: "其它知识库的内容"}}}}

	engine := NewDatabaseRetrieverEngine(documentRepo, chunkRepo)
	chunks, err := engine.Retrieve(context.Background(), domainservice.RetrieveRequest{Question: "测试索引器", KnowledgeName: "demo", TopK: 5, Score: 0.1})
	if err != nil {
		t.Fatalf("Retrieve returned error: %v", err)
	}
	if len(chunks) == 0 {
		t.Fatal("expected retrieval results")
	}
	if chunks[0].ChunkID != "chunk-1" {
		t.Fatalf("expected chunk-1 first, got %s", chunks[0].ChunkID)
	}
	for _, chunk := range chunks {
		if chunk.ChunkID == "chunk-3" {
			t.Fatal("expected results filtered by knowledge base")
		}
	}
}
