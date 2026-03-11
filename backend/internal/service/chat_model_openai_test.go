package service

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
	domainservice "github.com/gaohao-creator/go-rag/internal/domain/service"
)

type fakeRawChatModel struct {
	lastMessages []*schema.Message
	streamChunks []string
}

func (f *fakeRawChatModel) Generate(_ context.Context, input []*schema.Message, _ ...model.Option) (*schema.Message, error) {
	f.lastMessages = input
	return schema.AssistantMessage("真实模型回答", nil), nil
}

func (f *fakeRawChatModel) Stream(_ context.Context, input []*schema.Message, _ ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	f.lastMessages = input
	sr, sw := schema.Pipe[*schema.Message](0)
	go func() {
		defer sw.Close()
		for _, chunk := range f.streamChunks {
			sw.Send(&schema.Message{Role: schema.Assistant, Content: chunk}, nil)
		}
		sw.Send(nil, io.EOF)
	}()
	return sr, nil
}

func TestOpenAICompatibleChatModel_GenerateBuildsMessages(t *testing.T) {
	raw := &fakeRawChatModel{}
	model := NewOpenAICompatibleChatModel(raw, "你是一个专业的 AI 助手")
	answer, err := model.Generate(context.Background(), domainservice.ChatGenerateInput{
		ConvID:   "conv-1",
		Question: "什么是 RAG",
		History: []domainmodel.Message{
			{ConvID: "conv-1", Role: domainmodel.MessageRoleUser, Content: "上一轮问题"},
			{ConvID: "conv-1", Role: domainmodel.MessageRoleAssistant, Content: "上一轮回答"},
		},
		References: []domainmodel.RetrievedChunk{
			{ChunkID: "chunk-1", Content: "RAG 是检索增强生成。", Score: 0.9},
		},
	})
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if answer != "真实模型回答" {
		t.Fatalf("unexpected answer: %s", answer)
	}
	if len(raw.lastMessages) < 3 {
		t.Fatalf("expected at least 3 messages, got %d", len(raw.lastMessages))
	}
	if raw.lastMessages[0].Role != schema.System {
		t.Fatalf("expected first message system, got %s", raw.lastMessages[0].Role)
	}
	if !strings.Contains(raw.lastMessages[0].Content, "RAG 是检索增强生成") {
		t.Fatalf("expected system prompt to contain references, got %s", raw.lastMessages[0].Content)
	}
}

func TestOpenAICompatibleChatModel_GenerateStreamCollectsChunks(t *testing.T) {
	raw := &fakeRawChatModel{streamChunks: []string{"模拟", "流式", "回答"}}
	model := NewOpenAICompatibleChatModel(raw, "你是一个专业的 AI 助手")
	chunks, err := model.GenerateStream(context.Background(), domainservice.ChatGenerateInput{
		ConvID:     "conv-1",
		Question:   "什么是 RAG",
		References: []domainmodel.RetrievedChunk{{ChunkID: "chunk-1", Content: "RAG 是检索增强生成。"}},
	})
	if err != nil {
		t.Fatalf("GenerateStream returned error: %v", err)
	}
	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks, got %d", len(chunks))
	}
	if strings.Join(chunks, "") != "模拟流式回答" {
		t.Fatalf("unexpected stream output: %v", chunks)
	}
}

func TestOpenAICompatibleChatModel_GeneratePromptBuildsSystemAndUserMessages(t *testing.T) {
	raw := &fakeRawChatModel{}
	model := NewOpenAICompatibleChatModel(raw, "你是一个专业的 AI 助手")

	answer, err := model.GeneratePrompt(context.Background(), domainservice.PromptGenerateInput{
		SystemPrompt: "你是一个问题生成助手",
		UserPrompt:   "RAG 是检索增强生成。",
	})
	if err != nil {
		t.Fatalf("GeneratePrompt returned error: %v", err)
	}
	if answer != "真实模型回答" {
		t.Fatalf("unexpected answer: %s", answer)
	}
	if len(raw.lastMessages) != 2 {
		t.Fatalf("expected 2 prompt messages, got %d", len(raw.lastMessages))
	}
	if raw.lastMessages[0].Role != schema.System || raw.lastMessages[0].Content != "你是一个问题生成助手" {
		t.Fatalf("unexpected system message: %+v", raw.lastMessages[0])
	}
	if raw.lastMessages[1].Role != schema.User || raw.lastMessages[1].Content != "RAG 是检索增强生成。" {
		t.Fatalf("unexpected user message: %+v", raw.lastMessages[1])
	}
}
