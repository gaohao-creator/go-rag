package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type healthzResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    string `json:"data"`
}

func TestRouter_Healthz(t *testing.T) {
	engine := NewRouter()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	resp := httptest.NewRecorder()

	engine.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.Code)
	}

	var payload healthzResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if payload.Code != 0 {
		t.Fatalf("expected code 0, got %d", payload.Code)
	}
	if payload.Message != "ok" {
		t.Fatalf("expected message ok, got %s", payload.Message)
	}
	if payload.Data != "pong" {
		t.Fatalf("expected data pong, got %s", payload.Data)
	}
}
