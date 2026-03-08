package ioc

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func writeTestConfig(t *testing.T) string {
	t.Helper()
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	content := []byte("http:\n  port: \"18080\"\n\ndatabase:\n  driver: \"sqlite\"\n  dsn: \"file:task10_app?mode=memory&cache=shared\"\n  max_idle_conns: 1\n  max_open_conns: 1\n  conn_max_lifetime_seconds: 60\n")
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}
	return configPath
}

func TestNewApp_WiresPhase1Dependencies(t *testing.T) {
	configPath := writeTestConfig(t)
	app, err := NewApp(configPath)
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}
	if app.Router == nil {
		t.Fatal("expected router")
	}
	if app.Handler == nil {
		t.Fatal("expected handler")
	}
	if app.Services == nil || app.Services.KnowledgeBase == nil || app.Services.Indexer == nil || app.Services.Retriever == nil {
		t.Fatal("expected services to be wired")
	}
}

func TestNewApp_RouterServesPhase1Routes(t *testing.T) {
	configPath := writeTestConfig(t)
	app, err := NewApp(configPath)
	if err != nil {
		t.Fatalf("NewApp returned error: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/v1/kb", bytes.NewBufferString(`{"name":"demo","description":"desc","category":"general"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	app.Router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body=%s", response.Code, response.Body.String())
	}
}
