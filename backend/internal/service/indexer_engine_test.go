package service

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	domainservice "github.com/gaohao-creator/go-rag/internal/domain/service"
)

func TestDefaultIndexerEngine_IndexTextFile(t *testing.T) {
	engine, err := NewDefaultIndexerEngine()
	if err != nil {
		t.Fatalf("NewDefaultIndexerEngine returned error: %v", err)
	}

	chunks, err := engine.Index(context.Background(), domainservice.IndexRequest{
		URI:           filepath.Join("testdata", "index.txt"),
		KnowledgeName: "demo",
	})
	if err != nil {
		t.Fatalf("Index returned error: %v", err)
	}
	if len(chunks) == 0 {
		t.Fatal("expected at least one chunk")
	}
	if chunks[0].ChunkID == "" {
		t.Fatal("expected non-empty chunk id")
	}
	if !strings.Contains(chunks[0].Content, "测试索引器") {
		t.Fatalf("expected chunk content to contain 测试索引器, got %s", chunks[0].Content)
	}
}

func TestDefaultIndexerEngine_IndexMarkdownFile(t *testing.T) {
	engine, err := NewDefaultIndexerEngine()
	if err != nil {
		t.Fatalf("NewDefaultIndexerEngine returned error: %v", err)
	}

	chunks, err := engine.Index(context.Background(), domainservice.IndexRequest{
		URI:           filepath.Join("testdata", "index.md"),
		KnowledgeName: "demo",
	})
	if err != nil {
		t.Fatalf("Index returned error: %v", err)
	}
	if len(chunks) == 0 {
		t.Fatal("expected at least one chunk")
	}
	if chunks[0].ChunkID == "" {
		t.Fatal("expected non-empty chunk id")
	}
}
