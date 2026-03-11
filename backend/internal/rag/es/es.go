package es

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/elastic/go-elasticsearch/v8"
	appconfig "github.com/gaohao-creator/go-rag/config"
	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
	domainservice "github.com/gaohao-creator/go-rag/internal/domain/service"
)

// ES 负责封装 Elasticsearch 向量索引与检索组件。
type ES struct {
	config   appconfig.VectorConfig
	client   *elasticsearch.Client
	embedder embedding.Embedder
}

// NewES 创建 ES 向量能力门面。
func NewES(ctx context.Context, config appconfig.VectorConfig) (*ES, error) {
	if !config.Enabled {
		return &ES{config: config}, nil
	}
	client, err := newClient(config)
	if err != nil {
		return nil, err
	}
	embedder, err := newEmbedder(ctx, config)
	if err != nil {
		return nil, err
	}
	return &ES{
		config:   config,
		client:   client,
		embedder: embedder,
	}, nil
}

// NewStore 创建基于 ES 的 chunk store。
func (e *ES) NewStore(ctx context.Context) (domainservice.ChunkStore, error) {
	if e == nil || !e.config.Enabled {
		return nil, nil
	}
	runnable, err := buildIndexerGraph(ctx, func(ctx context.Context) error {
		indexName, err := indexFromContext(ctx)
		if err != nil {
			return err
		}
		return ensureIndex(ctx, e.client, e.config, indexName)
	}, newIndexerComponent(e.client, e.embedder, e.config))
	if err != nil {
		return nil, err
	}
	return &vectorStore{
		config:   e.config,
		runnable: runnable,
	}, nil
}

// NewContentRetriever 创建内容向量检索器。
func (e *ES) NewContentRetriever(ctx context.Context) (domainservice.Retriever, error) {
	return e.newVectorRetriever(ctx, e.config.ContentVectorField)
}

// NewQARetriever 创建 QA 向量检索器。
func (e *ES) NewQARetriever(ctx context.Context) (domainservice.Retriever, error) {
	return e.newVectorRetriever(ctx, e.config.QAContentVectorField)
}

func (e *ES) newVectorRetriever(ctx context.Context, queryVectorField string) (domainservice.Retriever, error) {
	if e == nil || !e.config.Enabled {
		return nil, nil
	}
	runnable, err := buildRetrieverGraph(ctx, newRetrieverComponent(e.client, e.embedder, e.config, queryVectorField))
	if err != nil {
		return nil, err
	}
	return &vectorRetriever{
		config:           e.config,
		queryVectorField: queryVectorField,
		runnable:         runnable,
	}, nil
}

type vectorStore struct {
	config   appconfig.VectorConfig
	runnable compose.Runnable[[]*schema.Document, []string]
}

func (s *vectorStore) Store(ctx context.Context, req domainservice.ChunkStoreRequest) error {
	if !s.config.Enabled || len(req.Chunks) == 0 {
		return nil
	}
	if req.KnowledgeName == "" {
		return fmt.Errorf("知识库名称不能为空")
	}

	indexName := resolveIndexName(s.config.IndexPrefix, req.KnowledgeName)
	_, err := s.runnable.Invoke(withIndex(ctx, indexName), toDocuments(req, s.config))
	return err
}

type vectorRetriever struct {
	config           appconfig.VectorConfig
	queryVectorField string
	runnable         compose.Runnable[string, []*schema.Document]
}

func (r *vectorRetriever) Retrieve(ctx context.Context, req domainservice.RetrieveRequest) ([]domainmodel.RetrievedChunk, error) {
	if !r.config.Enabled {
		return nil, nil
	}
	if req.KnowledgeName == "" {
		return nil, fmt.Errorf("知识库名称不能为空")
	}

	indexName := resolveIndexName(r.config.IndexPrefix, req.KnowledgeName)
	docs, err := r.runnable.Invoke(withIndex(ctx, indexName), req.Question)
	if err != nil {
		return nil, err
	}

	chunks := make([]domainmodel.RetrievedChunk, 0, len(docs))
	for _, doc := range docs {
		if doc == nil {
			continue
		}
		chunks = append(chunks, domainmodel.RetrievedChunk{
			ChunkID: doc.ID,
			Content: doc.Content,
			Ext:     stringMetadata(doc.MetaData, r.config.ExtField),
			Score:   doc.Score(),
		})
	}
	return chunks, nil
}

func toDocuments(req domainservice.ChunkStoreRequest, config appconfig.VectorConfig) []*schema.Document {
	docs := make([]*schema.Document, 0, len(req.Chunks))
	for _, chunk := range req.Chunks {
		metadata := map[string]any{
			config.ExtField:       chunk.Ext,
			config.KnowledgeField: req.KnowledgeName,
		}
		if qaContent, ok := req.QAContents[chunk.ChunkID]; ok && qaContent != "" {
			metadata[config.QAContentField] = qaContent
		}
		docs = append(docs, &schema.Document{
			ID:       chunk.ChunkID,
			Content:  chunk.Content,
			MetaData: metadata,
		})
	}
	return docs
}

func stringMetadata(metadata map[string]any, key string) string {
	if metadata == nil {
		return ""
	}
	value, _ := metadata[key].(string)
	return value
}
