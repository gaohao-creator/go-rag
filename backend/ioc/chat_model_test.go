package ioc

import (
	"testing"

	appconfig "github.com/gaohao-creator/go-rag/config"
	appservice "github.com/gaohao-creator/go-rag/internal/service"
)

func TestBuildChatModel_ExplicitFakeProviderUsesFakeModel(t *testing.T) {
	conf := &appconfig.Config{}
	conf.ApplyDefaults()
	conf.Chat.Provider = "fake"
	model, err := buildChatModel(conf)
	if err != nil {
		t.Fatalf("buildChatModel returned error: %v", err)
	}
	if _, ok := model.(*appservice.FakeChatModel); !ok {
		t.Fatalf("expected FakeChatModel, got %T", model)
	}
}

func TestBuildChatModel_OpenAIRequiresCredentials(t *testing.T) {
	conf := &appconfig.Config{}
	conf.ApplyDefaults()
	conf.Chat.APIKey = ""
	conf.Chat.Model = "gpt-4o-mini"
	_, err := buildChatModel(conf)
	if err == nil {
		t.Fatal("expected error")
	}
}
