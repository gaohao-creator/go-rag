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

type fakeKBService struct{}

func (f *fakeKBService) Create(_ context.Context, in service.CreateKnowledgeBaseInput) (int64, error) {
	if in.Name != "demo" { return 0, context.Canceled }
	return 12, nil
}
func (f *fakeKBService) List(_ context.Context, _ service.ListKnowledgeBasesInput) ([]domainmodel.KnowledgeBase, error) { return nil, nil }
func (f *fakeKBService) GetByID(_ context.Context, _ int64) (*domainmodel.KnowledgeBase, error) { return nil, nil }
func (f *fakeKBService) Update(_ context.Context, _ service.UpdateKnowledgeBaseInput) error { return nil }
func (f *fakeKBService) Delete(_ context.Context, _ int64) error { return nil }

func TestKnowledgeBaseHandler_Create(t *testing.T) {
	h := webhandler.NewHandler(&fakeKBService{}, nil, nil, nil, nil, nil)
	engine := webrouter.NewRouter(h)
	body := bytes.NewBufferString(`{"name":"demo","description":"desc","category":"general"}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/kb", body)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	engine.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	var payload struct {
		Code int `json:"code"`
		Data struct { ID int64 `json:"id"` } `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if payload.Code != 0 {
		t.Fatalf("expected code 0, got %d", payload.Code)
	}
	if payload.Data.ID != 12 {
		t.Fatalf("expected id 12, got %d", payload.Data.ID)
	}
}


