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

type fakeRetrieverService struct{}

func (f *fakeRetrieverService) Retrieve(_ context.Context, in service.RetrieveInput) ([]domainmodel.RetrievedChunk, error) {
	if in.KnowledgeName != "demo" {
		return nil, context.Canceled
	}
	return []domainmodel.RetrievedChunk{{ChunkID: "chunk-1", Content: "hello", Score: 0.9}}, nil
}

func TestRetrieverHandler_Retrieve(t *testing.T) {
	h := webhandler.NewHandler(nil, nil, nil, nil, &fakeRetrieverService{}, nil)
	engine := webrouter.NewRouter(h)
	body := bytes.NewBufferString(`{"question":"什么是RAG","knowledge_name":"demo"}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/retriever", body)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	engine.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	var payload struct {
		Code int `json:"code"`
		Data struct {
			Document []domainmodel.RetrievedChunk `json:"document"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(payload.Data.Document) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(payload.Data.Document))
	}
	if payload.Data.Document[0].ChunkID != "chunk-1" {
		t.Fatalf("expected chunk-1, got %s", payload.Data.Document[0].ChunkID)
	}
}


