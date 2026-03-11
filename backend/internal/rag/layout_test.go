package rag_test

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestRAGModuleUsesFacadeLayout(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("无法定位测试文件")
	}

	ragDir := filepath.Dir(filename)
	requiredPaths := []string{
		"rag.go",
		filepath.Join("index", "index.go"),
		filepath.Join("store", "store.go"),
		filepath.Join("retrieve", "retrieve.go"),
		filepath.Join("rerank", "rerank.go"),
		filepath.Join("grade", "grade.go"),
		filepath.Join("es", "es.go"),
	}
	for _, relativePath := range requiredPaths {
		fullPath := filepath.Join(ragDir, relativePath)
		if _, err := os.Stat(fullPath); err != nil {
			t.Fatalf("期望存在 %s: %v", relativePath, err)
		}
	}

	legacyPaths := []string{
		"context.go",
		"embedder.go",
		"es_client.go",
		"es_component.go",
		"es_mapping.go",
		"graph.go",
		"index_name.go",
		"indexer_es.go",
		"retriever_es.go",
		"service.go",
		"indexer",
		"retriever",
		"grader",
	}
	for _, relativePath := range legacyPaths {
		fullPath := filepath.Join(ragDir, relativePath)
		_, err := os.Stat(fullPath)
		if !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("期望旧布局已移除: %s", relativePath)
		}
	}
}

func TestRAGModuleFacadeObjectsAreClearlyNamed(t *testing.T) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("无法定位测试文件")
	}

	ragDir := filepath.Dir(filename)
	expected := map[string]string{
		"rag.go":                         "type RAG struct",
		filepath.Join("index", "index.go"):       "type Index struct",
		filepath.Join("store", "store.go"):       "type Store struct",
		filepath.Join("retrieve", "retrieve.go"): "type Retrieve struct",
		filepath.Join("rerank", "rerank.go"):     "type Rerank struct",
		filepath.Join("grade", "grade.go"):       "type Grade struct",
		filepath.Join("es", "es.go"):             "type ES struct",
	}

	for relativePath, marker := range expected {
		content, err := os.ReadFile(filepath.Join(ragDir, relativePath))
		if err != nil {
			t.Fatalf("读取 %s 失败: %v", relativePath, err)
		}
		if !strings.Contains(string(content), marker) {
			t.Fatalf("期望 %s 包含 %q", relativePath, marker)
		}
	}
}
