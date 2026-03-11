package es

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/elastic/go-elasticsearch/v8"
	appconfig "github.com/gaohao-creator/go-rag/config"
	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
	domainservice "github.com/gaohao-creator/go-rag/internal/domain/service"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

func TestDebugVectorStoreAndComponent(t *testing.T) {
	conf, err := loadDebugConfig("../../../config/config.yaml")
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}
	if conf == nil || !conf.Vector.Enabled {
		t.Fatal("vector config is not enabled")
	}

	ctx := context.Background()
	esClient, err := NewES(ctx, conf.Vector)
	if err != nil {
		t.Fatalf("new es client failed: %v", err)
	}

	knowledgeName := fmt.Sprintf("debug-%d", time.Now().UnixNano())
	indexName := resolveIndexName(conf.Vector.IndexPrefix, knowledgeName)
	if err = ensureIndex(ctx, esClient.client, conf.Vector, indexName); err != nil {
		t.Fatalf("ensure index failed: %v", err)
	}

	store, err := esClient.NewStore(ctx)
	if err != nil {
		t.Fatalf("new store failed: %v", err)
	}
	err = store.Store(ctx, domainservice.ChunkStoreRequest{
		KnowledgeName: knowledgeName,
		Chunks: []domainmodel.Chunk{
			{
				ChunkID: "chunk-store-1",
				Content: "这是通过 vectorStore.Store 写入 ES 的测试内容。",
				Ext:     "{\"source\":\"vector-store\"}",
			},
		},
	})
	t.Logf("vectorStore.Store err=%v", err)

	storeCount, countErr := countDocuments(ctx, esClient.client, indexName)
	t.Logf("vectorStore count=%d err=%v", storeCount, countErr)

	componentIndex := indexName + "-component"
	if err = ensureIndex(ctx, esClient.client, conf.Vector, componentIndex); err != nil {
		t.Fatalf("ensure component index failed: %v", err)
	}
	component := newIndexerComponent(esClient.client, esClient.embedder, conf.Vector)
	ids, componentErr := component.Store(
		withIndex(ctx, componentIndex),
		[]*schema.Document{
			{
				ID:      "chunk-component-1",
				Content: "这是通过 indexer component 直接写入 ES 的测试内容。",
				MetaData: map[string]any{
					conf.Vector.ExtField:       "{\"source\":\"component\"}",
					conf.Vector.KnowledgeField: knowledgeName,
				},
			},
		},
	)
	t.Logf("component.Store ids=%v err=%v", ids, componentErr)

	componentCount, componentCountErr := countDocuments(ctx, esClient.client, componentIndex)
	t.Logf("component count=%d err=%v", componentCount, componentCountErr)
}

func loadDebugConfig(path string) (*appconfig.Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	conf := &appconfig.Config{}
	if err = yaml.Unmarshal(content, conf); err != nil {
		return nil, err
	}

	dotenvPath := filepath.Join(filepath.Dir(filepath.Dir(path)), ".env")
	dotenvValues, err := godotenv.Read(dotenvPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	conf.ApplyEnvLookup(func(key string) (string, bool) {
		value, ok := dotenvValues[key]
		return value, ok
	})
	conf.ApplyEnvLookup(os.LookupEnv)
	conf.ApplyDefaults()
	return conf, nil
}

func countDocuments(_ context.Context, client *elasticsearch.Client, indexName string) (int64, error) {
	resp, err := client.Count(client.Count.WithIndex(indexName))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	var payload struct {
		Count int64 `json:"count"`
	}
	if err = json.Unmarshal(body, &payload); err != nil {
		return 0, err
	}
	return payload.Count, nil
}
