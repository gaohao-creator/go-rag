package router_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cloudwego/eino/schema"
	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
	"github.com/gaohao-creator/go-rag/internal/service"
	"github.com/gaohao-creator/go-rag/ioc"
)

func TestIntegration_ChatRouteWorks(t *testing.T) {
	configPath := writeRouterTestConfig(t)
	app, err := ioc.NewApp(configPath)
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}
	_, err = app.Services.KnowledgeBase.Create(context.Background(), service.CreateKnowledgeBaseInput{Name: "demo", Description: "desc", Category: "general"})
	if err != nil {
		t.Fatalf("create kb failed: %v", err)
	}
	docID, err := app.Services.Document.Create(context.Background(), service.CreateDocumentInput{KnowledgeBaseName: "demo", FileName: "readme.md", Status: service.DocumentStatusActive})
	if err != nil {
		t.Fatalf("create document failed: %v", err)
	}
	if err = app.Services.Chunk.BatchCreate(context.Background(), []domainmodel.Chunk{{KnowledgeDocID: docID, ChunkID: "chunk-1", Content: "RAG 是检索增强生成。", Ext: "{}", Status: service.DocumentStatusActive}}); err != nil {
		t.Fatalf("create chunk failed: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/v1/chat", bytes.NewBufferString(`{"conv_id":"conv-1","question":"什么是RAG","knowledge_name":"demo"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	app.Router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", response.Code, response.Body.String())
	}
	var payload struct {
		Data struct {
			Answer     string             `json:"answer"`
			References []*schema.Document `json:"references"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response failed: %v", err)
	}
	if payload.Data.Answer == "" {
		t.Fatal("expected answer")
	}
	if len(payload.Data.References) == 0 {
		t.Fatal("expected references")
	}
	if payload.Data.References[0].ID != "chunk-1" {
		t.Fatalf("expected chunk-1, got %s", payload.Data.References[0].ID)
	}
}

func TestIntegration_ChatStreamRouteWorks(t *testing.T) {
	configPath := writeRouterTestConfig(t)
	app, err := ioc.NewApp(configPath)
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}
	_, err = app.Services.KnowledgeBase.Create(context.Background(), service.CreateKnowledgeBaseInput{Name: "demo", Description: "desc", Category: "general"})
	if err != nil {
		t.Fatalf("create kb failed: %v", err)
	}
	docID, err := app.Services.Document.Create(context.Background(), service.CreateDocumentInput{KnowledgeBaseName: "demo", FileName: "readme.md", Status: service.DocumentStatusActive})
	if err != nil {
		t.Fatalf("create document failed: %v", err)
	}
	if err = app.Services.Chunk.BatchCreate(context.Background(), []domainmodel.Chunk{{KnowledgeDocID: docID, ChunkID: "chunk-1", Content: "RAG 是检索增强生成。", Ext: "{}", Status: service.DocumentStatusActive}}); err != nil {
		t.Fatalf("create chunk failed: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/v1/chat/stream", bytes.NewBufferString(`{"conv_id":"conv-stream-1","question":"什么是RAG","knowledge_name":"demo"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	app.Router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Header().Get("Content-Type"), "text/event-stream") {
		t.Fatalf("expected event-stream content type, got %s", response.Header().Get("Content-Type"))
	}
	body := response.Body.String()
	if !strings.Contains(body, "documents:") || !strings.Contains(body, "data:") || !strings.Contains(body, "[DONE]") {
		t.Fatalf("unexpected stream body: %s", body)
	}
}
