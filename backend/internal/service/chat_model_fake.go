package service

import (
	"context"
	"fmt"
	"strings"

	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
	domainservice "github.com/gaohao-creator/go-rag/internal/domain/service"
)

type FakeChatModel struct{}

func NewFakeChatModel() *FakeChatModel {
	return &FakeChatModel{}
}

func (m *FakeChatModel) Generate(_ context.Context, in domainservice.ChatGenerateInput) (string, error) {
	question := strings.TrimSpace(in.Question)
	if question == "" {
		return "", fmt.Errorf("问题不能为空")
	}
	referenceSummary := buildReferenceSummary(in.References)
	if referenceSummary == "" {
		referenceSummary = "当前没有可用参考内容。"
	}
	return fmt.Sprintf("模拟回答：你问的是“%s”。参考信息：%s", question, referenceSummary), nil
}

func (m *FakeChatModel) GeneratePrompt(_ context.Context, in domainservice.PromptGenerateInput) (string, error) {
	userPrompt := strings.TrimSpace(in.UserPrompt)
	if userPrompt == "" {
		return "", fmt.Errorf("提示词内容不能为空")
	}
	if strings.TrimSpace(in.SystemPrompt) == "" {
		return "模拟提示词回答：" + userPrompt, nil
	}
	return fmt.Sprintf("模拟提示词回答：%s | %s", strings.TrimSpace(in.SystemPrompt), userPrompt), nil
}

func (m *FakeChatModel) GenerateStream(ctx context.Context, in domainservice.ChatGenerateInput) ([]string, error) {
	answer, err := m.Generate(ctx, in)
	if err != nil {
		return nil, err
	}
	parts := strings.Split(answer, "。")
	chunks := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		chunks = append(chunks, part)
	}
	if len(chunks) == 0 {
		chunks = []string{answer}
	}
	return chunks, nil
}

func buildReferenceSummary(references []domainmodel.RetrievedChunk) string {
	if len(references) == 0 {
		return ""
	}
	parts := make([]string, 0, len(references))
	for _, ref := range references {
		content := strings.TrimSpace(ref.Content)
		if content == "" {
			continue
		}
		parts = append(parts, content)
		if len(parts) == 2 {
			break
		}
	}
	return strings.Join(parts, "；")
}
