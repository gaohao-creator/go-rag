package service

import (
	"context"
	"fmt"
	"io"
	"strings"

	chatmodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	openai "github.com/cloudwego/eino-ext/components/model/openai"
	appconfig "github.com/gaohao-creator/go-rag/config"
	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
	domainservice "github.com/gaohao-creator/go-rag/internal/domain/service"
)

type rawChatModel interface {
	Generate(ctx context.Context, input []*schema.Message, opts ...chatmodel.Option) (*schema.Message, error)
	Stream(ctx context.Context, input []*schema.Message, opts ...chatmodel.Option) (*schema.StreamReader[*schema.Message], error)
}

type OpenAICompatibleChatModel struct {
	model        rawChatModel
	systemPrompt string
}

func NewOpenAICompatibleChatModel(model rawChatModel, systemPrompt string) *OpenAICompatibleChatModel {
	return &OpenAICompatibleChatModel{model: model, systemPrompt: systemPrompt}
}

func NewOpenAICompatibleChatModelFromConfig(ctx context.Context, conf appconfig.ChatConfig) (domainservice.ChatModel, error) {
	if strings.TrimSpace(conf.APIKey) == "" {
		return nil, fmt.Errorf("chat.api_key 不能为空")
	}
	if strings.TrimSpace(conf.Model) == "" {
		return nil, fmt.Errorf("chat.model 不能为空")
	}
	client, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  conf.APIKey,
		BaseURL: conf.BaseURL,
		Model:   conf.Model,
	})
	if err != nil {
		return nil, err
	}
	return NewOpenAICompatibleChatModel(client, conf.SystemPrompt), nil
}

func (m *OpenAICompatibleChatModel) Generate(ctx context.Context, in domainservice.ChatGenerateInput) (string, error) {
	message, err := m.model.Generate(ctx, m.buildMessages(in))
	if err != nil {
		return "", err
	}
	return message.Content, nil
}

func (m *OpenAICompatibleChatModel) GeneratePrompt(ctx context.Context, in domainservice.PromptGenerateInput) (string, error) {
	message, err := m.model.Generate(ctx, []*schema.Message{
		schema.SystemMessage(strings.TrimSpace(in.SystemPrompt)),
		schema.UserMessage(strings.TrimSpace(in.UserPrompt)),
	})
	if err != nil {
		return "", err
	}
	return message.Content, nil
}

func (m *OpenAICompatibleChatModel) GenerateStream(ctx context.Context, in domainservice.ChatGenerateInput) ([]string, error) {
	reader, err := m.model.Stream(ctx, m.buildMessages(in))
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	chunks := make([]string, 0)
	for {
		chunk, recvErr := reader.Recv()
		if recvErr == io.EOF {
			break
		}
		if recvErr != nil {
			return nil, recvErr
		}
		if chunk == nil || strings.TrimSpace(chunk.Content) == "" {
			continue
		}
		chunks = append(chunks, chunk.Content)
	}
	return chunks, nil
}

func (m *OpenAICompatibleChatModel) buildMessages(in domainservice.ChatGenerateInput) []*schema.Message {
	messages := make([]*schema.Message, 0, len(in.History)+2)
	messages = append(messages, schema.SystemMessage(m.composeSystemPrompt(in.References)))
	for _, history := range in.History {
		switch history.Role {
		case domainmodel.MessageRoleSystem:
			messages = append(messages, schema.SystemMessage(history.Content))
		case domainmodel.MessageRoleAssistant:
			messages = append(messages, schema.AssistantMessage(history.Content, nil))
		default:
			messages = append(messages, schema.UserMessage(history.Content))
		}
	}
	if len(in.History) == 0 || in.History[len(in.History)-1].Content != in.Question || in.History[len(in.History)-1].Role != domainmodel.MessageRoleUser {
		messages = append(messages, schema.UserMessage(in.Question))
	}
	return messages
}

func (m *OpenAICompatibleChatModel) composeSystemPrompt(references []domainmodel.RetrievedChunk) string {
	builder := strings.Builder{}
	builder.WriteString(strings.TrimSpace(m.systemPrompt))
	if len(references) > 0 {
		builder.WriteString("\n\n参考内容：\n")
		for index, ref := range references {
			builder.WriteString(fmt.Sprintf("[%d] %s\n", index+1, strings.TrimSpace(ref.Content)))
		}
	}
	return builder.String()
}
