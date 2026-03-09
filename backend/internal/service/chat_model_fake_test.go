package service

import (
	"context"
	"fmt"
	"strings"
	"testing"

	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
	domainservice "github.com/gaohao-creator/go-rag/internal/domain/service"
)

func TestFakeChatModel_GenerateAnswerIncludesQuestionAndReferences(t *testing.T) {
	model := NewFakeChatModel()
	answer, err := model.Generate(context.Background(), domainservice.ChatGenerateInput{
		Question: "什么是 RAG",
		References: []domainmodel.RetrievedChunk{
			{ChunkID: "chunk-1", Content: "RAG 是检索增强生成。"},
		},
	})
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if !strings.Contains(answer, "什么是 RAG") {
		t.Fatalf("expected answer to contain question, got %s", answer)
	}
	if !strings.Contains(answer, "RAG 是检索增强生成") {
		t.Fatalf("expected answer to contain reference summary, got %s", answer)
	}
	fmt.Println(answer)
}
