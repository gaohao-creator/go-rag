package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
	"github.com/gaohao-creator/go-rag/internal/service"
	webhandler "github.com/gaohao-creator/go-rag/internal/web/handler"
	webrouter "github.com/gaohao-creator/go-rag/internal/web/router"
)

type fakeChatService struct{}

func (f *fakeChatService) Chat(_ context.Context, in service.ChatInput) (*service.ChatResult, error) {
	return &service.ChatResult{
		Answer: "模拟回答: " + in.Question,
		References: []domainmodel.RetrievedChunk{{ChunkID: "chunk-1", Content: "RAG 是检索增强生成。", Score: 0.9}},
	}, nil
}

func (f *fakeChatService) ChatStream(_ context.Context, in service.ChatInput) (*service.ChatStreamResult, error) {
	return &service.ChatStreamResult{
		Answer:     "模拟回答: " + in.Question,
		References: []domainmodel.RetrievedChunk{{ChunkID: "chunk-1", Content: "RAG 是检索增强生成。", Score: 0.9}},
		Chunks:     []string{"模拟回答: ", in.Question},
	}, nil
}

func TestChatHandler_Chat(t *testing.T) {
	h := webhandler.NewHandler(nil, nil, nil, nil, nil, &fakeChatService{})
	engine := webrouter.NewRouter(h)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/chat", bytes.NewBufferString(`{"conv_id":"conv-1","question":"什么是RAG","knowledge_name":"demo"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	engine.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", response.Code, response.Body.String())
	}
	var payload struct {
		Code int `json:"code"`
		Data struct {
			Answer     string                       `json:"answer"`
			References []domainmodel.RetrievedChunk `json:"references"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	if payload.Data.Answer == "" {
		t.Fatal("expected answer")
	}
	if len(payload.Data.References) != 1 {
		t.Fatalf("expected 1 reference, got %d", len(payload.Data.References))
	}
}
