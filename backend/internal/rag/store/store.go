package store

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/cloudwego/eino/compose"
	domainservice "github.com/gaohao-creator/go-rag/internal/domain/service"
)

// Store 负责 chunk 写入与可选 QA 增强。
type Store struct {
	base     domainservice.ChunkStore
	runnable compose.Runnable[domainservice.ChunkStoreRequest, domainservice.ChunkStoreRequest]
	buildErr error
}

// NewStore 创建带可选 QA 增强的 chunk store 门面。
func NewStore(
	base domainservice.ChunkStore,
	qaModel domainservice.PromptModel,
	qaQuestionCount int,
) *Store {
	if base == nil {
		return nil
	}
	result := &Store{base: base}
	generator := newQAGenerator(qaModel, qaQuestionCount)
	if generator == nil {
		return result
	}
	runnable, err := buildGraph(context.Background(), base, generator)
	result.runnable = runnable
	result.buildErr = err
	return result
}

// Store 将 chunk 写入底层存储，并在启用时补充 QA 内容。
func (s *Store) Store(ctx context.Context, req domainservice.ChunkStoreRequest) error {
	if s == nil || s.base == nil {
		return nil
	}
	if s.runnable == nil {
		return s.base.Store(ctx, req)
	}
	if s.buildErr != nil {
		return fmt.Errorf("构建 chunk store graph 失败: %w", s.buildErr)
	}
	_, err := s.runnable.Invoke(ctx, req)
	return err
}

func newQAGenerator(model domainservice.PromptModel, questionCount int) domainservice.QAGenerator {
	if model == nil {
		return nil
	}
	runnable, err := buildQAGraph(context.Background(), model, questionCount)
	return &qaGenerator{
		runnable: runnable,
		buildErr: err,
	}
}

type qaGenerator struct {
	runnable compose.Runnable[domainservice.QAGenerateInput, string]
	buildErr error
}

func (g *qaGenerator) Generate(ctx context.Context, in domainservice.QAGenerateInput) (string, error) {
	if g.buildErr != nil {
		return "", fmt.Errorf("构建 QA graph 失败: %w", g.buildErr)
	}
	return g.runnable.Invoke(ctx, in)
}

func buildQAContents(
	ctx context.Context,
	knowledgeName string,
	req domainservice.ChunkStoreRequest,
	generator domainservice.QAGenerator,
) map[string]string {
	output := make(map[string]string, len(req.Chunks))
	for _, chunk := range req.Chunks {
		qaContent, err := generator.Generate(ctx, domainservice.QAGenerateInput{
			KnowledgeName: knowledgeName,
			Content:       chunk.Content,
		})
		if err != nil {
			log.Printf("qa generation skipped after failure: %v", err)
			return nil
		}
		if strings.TrimSpace(qaContent) == "" {
			continue
		}
		output[chunk.ChunkID] = qaContent
	}
	if len(output) == 0 {
		return nil
	}
	return output
}

func buildQASystemPrompt(knowledgeName string, questionCount int) string {
	return fmt.Sprintf(
		"你是一个专业的问题生成助手，任务是从给定的文本中提取或生成可能的问题。你不需要回答这些问题，只需生成问题本身。\n"+
			"知识库名字是：《%s》\n\n"+
			"输出格式：\n"+
			"- 每个问题占一行\n"+
			"- 问题必须以问号结尾\n"+
			"- 避免重复或语义相似的问题\n\n"+
			"生成规则：\n"+
			"- 生成的问题必须严格基于文本内容，不能脱离文本虚构。\n"+
			"- 优先生成事实性问题（如谁、何时、何地、如何）。\n"+
			"- 对于复杂文本，可生成多层次问题（基础事实 + 推理问题）。\n"+
			"- 禁止生成主观或开放式问题（如“你认为...？”）。\n"+
			"- 数量控制在%d个",
		strings.TrimSpace(knowledgeName),
		questionCount,
	)
}
