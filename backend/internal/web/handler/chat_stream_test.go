package handler_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
	"github.com/gaohao-creator/go-rag/internal/service"
	webhandler "github.com/gaohao-creator/go-rag/internal/web/handler"
	webrouter "github.com/gaohao-creator/go-rag/internal/web/router"
)

type fakeChatStreamService struct{}

func (f *fakeChatStreamService) Chat(_ context.Context, in service.ChatInput) (*service.ChatResult, error) {
	return &service.ChatResult{Answer: "模拟回答: " + in.Question}, nil
}

func (f *fakeChatStreamService) ChatStream(_ context.Context, in service.ChatInput) (*service.ChatStreamResult, error) {
	return &service.ChatStreamResult{
		Answer: "模拟回答: " + in.Question,
		References: []domainmodel.RetrievedChunk{
			{ChunkID: "chunk-1", Content: "RAG 是检索增强生成。", Ext: "{\"source\":\"demo\"}", Score: 0.9},
		},
		Chunks: []string{"模拟回答", ": ", in.Question},
	}, nil
}

func TestChatHandler_ChatStream(t *testing.T) {
	h := webhandler.NewHandler(nil, nil, nil, nil, nil, &fakeChatStreamService{})
	engine := webrouter.NewRouter(h)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/chat/stream", bytes.NewBufferString(`{"conv_id":"conv-1","question":"什么是RAG","knowledge_name":"demo"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	engine.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Header().Get("Content-Type"), "text/event-stream") {
		t.Fatalf("expected event-stream content type, got %s", response.Header().Get("Content-Type"))
	}
	body := response.Body.String()
	if !strings.Contains(body, "documents:") {
		t.Fatalf("expected documents prefix, got %s", body)
	}
	if !strings.Contains(body, "data:") {
		t.Fatalf("expected data prefix, got %s", body)
	}
	if !strings.Contains(body, "[DONE]") {
		t.Fatalf("expected done marker, got %s", body)
	}
}
