package router_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
	"github.com/gaohao-creator/go-rag/internal/service"
	"github.com/gaohao-creator/go-rag/ioc"
)

func writeRouterTestConfig(t *testing.T) string {
	t.Helper()
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	dsnName := strings.ReplaceAll(strings.ToLower(t.Name()), "/", "_")
	content := []byte(fmt.Sprintf("http:\n  port: \"18082\"\n\ndatabase:\n  driver: \"sqlite\"\n  dsn: \"file:%s?mode=memory&cache=shared\"\n  max_idle_conns: 1\n  max_open_conns: 1\n  conn_max_lifetime_seconds: 60\n\nchat:\n  provider: \"fake\"\n", dsnName))
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	return configPath
}

func performJSON(t *testing.T, app *ioc.App, method string, path string, body string) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	app.Router.ServeHTTP(response, request)
	return response
}

func createKnowledgeBaseForTest(t *testing.T, app *ioc.App) int64 {
	t.Helper()
	response := performJSON(t, app, http.MethodPost, "/api/v1/kb", `{"name":"demo","description":"desc","category":"general"}`)
	if response.Code != http.StatusOK {
		t.Fatalf("create kb expected 200, got %d, body=%s", response.Code, response.Body.String())
	}
	var payload struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode create kb response failed: %v", err)
	}
	return payload.Data.ID
}

func TestIntegration_IndexerRouteExists(t *testing.T) {
	configPath := writeRouterTestConfig(t)
	app, err := ioc.NewApp(configPath)
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/indexer", bytes.NewBufferString(`{"uri":"testdata/index.txt","knowledge_name":"demo","file_name":"index.txt"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	app.Router.ServeHTTP(response, request)
	if response.Code == http.StatusNotFound {
		t.Fatalf("expected route to exist, got 404")
	}
}

func TestIntegration_KnowledgeBaseLifecycle(t *testing.T) {
	configPath := writeRouterTestConfig(t)
	app, err := ioc.NewApp(configPath)
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}
	id := createKnowledgeBaseForTest(t, app)

	listResp := httptest.NewRecorder()
	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/kb", nil)
	app.Router.ServeHTTP(listResp, listReq)
	if listResp.Code != http.StatusOK {
		t.Fatalf("list kb expected 200, got %d, body=%s", listResp.Code, listResp.Body.String())
	}

	getResp := httptest.NewRecorder()
	getReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/kb/%d", id), nil)
	app.Router.ServeHTTP(getResp, getReq)
	if getResp.Code != http.StatusOK {
		t.Fatalf("get kb expected 200, got %d, body=%s", getResp.Code, getResp.Body.String())
	}

	updateResp := performJSON(t, app, http.MethodPut, fmt.Sprintf("/api/v1/kb/%d", id), `{"description":"desc-updated","status":2}`)
	if updateResp.Code != http.StatusOK {
		t.Fatalf("update kb expected 200, got %d, body=%s", updateResp.Code, updateResp.Body.String())
	}

	getUpdatedResp := httptest.NewRecorder()
	getUpdatedReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/kb/%d", id), nil)
	app.Router.ServeHTTP(getUpdatedResp, getUpdatedReq)
	var getPayload struct {
		Data domainmodel.KnowledgeBase `json:"data"`
	}
	if err := json.Unmarshal(getUpdatedResp.Body.Bytes(), &getPayload); err != nil {
		t.Fatalf("decode get kb failed: %v", err)
	}
	if getPayload.Data.Description != "desc-updated" || getPayload.Data.Status != 2 {
		t.Fatalf("unexpected kb after update: %+v", getPayload.Data)
	}

	deleteResp := httptest.NewRecorder()
	deleteReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/kb/%d", id), nil)
	app.Router.ServeHTTP(deleteResp, deleteReq)
	if deleteResp.Code != http.StatusOK {
		t.Fatalf("delete kb expected 200, got %d, body=%s", deleteResp.Code, deleteResp.Body.String())
	}
}

func TestIntegration_DocumentDeleteAndChunkMutations(t *testing.T) {
	configPath := writeRouterTestConfig(t)
	app, err := ioc.NewApp(configPath)
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}
	_, err = app.Services.KnowledgeBase.Create(context.Background(), service.CreateKnowledgeBaseInput{Name: "demo", Description: "desc", Category: "general"})
	if err != nil {
		t.Fatalf("create kb via service failed: %v", err)
	}
	docID, err := app.Services.Document.Create(context.Background(), service.CreateDocumentInput{KnowledgeBaseName: "demo", FileName: "readme.md", Status: 2})
	if err != nil {
		t.Fatalf("create document failed: %v", err)
	}
	if err = app.Services.Chunk.BatchCreate(context.Background(), []domainmodel.Chunk{
		{KnowledgeDocID: docID, ChunkID: "chunk-a", Content: "alpha", Ext: "{}", Status: 2},
		{KnowledgeDocID: docID, ChunkID: "chunk-b", Content: "beta", Ext: "{}", Status: 2},
	}); err != nil {
		t.Fatalf("batch create chunks failed: %v", err)
	}

	listChunksResp := httptest.NewRecorder()
	listChunksReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/chunks?knowledge_doc_id=%d", docID), nil)
	app.Router.ServeHTTP(listChunksResp, listChunksReq)
	if listChunksResp.Code != http.StatusOK {
		t.Fatalf("list chunks expected 200, got %d, body=%s", listChunksResp.Code, listChunksResp.Body.String())
	}
	var chunkListPayload struct {
		Data struct {
			Data []domainmodel.Chunk `json:"data"`
		} `json:"data"`
	}
	if err := json.Unmarshal(listChunksResp.Body.Bytes(), &chunkListPayload); err != nil {
		t.Fatalf("decode chunk list failed: %v", err)
	}
	if len(chunkListPayload.Data.Data) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(chunkListPayload.Data.Data))
	}
	firstChunkID := chunkListPayload.Data.Data[0].ID
	secondChunkID := chunkListPayload.Data.Data[1].ID

	updateStatusResp := performJSON(t, app, http.MethodPut, "/api/v1/chunks", fmt.Sprintf(`{"ids":[%d],"status":1}`, firstChunkID))
	if updateStatusResp.Code != http.StatusOK {
		t.Fatalf("update chunk status expected 200, got %d, body=%s", updateStatusResp.Code, updateStatusResp.Body.String())
	}
	updateContentResp := performJSON(t, app, http.MethodPut, "/api/v1/chunks_content", fmt.Sprintf(`{"id":%d,"content":"updated-alpha"}`, firstChunkID))
	if updateContentResp.Code != http.StatusOK {
		t.Fatalf("update chunk content expected 200, got %d, body=%s", updateContentResp.Code, updateContentResp.Body.String())
	}
	deleteChunkResp := httptest.NewRecorder()
	deleteChunkReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/chunks?id=%d", secondChunkID), nil)
	app.Router.ServeHTTP(deleteChunkResp, deleteChunkReq)
	if deleteChunkResp.Code != http.StatusOK {
		t.Fatalf("delete chunk expected 200, got %d, body=%s", deleteChunkResp.Code, deleteChunkResp.Body.String())
	}

	listChunksAfterResp := httptest.NewRecorder()
	listChunksAfterReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/chunks?knowledge_doc_id=%d", docID), nil)
	app.Router.ServeHTTP(listChunksAfterResp, listChunksAfterReq)
	var chunkAfterPayload struct {
		Data struct {
			Data []domainmodel.Chunk `json:"data"`
		} `json:"data"`
	}
	if err := json.Unmarshal(listChunksAfterResp.Body.Bytes(), &chunkAfterPayload); err != nil {
		t.Fatalf("decode chunk list after failed: %v", err)
	}
	if len(chunkAfterPayload.Data.Data) != 1 {
		t.Fatalf("expected 1 chunk after delete, got %d", len(chunkAfterPayload.Data.Data))
	}
	if chunkAfterPayload.Data.Data[0].Content != "updated-alpha" || chunkAfterPayload.Data.Data[0].Status != 1 {
		t.Fatalf("unexpected chunk after update: %+v", chunkAfterPayload.Data.Data[0])
	}

	deleteDocumentResp := httptest.NewRecorder()
	deleteDocumentReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/documents?document_id=%d", docID), nil)
	app.Router.ServeHTTP(deleteDocumentResp, deleteDocumentReq)
	if deleteDocumentResp.Code != http.StatusOK {
		t.Fatalf("delete document expected 200, got %d, body=%s", deleteDocumentResp.Code, deleteDocumentResp.Body.String())
	}

	listDocumentsResp := httptest.NewRecorder()
	listDocumentsReq := httptest.NewRequest(http.MethodGet, "/api/v1/documents?knowledge_name=demo", nil)
	app.Router.ServeHTTP(listDocumentsResp, listDocumentsReq)
	var documentsPayload struct {
		Data struct {
			Data []domainmodel.Document `json:"data"`
		} `json:"data"`
	}
	if err := json.Unmarshal(listDocumentsResp.Body.Bytes(), &documentsPayload); err != nil {
		t.Fatalf("decode document list failed: %v", err)
	}
	if len(documentsPayload.Data.Data) != 0 {
		t.Fatalf("expected 0 documents after delete, got %d", len(documentsPayload.Data.Data))
	}
}

func TestIntegration_ServerCompatiblePayloadsAndRoutes(t *testing.T) {
	configPath := writeRouterTestConfig(t)
	app, err := ioc.NewApp(configPath)
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}
	_, err = app.Services.KnowledgeBase.Create(context.Background(), service.CreateKnowledgeBaseInput{Name: "demo", Description: "desc", Category: "general"})
	if err != nil {
		t.Fatalf("create kb via service failed: %v", err)
	}
	docID, err := app.Services.Document.Create(context.Background(), service.CreateDocumentInput{KnowledgeBaseName: "demo", FileName: "readme.md", Status: 2})
	if err != nil {
		t.Fatalf("create document failed: %v", err)
	}
	if err = app.Services.Chunk.BatchCreate(context.Background(), []domainmodel.Chunk{
		{KnowledgeDocID: docID, ChunkID: "chunk-a", Content: "alpha", Ext: "{\"source\":\"demo\"}", Status: 2},
	}); err != nil {
		t.Fatalf("batch create chunk failed: %v", err)
	}

	kbResp := httptest.NewRecorder()
	kbReq := httptest.NewRequest(http.MethodGet, "/api/v1/kb", nil)
	app.Router.ServeHTTP(kbResp, kbReq)
	if !strings.Contains(kbResp.Body.String(), "createTime") {
		t.Fatalf("expected server-compatible kb payload, got %s", kbResp.Body.String())
	}

	docResp := httptest.NewRecorder()
	docReq := httptest.NewRequest(http.MethodGet, "/api/v1/documents?knowledge_name=demo", nil)
	app.Router.ServeHTTP(docResp, docReq)
	if !strings.Contains(docResp.Body.String(), "knowledgeBaseName") || !strings.Contains(docResp.Body.String(), "fileName") {
		t.Fatalf("expected server-compatible document payload, got %s", docResp.Body.String())
	}

	chunkResp := httptest.NewRecorder()
	chunkReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/chunks?knowledge_doc_id=%d", docID), nil)
	app.Router.ServeHTTP(chunkResp, chunkReq)
	if !strings.Contains(chunkResp.Body.String(), "knowledgeDocId") || !strings.Contains(chunkResp.Body.String(), "chunkId") {
		t.Fatalf("expected server-compatible chunk payload, got %s", chunkResp.Body.String())
	}
}

func TestIntegration_RetrieverDifyRouteWorksWithoutEnvelope(t *testing.T) {
	configPath := writeRouterTestConfig(t)
	app, err := ioc.NewApp(configPath)
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}
	_, err = app.Services.KnowledgeBase.Create(context.Background(), service.CreateKnowledgeBaseInput{Name: "demo", Description: "desc", Category: "general"})
	if err != nil {
		t.Fatalf("create kb failed: %v", err)
	}
	docID, err := app.Services.Document.Create(context.Background(), service.CreateDocumentInput{KnowledgeBaseName: "demo", FileName: "readme.md", Status: 2})
	if err != nil {
		t.Fatalf("create document failed: %v", err)
	}
	if err = app.Services.Chunk.BatchCreate(context.Background(), []domainmodel.Chunk{{KnowledgeDocID: docID, ChunkID: "chunk-1", Content: "RAG 是检索增强生成。", Ext: "{\"source\":\"demo\"}", Status: 2}}); err != nil {
		t.Fatalf("create chunk failed: %v", err)
	}

	response := performJSON(t, app, http.MethodPost, "/api/v1/dify/retrieval", `{"knowledge_id":"demo","query":"什么是RAG","retrieval_setting":{"top_k":5,"score_threshold":0.2}}`)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", response.Code, response.Body.String())
	}
	if strings.Contains(response.Body.String(), "\"code\"") {
		t.Fatalf("expected raw dify payload without envelope, got %s", response.Body.String())
	}
	var payload struct {
		Records []struct {
			Content string  `json:"content"`
			Score   float64 `json:"score"`
		} `json:"records"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode dify response failed: %v", err)
	}
	if len(payload.Records) != 1 || payload.Records[0].Content == "" {
		t.Fatalf("unexpected dify records: %+v", payload.Records)
	}
}



