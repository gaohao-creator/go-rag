package es

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	openaiembedder "github.com/cloudwego/eino-ext/components/embedding/openai"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/indices/create"
	"github.com/elastic/go-elasticsearch/v8/typedapi/indices/exists"
	appconfig "github.com/gaohao-creator/go-rag/config"
)

func newClient(config appconfig.VectorConfig) (*elasticsearch.Client, error) {
	address := strings.TrimSpace(config.Address)
	if address == "" {
		return nil, fmt.Errorf("vector.address 不能为空")
	}
	return elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{address},
		Username:  config.Username,
		Password:  config.Password,
	})
}

func newEmbedder(ctx context.Context, config appconfig.VectorConfig) (embedding.Embedder, error) {
	return openaiembedder.NewEmbedder(ctx, &openaiembedder.EmbeddingConfig{
		APIKey:     config.EmbeddingAPIKey,
		BaseURL:    config.EmbeddingBaseURL,
		Model:      config.EmbeddingModel,
		Dimensions: &config.Dimensions,
		Timeout:    30 * time.Second,
	})
}

func ensureIndex(ctx context.Context, client *elasticsearch.Client, config appconfig.VectorConfig, indexName string) error {
	existsFlag, err := exists.NewExistsFunc(client)(indexName).Do(ctx)
	if err != nil {
		return fmt.Errorf("检查 ES 索引失败: %w", err)
	}
	if existsFlag {
		return nil
	}

	payload, err := json.Marshal(map[string]any{
		"mappings": buildMappings(config),
	})
	if err != nil {
		return fmt.Errorf("序列化 ES mapping 失败: %w", err)
	}
	_, err = create.NewCreateFunc(client)(indexName).Raw(bytes.NewReader(payload)).Do(ctx)
	if err != nil && !strings.Contains(err.Error(), "resource_already_exists_exception") {
		return fmt.Errorf("创建 ES 索引失败: %w", err)
	}
	return nil
}
