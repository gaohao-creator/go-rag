package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gaohao-creator/go-rag/internal/service"
	webhandler "github.com/gaohao-creator/go-rag/internal/web/handler"
	webrouter "github.com/gaohao-creator/go-rag/internal/web/router"
)

type fakeIndexerService struct{}

func (f *fakeIndexerService) Index(_ context.Context, in service.IndexInput) ([]string, error) {
	if in.KnowledgeName != "demo" {
		return nil, context.Canceled
	}
	return []string{"chunk-1", "chunk-2"}, nil
}

func TestIndexerHandler_Index(t *testing.T) {
	h := webhandler.NewHandler(nil, nil, nil, &fakeIndexerService{}, nil)
	engine := webrouter.NewRouter(h)
	body := bytes.NewBufferString(`{"uri":"testdata/index.txt","knowledge_name":"demo","file_name":"index.txt"}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/indexer", body)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	engine.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	var payload struct {
		Code int `json:"code"`
		Data struct {
			DocIDs []string `json:"doc_ids"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(payload.Data.DocIDs) != 2 {
		t.Fatalf("expected 2 ids, got %d", len(payload.Data.DocIDs))
	}
}
