package service

import (
	"context"
	"fmt"
	"strings"
	"testing"

	apprag "github.com/gaohao-creator/go-rag/internal/rag"
)

func TestChatService_ChatStreamPersistsConversationAndReturnsChunks(t *testing.T) {
	history := newFakeMessageRepository()
	retriever := &fakeChatRetriever{}
	model := &fakeChatModel{}
	svc := NewChatService(history, retriever, model, apprag.NewRAG(nil, nil, nil, nil))

	result, err := svc.ChatStream(context.Background(), ChatInput{
		ConvID:        "conv-stream-1",
		Question:      "什么是 RAG",
		KnowledgeName: "demo",
	})
	if err != nil {
		t.Fatalf("ChatStream returned error: %v", err)
	}
	if len(result.Chunks) == 0 {
		t.Fatal("expected stream chunks")
	}
	if !strings.Contains(result.Answer, "什么是 RAG") {
		t.Fatalf("expected answer to contain question, got %s", result.Answer)
	}
	if history.CountByConversation("conv-stream-1") != 2 {
		t.Fatalf("expected 2 persisted messages, got %d", history.CountByConversation("conv-stream-1"))
	}
	fmt.Printf("result: %+v", result)

}
