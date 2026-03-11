package service

import (
	"context"
	"testing"

	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
	domainservice "github.com/gaohao-creator/go-rag/internal/domain/service"
	apprag "github.com/gaohao-creator/go-rag/internal/rag"
)

type fakeMessageRepository struct {
	messages map[string][]domainmodel.Message
}

func newFakeMessageRepository() *fakeMessageRepository {
	return &fakeMessageRepository{messages: map[string][]domainmodel.Message{}}
}

func (f *fakeMessageRepository) Create(_ context.Context, message domainmodel.Message) error {
	f.messages[message.ConvID] = append(f.messages[message.ConvID], message)
	return nil
}

func (f *fakeMessageRepository) ListByConversation(_ context.Context, convID string) ([]domainmodel.Message, error) {
	return append([]domainmodel.Message(nil), f.messages[convID]...), nil
}

func (f *fakeMessageRepository) CountByConversation(convID string) int {
	return len(f.messages[convID])
}

type fakeChatRetriever struct{}

func (f *fakeChatRetriever) Retrieve(_ context.Context, _ RetrieveInput) ([]domainmodel.RetrievedChunk, error) {
	return []domainmodel.RetrievedChunk{{ChunkID: "chunk-1", Content: "RAG 是检索增强生成。", Score: 0.9}}, nil
}

type fakeChatModel struct{}

func (f *fakeChatModel) Generate(_ context.Context, in domainservice.ChatGenerateInput) (string, error) {
	return "模拟回答: " + in.Question, nil
}

func (f *fakeChatModel) GenerateStream(_ context.Context, in domainservice.ChatGenerateInput) ([]string, error) {
	return []string{"模拟回答: ", in.Question}, nil
}

type fakeChatGrader struct {
	request domainservice.GradeInput
	pass    bool
	err     error
}

func (f *fakeChatGrader) Grade(_ context.Context, in domainservice.GradeInput) (bool, error) {
	f.request = in
	if f.err != nil {
		return false, f.err
	}
	return f.pass, nil
}

func TestChatService_ChatPersistsConversationAndReturnsAnswer(t *testing.T) {
	history := newFakeMessageRepository()
	retriever := &fakeChatRetriever{}
	model := &fakeChatModel{}
	svc := NewChatService(history, retriever, model, apprag.NewRAG(nil, nil, nil, nil))

	result, err := svc.Chat(context.Background(), ChatInput{
		ConvID:        "conv-1",
		Question:      "什么是 RAG",
		KnowledgeName: "demo",
	})
	if err != nil {
		t.Fatalf("Chat returned error: %v", err)
	}
	if result.Answer == "" {
		t.Fatal("expected answer")
	}
	if len(result.References) == 0 {
		t.Fatal("expected references")
	}
	if history.CountByConversation("conv-1") != 2 {
		t.Fatalf("expected 2 persisted messages, got %d", history.CountByConversation("conv-1"))
	}
	if history.messages["conv-1"][0].Role != domainmodel.MessageRoleUser {
		t.Fatalf("expected first role user, got %s", history.messages["conv-1"][0].Role)
	}
	if history.messages["conv-1"][1].Role != domainmodel.MessageRoleAssistant {
		t.Fatalf("expected second role assistant, got %s", history.messages["conv-1"][1].Role)
	}
}

func TestChatService_ChatUsesGraderWhenEnabled(t *testing.T) {
	history := newFakeMessageRepository()
	retriever := &fakeChatRetriever{}
	model := &fakeChatModel{}
	grader := &fakeChatGrader{pass: true}
	svc := NewChatService(history, retriever, model, apprag.NewRAG(nil, nil, nil, grader))

	result, err := svc.Chat(context.Background(), ChatInput{
		ConvID:        "conv-2",
		Question:      "什么是 RAG",
		KnowledgeName: "demo",
	})
	if err != nil {
		t.Fatalf("Chat returned error: %v", err)
	}
	if result.Answer == "" {
		t.Fatal("expected answer")
	}
	if grader.request.Question != "什么是 RAG" {
		t.Fatalf("expected grader question passthrough, got %s", grader.request.Question)
	}
	if grader.request.Answer == "" {
		t.Fatal("expected grader to receive answer")
	}
	if len(grader.request.References) != 1 {
		t.Fatalf("expected grader references, got %+v", grader.request.References)
	}
}

func TestChatService_ChatRejectsAnswerWhenGraderDoesNotPass(t *testing.T) {
	history := newFakeMessageRepository()
	retriever := &fakeChatRetriever{}
	model := &fakeChatModel{}
	grader := &fakeChatGrader{pass: false}
	svc := NewChatService(history, retriever, model, apprag.NewRAG(nil, nil, nil, grader))

	_, err := svc.Chat(context.Background(), ChatInput{
		ConvID:        "conv-3",
		Question:      "什么是 RAG",
		KnowledgeName: "demo",
	})
	if err == nil {
		t.Fatal("expected grader reject error")
	}
	if history.CountByConversation("conv-3") != 1 {
		t.Fatalf("expected only user message persisted, got %d", history.CountByConversation("conv-3"))
	}
}
