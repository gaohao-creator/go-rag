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
	Vector   VectorConfig   `yaml:"vector"`
	Rerank   RerankConfig   `yaml:"rerank"`
	Quality  QualityConfig  `yaml:"quality"`
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

type VectorConfig struct {
	Enabled              bool   `yaml:"enabled"`
	Backend              string `yaml:"backend"`
	Address              string `yaml:"address"`
	Username             string `yaml:"username"`
	Password             string `yaml:"password"`
	IndexPrefix          string `yaml:"index_prefix"`
	ContentField         string `yaml:"content_field"`
	ContentVectorField   string `yaml:"content_vector_field"`
	QAContentField       string `yaml:"qa_content_field"`
	QAContentVectorField string `yaml:"qa_content_vector_field"`
	KnowledgeField       string `yaml:"knowledge_field"`
	ExtField             string `yaml:"ext_field"`
	Dimensions           int    `yaml:"dimensions"`
	EmbeddingAPIKey      string `yaml:"embedding_api_key"`
	EmbeddingBaseURL     string `yaml:"embedding_base_url"`
	EmbeddingModel       string `yaml:"embedding_model"`
}

type RerankConfig struct {
	Enabled  bool    `yaml:"enabled"`
	BaseURL  string  `yaml:"base_url"`
	APIKey   string  `yaml:"api_key"`
	Model    string  `yaml:"model"`
	TopN     int     `yaml:"top_n"`
	MinScore float64 `yaml:"min_score"`
}

type QualityConfig struct {
	QA     QAConfig     `yaml:"qa"`
	Grader GraderConfig `yaml:"grader"`
}

type QAConfig struct {
	Enabled       bool `yaml:"enabled"`
	QuestionCount int  `yaml:"question_count"`
}

type GraderConfig struct {
	Enabled bool `yaml:"enabled"`
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
	if c.Vector.Backend == "" {
		c.Vector.Backend = "es"
	}
	if c.Vector.IndexPrefix == "" {
		c.Vector.IndexPrefix = "rag-"
	}
	if c.Vector.ContentField == "" {
		c.Vector.ContentField = "content"
	}
	if c.Vector.ContentVectorField == "" {
		c.Vector.ContentVectorField = "content_vector"
	}
	if c.Vector.QAContentField == "" {
		c.Vector.QAContentField = "qa_content"
	}
	if c.Vector.QAContentVectorField == "" {
		c.Vector.QAContentVectorField = "qa_content_vector"
	}
	if c.Vector.KnowledgeField == "" {
		c.Vector.KnowledgeField = "_knowledge_name"
	}
	if c.Vector.ExtField == "" {
		c.Vector.ExtField = "ext"
	}
	if c.Vector.Dimensions == 0 {
		c.Vector.Dimensions = 1024
	}
	if c.Vector.EmbeddingModel == "" {
		c.Vector.EmbeddingModel = "text-embedding-3-large"
	}
	if c.Vector.EmbeddingAPIKey == "" {
		c.Vector.EmbeddingAPIKey = c.Chat.APIKey
	}
	if c.Vector.EmbeddingBaseURL == "" {
		c.Vector.EmbeddingBaseURL = c.Chat.BaseURL
	}
	if c.Rerank.TopN == 0 {
		c.Rerank.TopN = 5
	}
	if c.Quality.QA.QuestionCount == 0 {
		c.Quality.QA.QuestionCount = 3
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
	applyFloat := func(key string, target *float64) {
		if value, ok := lookup(key); ok {
			value = strings.TrimSpace(value)
			if value == "" {
				return
			}
			if parsed, err := strconv.ParseFloat(value, 64); err == nil {
				*target = parsed
			}
		}
	}
	applyBool := func(key string, target *bool) {
		if value, ok := lookup(key); ok {
			value = strings.TrimSpace(value)
			if value == "" {
				return
			}
			if parsed, err := strconv.ParseBool(value); err == nil {
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
	applyBool("GO_RAG_VECTOR_ENABLED", &c.Vector.Enabled)
	applyString("GO_RAG_VECTOR_BACKEND", &c.Vector.Backend)
	applyString("GO_RAG_VECTOR_ADDRESS", &c.Vector.Address)
	applyString("GO_RAG_VECTOR_USERNAME", &c.Vector.Username)
	applyString("GO_RAG_VECTOR_PASSWORD", &c.Vector.Password)
	applyString("GO_RAG_VECTOR_INDEX_PREFIX", &c.Vector.IndexPrefix)
	applyString("GO_RAG_VECTOR_CONTENT_FIELD", &c.Vector.ContentField)
	applyString("GO_RAG_VECTOR_CONTENT_VECTOR_FIELD", &c.Vector.ContentVectorField)
	applyString("GO_RAG_VECTOR_QA_CONTENT_FIELD", &c.Vector.QAContentField)
	applyString("GO_RAG_VECTOR_QA_CONTENT_VECTOR_FIELD", &c.Vector.QAContentVectorField)
	applyString("GO_RAG_VECTOR_KNOWLEDGE_FIELD", &c.Vector.KnowledgeField)
	applyString("GO_RAG_VECTOR_EXT_FIELD", &c.Vector.ExtField)
	applyInt("GO_RAG_VECTOR_DIMENSIONS", &c.Vector.Dimensions)
	applyString("GO_RAG_VECTOR_EMBEDDING_API_KEY", &c.Vector.EmbeddingAPIKey)
	applyString("GO_RAG_VECTOR_EMBEDDING_BASE_URL", &c.Vector.EmbeddingBaseURL)
	applyString("GO_RAG_VECTOR_EMBEDDING_MODEL", &c.Vector.EmbeddingModel)
	applyBool("GO_RAG_RERANK_ENABLED", &c.Rerank.Enabled)
	applyString("GO_RAG_RERANK_BASE_URL", &c.Rerank.BaseURL)
	applyString("GO_RAG_RERANK_API_KEY", &c.Rerank.APIKey)
	applyString("GO_RAG_RERANK_MODEL", &c.Rerank.Model)
	applyInt("GO_RAG_RERANK_TOP_N", &c.Rerank.TopN)
	applyFloat("GO_RAG_RERANK_MIN_SCORE", &c.Rerank.MinScore)
	applyBool("GO_RAG_QUALITY_QA_ENABLED", &c.Quality.QA.Enabled)
	applyInt("GO_RAG_QUALITY_QA_QUESTION_COUNT", &c.Quality.QA.QuestionCount)
	applyBool("GO_RAG_QUALITY_GRADER_ENABLED", &c.Quality.Grader.Enabled)
}
