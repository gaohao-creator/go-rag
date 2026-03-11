package service_test

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRAGContractsAreMergedIntoSingleFile(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("无法定位测试文件")
	}

	serviceDir := filepath.Dir(filename)
	if _, err := os.Stat(filepath.Join(serviceDir, "rag.go")); err != nil {
		t.Fatalf("期望存在 rag.go: %v", err)
	}

	for _, name := range []string{"indexer.go", "retriever.go", "quality.go", "vector.go"} {
		_, err := os.Stat(filepath.Join(serviceDir, name))
		if !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("期望 %s 已合并进 rag.go", name)
		}
	}
}

func TestRAGContractUsesChunkStoreNaming(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("无法定位测试文件")
	}

	content, err := os.ReadFile(filepath.Join(filepath.Dir(filename), "rag.go"))
	if err != nil {
		t.Fatalf("读取 rag.go 失败: %v", err)
	}

	text := string(content)
	for _, expected := range []string{
		"type ChunkStoreRequest struct",
		"type ChunkStore interface",
		"Store(ctx context.Context, req ChunkStoreRequest) error",
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("期望 rag.go 包含 %q", expected)
		}
	}

	for _, legacy := range []string{
		"type VectorWriteRequest struct",
		"type VectorWriter interface",
		"IndexChunks(ctx context.Context, req VectorWriteRequest) error",
	} {
		if strings.Contains(text, legacy) {
			t.Fatalf("期望 rag.go 不再包含 %q", legacy)
		}
	}
}
