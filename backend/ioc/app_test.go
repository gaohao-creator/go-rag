package ioc

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	appconfig "github.com/gaohao-creator/go-rag/config"
	domainservice "github.com/gaohao-creator/go-rag/internal/domain/service"
	appservice "github.com/gaohao-creator/go-rag/internal/service"
)

func writeTestConfig(t *testing.T) string {
	t.Helper()
	t.Setenv("GO_RAG_CHAT_PROVIDER", "fake")
	t.Setenv("GO_RAG_CHAT_API_KEY", "")
	t.Setenv("GO_RAG_CHAT_BASE_URL", "")
	t.Setenv("GO_RAG_CHAT_MODEL", "")
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	content := []byte("http:\n  port: \"18080\"\n\ndatabase:\n  driver: \"sqlite\"\n  dsn: \"file:task10_app?mode=memory&cache=shared\"\n  max_idle_conns: 1\n  max_open_conns: 1\n  conn_max_lifetime_seconds: 60\n\nchat:\n  provider: \"fake\"\n")
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
	if app.Config == nil {
		t.Fatal("expected app config")
	}
	if app.Config.Vector.Enabled {
		t.Fatal("expected vector disabled in default test config")
	}
	if app.Config.Rerank.Enabled {
		t.Fatal("expected rerank disabled in default test config")
	}
	if app.Config.Quality.QA.Enabled {
		t.Fatal("expected qa disabled in default test config")
	}
	if app.Config.Quality.Grader.Enabled {
		t.Fatal("expected grader disabled in default test config")
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

type fakeOnlyChatModel struct{}

func (f *fakeOnlyChatModel) Generate(_ context.Context, _ domainservice.ChatGenerateInput) (string, error) {
	return "ok", nil
}

func TestBuildQualityComponents_ReusesPromptCapableChatModel(t *testing.T) {
	promptModel, reranker, grader, err := buildQualityComponents(&appconfig.Config{
		Rerank: appconfig.RerankConfig{
			Enabled: true,
			BaseURL: "https://rerank.example.com/v1",
			TopN:    6,
		},
		Quality: appconfig.QualityConfig{
			QA: appconfig.QAConfig{
				Enabled:       true,
				QuestionCount: 4,
			},
			Grader: appconfig.GraderConfig{
				Enabled: true,
			},
		},
	}, appservice.NewFakeChatModel())
	if err != nil {
		t.Fatalf("buildQualityComponents returned error: %v", err)
	}
	if promptModel == nil {
		t.Fatal("expected prompt model")
	}
	if reranker == nil {
		t.Fatal("expected reranker")
	}
	if grader == nil {
		t.Fatal("expected grader")
	}
}

func TestBuildQualityComponents_RejectsChatModelWithoutPromptCapability(t *testing.T) {
	_, _, _, err := buildQualityComponents(&appconfig.Config{
		Quality: appconfig.QualityConfig{
			QA: appconfig.QAConfig{Enabled: true},
		},
	}, &fakeOnlyChatModel{})
	if err == nil {
		t.Fatal("expected error")
	}
}
