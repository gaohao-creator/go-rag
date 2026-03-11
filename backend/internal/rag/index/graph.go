package index

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	doccomponent "github.com/cloudwego/eino/components/document"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	domainservice "github.com/gaohao-creator/go-rag/internal/domain/service"
)

type loadedDocuments struct {
	req  domainservice.IndexRequest
	docs []*schema.Document
}

func buildGraph(
	ctx context.Context,
	fileLoader doccomponent.Loader,
	urlLoader doccomponent.Loader,
	markdownTransformer doccomponent.Transformer,
	recursiveTransformer doccomponent.Transformer,
) (compose.Runnable[domainservice.IndexRequest, []domainservice.IndexedChunk], error) {
	graph := compose.NewGraph[domainservice.IndexRequest, []domainservice.IndexedChunk]()
	if err := graph.AddLambdaNode("validate_request", compose.InvokableLambda(func(_ context.Context, req domainservice.IndexRequest) (domainservice.IndexRequest, error) {
		if strings.TrimSpace(req.URI) == "" {
			return domainservice.IndexRequest{}, fmt.Errorf("文档地址不能为空")
		}
		return req, nil
	})); err != nil {
		return nil, err
	}
	if err := graph.AddLambdaNode("load_documents", compose.InvokableLambda(func(ctx context.Context, req domainservice.IndexRequest) (loadedDocuments, error) {
		loader := fileLoader
		if isURL(req.URI) {
			loader = urlLoader
		}
		if loader == nil {
			return loadedDocuments{}, fmt.Errorf("文档加载器未配置")
		}
		docs, err := loader.Load(ctx, doccomponent.Source{URI: req.URI})
		if err != nil {
			return loadedDocuments{}, fmt.Errorf("加载文档失败: %w", err)
		}
		return loadedDocuments{req: req, docs: docs}, nil
	})); err != nil {
		return nil, err
	}
	if err := graph.AddLambdaNode("transform_documents", compose.InvokableLambda(func(ctx context.Context, payload loadedDocuments) ([]*schema.Document, error) {
		transformer := recursiveTransformer
		if strings.EqualFold(filepath.Ext(payload.req.URI), ".md") {
			transformer = markdownTransformer
		}
		if transformer == nil {
			return nil, fmt.Errorf("文档切分器未配置")
		}
		docs, err := transformer.Transform(ctx, payload.docs)
		if err != nil {
			return nil, fmt.Errorf("切分文档失败: %w", err)
		}
		return docs, nil
	})); err != nil {
		return nil, err
	}
	if err := graph.AddLambdaNode("map_chunks", compose.InvokableLambda(func(_ context.Context, docs []*schema.Document) ([]domainservice.IndexedChunk, error) {
		return toIndexedChunks(docs)
	})); err != nil {
		return nil, err
	}
	if err := graph.AddEdge(compose.START, "validate_request"); err != nil {
		return nil, err
	}
	if err := graph.AddEdge("validate_request", "load_documents"); err != nil {
		return nil, err
	}
	if err := graph.AddEdge("load_documents", "transform_documents"); err != nil {
		return nil, err
	}
	if err := graph.AddEdge("transform_documents", "map_chunks"); err != nil {
		return nil, err
	}
	if err := graph.AddEdge("map_chunks", compose.END); err != nil {
		return nil, err
	}
	return graph.Compile(ctx, compose.WithGraphName("default_indexer"))
}
