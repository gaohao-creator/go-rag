package grade

import (
	"context"
	"fmt"
	"strings"

	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
	domainservice "github.com/gaohao-creator/go-rag/internal/domain/service"
)

// Grade 负责对回答进行质量判定。
type Grade struct {
	model domainservice.PromptModel
}

// NewGrade 创建回答质量判定器。
func NewGrade(model domainservice.PromptModel) *Grade {
	return &Grade{model: model}
}

// Grade 根据问题、答案和参考内容判断是否通过质量检查。
func (g *Grade) Grade(ctx context.Context, in domainservice.GradeInput) (bool, error) {
	if g.model == nil {
		return false, fmt.Errorf("prompt model 不能为空")
	}
	if strings.TrimSpace(in.Question) == "" {
		return false, fmt.Errorf("问题不能为空")
	}
	if strings.TrimSpace(in.Answer) == "" {
		return false, fmt.Errorf("回答不能为空")
	}
	result, err := g.model.GeneratePrompt(ctx, domainservice.PromptGenerateInput{
		SystemPrompt: "你是一个严格的回答质量检查器。请只根据给定问题、回答和参考内容判断回答是否被参考内容直接支持。只能回答 yes 或 no，不要输出其他内容。",
		UserPrompt:   buildUserPrompt(in.Question, in.Answer, in.References),
	})
	if err != nil {
		return false, err
	}
	return strings.Contains(strings.ToLower(strings.TrimSpace(result)), "yes"), nil
}

func buildUserPrompt(question string, answer string, references []domainmodel.RetrievedChunk) string {
	builder := strings.Builder{}
	builder.WriteString("问题：")
	builder.WriteString(strings.TrimSpace(question))
	builder.WriteString("\n回答：")
	builder.WriteString(strings.TrimSpace(answer))
	builder.WriteString("\n参考内容：")
	for index, reference := range references {
		builder.WriteString(fmt.Sprintf("\n[%d] %s", index+1, strings.TrimSpace(reference.Content)))
	}
	return builder.String()
}
