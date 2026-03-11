package es

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"
	es8indexer "github.com/cloudwego/eino-ext/components/indexer/es8"
	es8retriever "github.com/cloudwego/eino-ext/components/retriever/es8"
	"github.com/cloudwego/eino-ext/components/retriever/es8/search_mode"
	"github.com/cloudwego/eino/components/embedding"
	einoindexer "github.com/cloudwego/eino/components/indexer"
	einoretriever "github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	appconfig "github.com/gaohao-creator/go-rag/config"
)

type indexerComponent struct {
	client   *elasticsearch.Client
	embedder embedding.Embedder
	config   appconfig.VectorConfig
}

func newIndexerComponent(client *elasticsearch.Client, embedder embedding.Embedder, config appconfig.VectorConfig) einoindexer.Indexer {
	return &indexerComponent{
		client:   client,
		embedder: embedder,
		config:   config,
	}
}

func (c *indexerComponent) Store(ctx context.Context, docs []*schema.Document, opts ...einoindexer.Option) ([]string, error) {
	indexName, err := indexFromContext(ctx)
	if err != nil {
		return nil, err
	}
	component, err := es8indexer.NewIndexer(ctx, &es8indexer.IndexerConfig{
		Client:           c.client,
		Index:            indexName,
		BatchSize:        10,
		DocumentToFields: c.documentToFields,
		Embedding:        c.embedder,
	})
	if err != nil {
		return nil, err
	}
	return component.Store(ctx, docs, opts...)
}

func (c *indexerComponent) documentToFields(_ context.Context, doc *schema.Document) (map[string]es8indexer.FieldValue, error) {
	extra := stringMetadata(doc.MetaData, c.config.ExtField)
	knowledgeName := stringMetadata(doc.MetaData, c.config.KnowledgeField)
	fields := map[string]es8indexer.FieldValue{
		c.config.ContentField: {
			Value:    doc.Content,
			EmbedKey: c.config.ContentVectorField,
		},
		c.config.ExtField: {
			Value: extra,
		},
		c.config.KnowledgeField: {
			Value: knowledgeName,
		},
	}
	if qaContent := stringMetadata(doc.MetaData, c.config.QAContentField); qaContent != "" {
		fields[c.config.QAContentField] = es8indexer.FieldValue{
			Value:    qaContent,
			EmbedKey: c.config.QAContentVectorField,
		}
	}
	return fields, nil
}

type retrieverComponent struct {
	client           *elasticsearch.Client
	embedder         embedding.Embedder
	config           appconfig.VectorConfig
	queryVectorField string
}

func newRetrieverComponent(
	client *elasticsearch.Client,
	embedder embedding.Embedder,
	config appconfig.VectorConfig,
	queryVectorField string,
) einoretriever.Retriever {
	return &retrieverComponent{
		client:           client,
		embedder:         embedder,
		config:           config,
		queryVectorField: queryVectorField,
	}
}

func (c *retrieverComponent) Retrieve(ctx context.Context, query string, opts ...einoretriever.Option) ([]*schema.Document, error) {
	indexName, err := indexFromContext(ctx)
	if err != nil {
		return nil, err
	}
	component, err := es8retriever.NewRetriever(ctx, &es8retriever.RetrieverConfig{
		Client: c.client,
		Index:  indexName,
		SearchMode: search_mode.SearchModeDenseVectorSimilarity(
			search_mode.DenseVectorSimilarityTypeCosineSimilarity,
			c.queryVectorField,
		),
		ResultParser: buildHitParser(c.config),
		Embedding:    c.embedder,
	})
	if err != nil {
		return nil, err
	}
	return component.Retrieve(ctx, query, opts...)
}

func buildHitParser(config appconfig.VectorConfig) func(ctx context.Context, hit types.Hit) (*schema.Document, error) {
	return func(_ context.Context, hit types.Hit) (*schema.Document, error) {
		doc := &schema.Document{
			MetaData: map[string]any{},
		}
		if hit.Id_ != nil {
			doc.ID = *hit.Id_
		}

		var src map[string]any
		if err := sonic.Unmarshal(hit.Source_, &src); err != nil {
			return nil, err
		}

		for field, value := range src {
			switch field {
			case config.ContentField:
				text, _ := value.(string)
				doc.Content = text
			case config.ContentVectorField:
				items, ok := value.([]interface{})
				if !ok {
					continue
				}
				vector := make([]float64, 0, len(items))
				for _, item := range items {
					switch number := item.(type) {
					case float64:
						vector = append(vector, number)
					case float32:
						vector = append(vector, float64(number))
					}
				}
				doc.WithDenseVector(vector)
			case config.QAContentField:
				text, _ := value.(string)
				doc.MetaData[config.QAContentField] = text
			case config.QAContentVectorField:
				continue
			case config.ExtField:
				text, _ := value.(string)
				doc.MetaData[config.ExtField] = text
			case config.KnowledgeField:
				text, _ := value.(string)
				doc.MetaData[config.KnowledgeField] = text
			default:
				return nil, fmt.Errorf("unexpected field=%s", field)
			}
		}
		if hit.Score_ != nil {
			doc.WithScore(float64(*hit.Score_))
		}
		return doc, nil
	}
}
