package es

import (
	"context"
	"fmt"
	"strings"

	einoindexer "github.com/cloudwego/eino/components/indexer"
	einoretriever "github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

type indexKey struct{}

func buildIndexerGraph(
	ctx context.Context,
	ensure func(context.Context) error,
	component einoindexer.Indexer,
) (compose.Runnable[[]*schema.Document, []string], error) {
	graph := compose.NewGraph[[]*schema.Document, []string]()
	if err := graph.AddLambdaNode("ensure_index", compose.InvokableLambda(func(ctx context.Context, docs []*schema.Document) ([]*schema.Document, error) {
		if err := ensure(ctx); err != nil {
			return nil, err
		}
		return docs, nil
	})); err != nil {
		return nil, err
	}
	if err := graph.AddIndexerNode("index_chunks", component); err != nil {
		return nil, err
	}
	if err := graph.AddEdge(compose.START, "ensure_index"); err != nil {
		return nil, err
	}
	if err := graph.AddEdge("ensure_index", "index_chunks"); err != nil {
		return nil, err
	}
	if err := graph.AddEdge("index_chunks", compose.END); err != nil {
		return nil, err
	}
	return graph.Compile(ctx, compose.WithGraphName("vector_indexer"))
}

func buildRetrieverGraph(
	ctx context.Context,
	component einoretriever.Retriever,
) (compose.Runnable[string, []*schema.Document], error) {
	graph := compose.NewGraph[string, []*schema.Document]()
	if err := graph.AddRetrieverNode("retrieve_chunks", component); err != nil {
		return nil, err
	}
	if err := graph.AddEdge(compose.START, "retrieve_chunks"); err != nil {
		return nil, err
	}
	if err := graph.AddEdge("retrieve_chunks", compose.END); err != nil {
		return nil, err
	}
	return graph.Compile(ctx, compose.WithGraphName("vector_retriever"))
}

func withIndex(ctx context.Context, indexName string) context.Context {
	return context.WithValue(ctx, indexKey{}, strings.TrimSpace(indexName))
}

func indexFromContext(ctx context.Context) (string, error) {
	if ctx == nil {
		return "", fmt.Errorf("上下文不能为空")
	}
	indexName, _ := ctx.Value(indexKey{}).(string)
	indexName = strings.TrimSpace(indexName)
	if indexName == "" {
		return "", fmt.Errorf("缺少向量索引名")
	}
	return indexName, nil
}

func resolveIndexName(prefix string, knowledgeName string) string {
	sanitized := sanitizeKnowledgeName(knowledgeName)
	if sanitized == "" {
		sanitized = "kb"
	}
	return prefix + sanitized
}

func sanitizeKnowledgeName(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return ""
	}

	var builder strings.Builder
	lastDash := false
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
			lastDash = false
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
			lastDash = false
		default:
			if builder.Len() == 0 || lastDash {
				continue
			}
			builder.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(builder.String(), "-")
}
