package ioc

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewConfig_LoadsDefaultConfig(t *testing.T) {
	configPath := filepath.Join("..", "config", "config.yaml")
	conf, err := NewConfig(configPath)
	if err != nil {
		t.Fatalf("NewConfig returned error: %v", err)
	}
	if conf.HTTP.Port == "" {
		t.Fatal("expected HTTP port")
	}
	if conf.Database.Driver == "" {
		t.Fatal("expected database driver")
	}
}

func TestNewConfig_LoadsDotEnvOverrides(t *testing.T) {
	rootDir := t.TempDir()
	configDir := filepath.Join(rootDir, "config")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	configPath := filepath.Join(configDir, "config.yaml")
	configContent := []byte("http:\n  port: \"8080\"\n\ndatabase:\n  driver: \"sqlite\"\n  dsn: \"file:test?mode=memory&cache=shared\"\n\nchat:\n  provider: \"fake\"\n  model: \"fake-model\"\n")
	if err := os.WriteFile(configPath, configContent, 0o644); err != nil {
		t.Fatalf("WriteFile config returned error: %v", err)
	}
	dotenvPath := filepath.Join(rootDir, ".env")
	dotenvContent := []byte("GO_RAG_CHAT_PROVIDER=openai\nGO_RAG_CHAT_API_KEY=test-key\nGO_RAG_CHAT_BASE_URL=https://apis.iflow.cn/v1\nGO_RAG_CHAT_MODEL=qwen3-max\n")
	if err := os.WriteFile(dotenvPath, dotenvContent, 0o644); err != nil {
		t.Fatalf("WriteFile dotenv returned error: %v", err)
	}

	conf, err := NewConfig(configPath)
	if err != nil {
		t.Fatalf("NewConfig returned error: %v", err)
	}
	if conf.Chat.Provider != "openai" {
		t.Fatalf("expected provider openai, got %s", conf.Chat.Provider)
	}
	if conf.Chat.APIKey != "test-key" {
		t.Fatalf("expected api key from .env, got %s", conf.Chat.APIKey)
	}
	if conf.Chat.BaseURL != "https://apis.iflow.cn/v1" {
		t.Fatalf("expected base url from .env, got %s", conf.Chat.BaseURL)
	}
	if conf.Chat.Model != "qwen3-max" {
		t.Fatalf("expected model from .env, got %s", conf.Chat.Model)
	}
}
