package service

import (
	"context"
	"testing"

	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
	domainservice "github.com/gaohao-creator/go-rag/internal/domain/service"
	apprag "github.com/gaohao-creator/go-rag/internal/rag"
)

type fakeRetrieverEngine struct {
	request domainservice.RetrieveRequest
	results []domainmodel.RetrievedChunk
	err     error
}

func (f *fakeRetrieverEngine) Retrieve(_ context.Context, req domainservice.RetrieveRequest) ([]domainmodel.RetrievedChunk, error) {
	f.request = req
	if f.err != nil {
		return nil, f.err
	}
	return f.results, nil
}

func TestRetrieverService_RetrieveAppliesDefaultsAndSorts(t *testing.T) {
	engine := &fakeRetrieverEngine{results: []domainmodel.RetrievedChunk{{ChunkID: "chunk-1", Content: "low", Score: 0.4}, {ChunkID: "chunk-2", Content: "high", Score: 0.9}}}

	svc := NewRetrieverService(apprag.NewRAG(nil, nil, engine, nil))
	chunks, err := svc.Retrieve(context.Background(), RetrieveInput{Question: "什么是 RAG", KnowledgeName: "demo"})
	if err != nil {
		t.Fatalf("Retrieve returned error: %v", err)
	}
	if engine.request.TopK != 5 {
		t.Fatalf("expected default topK 5, got %d", engine.request.TopK)
	}
	if engine.request.Score != 0.2 {
		t.Fatalf("expected default score 0.2, got %v", engine.request.Score)
	}
	if len(chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunks))
	}
	if chunks[0].ChunkID != "chunk-2" {
		t.Fatalf("expected first chunk chunk-2, got %s", chunks[0].ChunkID)
	}
}
