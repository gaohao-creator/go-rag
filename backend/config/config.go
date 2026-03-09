package config

import (
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTP     HTTPConfig     `yaml:"http"`
	Database DatabaseConfig `yaml:"database"`
	Chat     ChatConfig     `yaml:"chat"`
}

type HTTPConfig struct {
	Port string `yaml:"port"`
}

type DatabaseConfig struct {
	Driver                 string `yaml:"driver"`
	DSN                    string `yaml:"dsn"`
	MaxIdleConns           int    `yaml:"max_idle_conns"`
	MaxOpenConns           int    `yaml:"max_open_conns"`
	ConnMaxLifetimeSeconds int    `yaml:"conn_max_lifetime_seconds"`
}

type ChatConfig struct {
	Provider     string `yaml:"provider"`
	APIKey       string `yaml:"api_key"`
	BaseURL      string `yaml:"base_url"`
	Model        string `yaml:"model"`
	SystemPrompt string `yaml:"system_prompt"`
}

func (c *Config) ApplyDefaults() {
	if c.HTTP.Port == "" {
		c.HTTP.Port = "8080"
	}
	if c.Database.Driver == "" {
		c.Database.Driver = "mysql"
	}
	if c.Database.MaxIdleConns == 0 {
		c.Database.MaxIdleConns = 10
	}
	if c.Database.MaxOpenConns == 0 {
		c.Database.MaxOpenConns = 50
	}
	if c.Database.ConnMaxLifetimeSeconds == 0 {
		c.Database.ConnMaxLifetimeSeconds = int((30 * time.Minute).Seconds())
	}
	if c.Chat.Provider == "" {
		c.Chat.Provider = "openai"
	}
	if c.Chat.BaseURL == "" {
		c.Chat.BaseURL = "https://api.openai.com/v1"
	}
	if c.Chat.SystemPrompt == "" {
		c.Chat.SystemPrompt = "你是一个专业的 AI 助手，请基于提供的参考内容回答问题。"
	}
}

func (c *Config) ApplyEnvLookup(lookup func(string) (string, bool)) {
	applyString := func(key string, target *string) {
		if value, ok := lookup(key); ok {
			value = strings.TrimSpace(value)
			if value != "" {
				*target = value
			}
		}
	}
	applyInt := func(key string, target *int) {
		if value, ok := lookup(key); ok {
			value = strings.TrimSpace(value)
			if value == "" {
				return
			}
			if parsed, err := strconv.Atoi(value); err == nil {
				*target = parsed
			}
		}
	}

	applyString("GO_RAG_HTTP_PORT", &c.HTTP.Port)
	applyString("GO_RAG_DATABASE_DRIVER", &c.Database.Driver)
	applyString("GO_RAG_DATABASE_DSN", &c.Database.DSN)
	applyInt("GO_RAG_DATABASE_MAX_IDLE_CONNS", &c.Database.MaxIdleConns)
	applyInt("GO_RAG_DATABASE_MAX_OPEN_CONNS", &c.Database.MaxOpenConns)
	applyInt("GO_RAG_DATABASE_CONN_MAX_LIFETIME_SECONDS", &c.Database.ConnMaxLifetimeSeconds)
	applyString("GO_RAG_CHAT_PROVIDER", &c.Chat.Provider)
	applyString("GO_RAG_CHAT_API_KEY", &c.Chat.APIKey)
	applyString("GO_RAG_CHAT_BASE_URL", &c.Chat.BaseURL)
	applyString("GO_RAG_CHAT_MODEL", &c.Chat.Model)
	applyString("GO_RAG_CHAT_SYSTEM_PROMPT", &c.Chat.SystemPrompt)
}
