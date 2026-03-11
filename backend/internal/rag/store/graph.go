package store

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/compose"
	domainservice "github.com/gaohao-creator/go-rag/internal/domain/service"
)

func buildGraph(
	ctx context.Context,
	base domainservice.ChunkStore,
	generator domainservice.QAGenerator,
) (compose.Runnable[domainservice.ChunkStoreRequest, domainservice.ChunkStoreRequest], error) {
	graph := compose.NewGraph[domainservice.ChunkStoreRequest, domainservice.ChunkStoreRequest]()
	if err := graph.AddLambdaNode("enrich_qa", compose.InvokableLambda(func(ctx context.Context, req domainservice.ChunkStoreRequest) (domainservice.ChunkStoreRequest, error) {
		if len(req.Chunks) == 0 || generator == nil {
			return req, nil
		}
		req.QAContents = buildQAContents(ctx, req.KnowledgeName, req, generator)
		return req, nil
	})); err != nil {
		return nil, err
	}
	if err := graph.AddLambdaNode("store_chunks", compose.InvokableLambda(func(ctx context.Context, req domainservice.ChunkStoreRequest) (domainservice.ChunkStoreRequest, error) {
		if base == nil {
			return req, fmt.Errorf("chunk store 未配置")
		}
		return req, base.Store(ctx, req)
	})); err != nil {
		return nil, err
	}
	if err := graph.AddEdge(compose.START, "enrich_qa"); err != nil {
		return nil, err
	}
	if err := graph.AddEdge("enrich_qa", "store_chunks"); err != nil {
		return nil, err
	}
	if err := graph.AddEdge("store_chunks", compose.END); err != nil {
		return nil, err
	}
	return graph.Compile(ctx, compose.WithGraphName("chunk_store"))
}

func buildQAGraph(
	ctx context.Context,
	model domainservice.PromptModel,
	questionCount int,
) (compose.Runnable[domainservice.QAGenerateInput, string], error) {
	if questionCount <= 0 {
		questionCount = 3
	}
	graph := compose.NewGraph[domainservice.QAGenerateInput, string]()
	if err := graph.AddLambdaNode("validate_input", compose.InvokableLambda(func(_ context.Context, in domainservice.QAGenerateInput) (domainservice.QAGenerateInput, error) {
		if model == nil {
			return domainservice.QAGenerateInput{}, fmt.Errorf("prompt model 不能为空")
		}
		if strings.TrimSpace(in.KnowledgeName) == "" {
			return domainservice.QAGenerateInput{}, fmt.Errorf("知识库名称不能为空")
		}
		if strings.TrimSpace(in.Content) == "" {
			return domainservice.QAGenerateInput{}, fmt.Errorf("内容不能为空")
		}
		return in, nil
	})); err != nil {
		return nil, err
	}
	if err := graph.AddLambdaNode("build_prompt", compose.InvokableLambda(func(_ context.Context, in domainservice.QAGenerateInput) (domainservice.PromptGenerateInput, error) {
		return domainservice.PromptGenerateInput{
			SystemPrompt: buildQASystemPrompt(in.KnowledgeName, questionCount),
			UserPrompt:   in.Content,
		}, nil
	})); err != nil {
		return nil, err
	}
	if err := graph.AddLambdaNode("generate_questions", compose.InvokableLambda(func(ctx context.Context, in domainservice.PromptGenerateInput) (string, error) {
		return model.GeneratePrompt(ctx, in)
	})); err != nil {
		return nil, err
	}
	if err := graph.AddEdge(compose.START, "validate_input"); err != nil {
		return nil, err
	}
	if err := graph.AddEdge("validate_input", "build_prompt"); err != nil {
		return nil, err
	}
	if err := graph.AddEdge("build_prompt", "generate_questions"); err != nil {
		return nil, err
	}
	if err := graph.AddEdge("generate_questions", compose.END); err != nil {
		return nil, err
	}
	return graph.Compile(ctx, compose.WithGraphName("qa_generator"))
}
