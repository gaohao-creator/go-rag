package config

import "testing"

func TestApplyDefaults_SetsOpenAIChatDefaults(t *testing.T) {
	conf := &Config{}
	conf.ApplyDefaults()
	if conf.Chat.Provider != "openai" {
		t.Fatalf("expected provider openai, got %s", conf.Chat.Provider)
	}
	if conf.Chat.BaseURL != "https://api.openai.com/v1" {
		t.Fatalf("expected default base url, got %s", conf.Chat.BaseURL)
	}
}
