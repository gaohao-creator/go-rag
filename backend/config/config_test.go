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

func TestApplyDefaults_SetsVectorDefaults(t *testing.T) {
	conf := &Config{}
	conf.ApplyDefaults()
	if conf.Vector.Enabled {
		t.Fatal("expected vector disabled by default")
	}
	if conf.Vector.Backend != "es" {
		t.Fatalf("expected vector backend es, got %s", conf.Vector.Backend)
	}
	if conf.Vector.IndexPrefix != "rag-" {
		t.Fatalf("expected default index prefix rag-, got %s", conf.Vector.IndexPrefix)
	}
	if conf.Vector.ContentField != "content" {
		t.Fatalf("expected default content field, got %s", conf.Vector.ContentField)
	}
	if conf.Vector.ContentVectorField != "content_vector" {
		t.Fatalf("expected default content vector field, got %s", conf.Vector.ContentVectorField)
	}
	if conf.Vector.QAContentField != "qa_content" {
		t.Fatalf("expected default qa content field, got %s", conf.Vector.QAContentField)
	}
	if conf.Vector.QAContentVectorField != "qa_content_vector" {
		t.Fatalf("expected default qa content vector field, got %s", conf.Vector.QAContentVectorField)
	}
	if conf.Vector.KnowledgeField != "_knowledge_name" {
		t.Fatalf("expected default knowledge field, got %s", conf.Vector.KnowledgeField)
	}
	if conf.Vector.ExtField != "ext" {
		t.Fatalf("expected default ext field, got %s", conf.Vector.ExtField)
	}
	if conf.Vector.Dimensions != 1024 {
		t.Fatalf("expected default vector dimensions 1024, got %d", conf.Vector.Dimensions)
	}
	if conf.Vector.EmbeddingModel != "text-embedding-3-large" {
		t.Fatalf("expected default embedding model, got %s", conf.Vector.EmbeddingModel)
	}
	if conf.Rerank.Enabled {
		t.Fatal("expected rerank disabled by default")
	}
	if conf.Rerank.TopN != 5 {
		t.Fatalf("expected default rerank top n 5, got %d", conf.Rerank.TopN)
	}
	if conf.Rerank.MinScore != 0 {
		t.Fatalf("expected default rerank min score 0, got %v", conf.Rerank.MinScore)
	}
	if conf.Quality.QA.Enabled {
		t.Fatal("expected qa disabled by default")
	}
	if conf.Quality.QA.QuestionCount != 3 {
		t.Fatalf("expected default qa question count 3, got %d", conf.Quality.QA.QuestionCount)
	}
	if conf.Quality.Grader.Enabled {
		t.Fatal("expected grader disabled by default")
	}
}

func TestApplyEnvLookup_OverridesVectorConfig(t *testing.T) {
	conf := &Config{}
	conf.ApplyEnvLookup(func(key string) (string, bool) {
		values := map[string]string{
			"GO_RAG_VECTOR_ENABLED":              "true",
			"GO_RAG_VECTOR_BACKEND":              "es",
			"GO_RAG_VECTOR_ADDRESS":              "http://localhost:9200",
			"GO_RAG_VECTOR_INDEX_PREFIX":         "knowledge-",
			"GO_RAG_VECTOR_CONTENT_FIELD":        "body",
			"GO_RAG_VECTOR_CONTENT_VECTOR_FIELD": "body_vector",
			"GO_RAG_VECTOR_QA_CONTENT_FIELD":     "questions",
			"GO_RAG_VECTOR_QA_CONTENT_VECTOR_FIELD": "questions_vector",
			"GO_RAG_VECTOR_KNOWLEDGE_FIELD":      "knowledge",
			"GO_RAG_VECTOR_EXT_FIELD":            "metadata",
			"GO_RAG_VECTOR_DIMENSIONS":           "2048",
			"GO_RAG_VECTOR_EMBEDDING_MODEL":      "text-embedding-3-small",
			"GO_RAG_VECTOR_EMBEDDING_API_KEY":    "vector-key",
			"GO_RAG_VECTOR_EMBEDDING_BASE_URL":   "https://example.com/v1",
			"GO_RAG_RERANK_ENABLED":              "true",
			"GO_RAG_RERANK_BASE_URL":             "https://rerank.example.com/v1",
			"GO_RAG_RERANK_API_KEY":              "rerank-key",
			"GO_RAG_RERANK_MODEL":                "bge-reranker-v2",
			"GO_RAG_RERANK_TOP_N":                "8",
			"GO_RAG_RERANK_MIN_SCORE":            "0.35",
			"GO_RAG_QUALITY_QA_ENABLED":          "true",
			"GO_RAG_QUALITY_QA_QUESTION_COUNT":   "4",
			"GO_RAG_QUALITY_GRADER_ENABLED":      "true",
		}
		value, ok := values[key]
		return value, ok
	})
	if !conf.Vector.Enabled {
		t.Fatal("expected vector enabled from env")
	}
	if conf.Vector.Address != "http://localhost:9200" {
		t.Fatalf("expected vector address from env, got %s", conf.Vector.Address)
	}
	if conf.Vector.IndexPrefix != "knowledge-" {
		t.Fatalf("expected vector index prefix from env, got %s", conf.Vector.IndexPrefix)
	}
	if conf.Vector.ContentField != "body" {
		t.Fatalf("expected vector content field from env, got %s", conf.Vector.ContentField)
	}
	if conf.Vector.ContentVectorField != "body_vector" {
		t.Fatalf("expected vector content vector field from env, got %s", conf.Vector.ContentVectorField)
	}
	if conf.Vector.QAContentField != "questions" {
		t.Fatalf("expected vector qa content field from env, got %s", conf.Vector.QAContentField)
	}
	if conf.Vector.QAContentVectorField != "questions_vector" {
		t.Fatalf("expected vector qa content vector field from env, got %s", conf.Vector.QAContentVectorField)
	}
	if conf.Vector.KnowledgeField != "knowledge" {
		t.Fatalf("expected vector knowledge field from env, got %s", conf.Vector.KnowledgeField)
	}
	if conf.Vector.ExtField != "metadata" {
		t.Fatalf("expected vector ext field from env, got %s", conf.Vector.ExtField)
	}
	if conf.Vector.Dimensions != 2048 {
		t.Fatalf("expected vector dimensions from env, got %d", conf.Vector.Dimensions)
	}
	if conf.Vector.EmbeddingModel != "text-embedding-3-small" {
		t.Fatalf("expected embedding model from env, got %s", conf.Vector.EmbeddingModel)
	}
	if conf.Vector.EmbeddingAPIKey != "vector-key" {
		t.Fatalf("expected embedding api key from env, got %s", conf.Vector.EmbeddingAPIKey)
	}
	if conf.Vector.EmbeddingBaseURL != "https://example.com/v1" {
		t.Fatalf("expected embedding base url from env, got %s", conf.Vector.EmbeddingBaseURL)
	}
	if !conf.Rerank.Enabled {
		t.Fatal("expected rerank enabled from env")
	}
	if conf.Rerank.BaseURL != "https://rerank.example.com/v1" {
		t.Fatalf("expected rerank base url from env, got %s", conf.Rerank.BaseURL)
	}
	if conf.Rerank.APIKey != "rerank-key" {
		t.Fatalf("expected rerank api key from env, got %s", conf.Rerank.APIKey)
	}
	if conf.Rerank.Model != "bge-reranker-v2" {
		t.Fatalf("expected rerank model from env, got %s", conf.Rerank.Model)
	}
	if conf.Rerank.TopN != 8 {
		t.Fatalf("expected rerank top n from env, got %d", conf.Rerank.TopN)
	}
	if conf.Rerank.MinScore != 0.35 {
		t.Fatalf("expected rerank min score from env, got %v", conf.Rerank.MinScore)
	}
	if !conf.Quality.QA.Enabled {
		t.Fatal("expected qa enabled from env")
	}
	if conf.Quality.QA.QuestionCount != 4 {
		t.Fatalf("expected qa question count from env, got %d", conf.Quality.QA.QuestionCount)
	}
	if !conf.Quality.Grader.Enabled {
		t.Fatal("expected grader enabled from env")
	}
}
