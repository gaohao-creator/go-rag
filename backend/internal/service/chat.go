package service

import (
	"context"
	"errors"
	"strings"

	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
	domainrepo "github.com/gaohao-creator/go-rag/internal/domain/repository"
	domainservice "github.com/gaohao-creator/go-rag/internal/domain/service"
	apprag "github.com/gaohao-creator/go-rag/internal/rag"
)

type ChatRetriever interface {
	Retrieve(ctx context.Context, in RetrieveInput) ([]domainmodel.RetrievedChunk, error)
}

type ChatInput struct {
	ConvID        string
	Question      string
	KnowledgeName string
	TopK          int
	Score         float64
}

type ChatResult struct {
	Answer     string
	References []domainmodel.RetrievedChunk
}

type ChatStreamResult struct {
	Answer     string
	References []domainmodel.RetrievedChunk
	Chunks     []string
}

type ChatService struct {
	history     domainrepo.MessageRepository
	retriever   ChatRetriever
	model       domainservice.ChatModel
	streamModel domainservice.ChatStreamModel
	rag         *apprag.RAG
}

func NewChatService(
	history domainrepo.MessageRepository,
	retriever ChatRetriever,
	model domainservice.ChatModel,
	rag *apprag.RAG,
) *ChatService {
	var streamModel domainservice.ChatStreamModel
	if value, ok := any(model).(domainservice.ChatStreamModel); ok {
		streamModel = value
	}
	return &ChatService{
		history:     history,
		retriever:   retriever,
		model:       model,
		streamModel: streamModel,
		rag:         rag,
	}
}

func (s *ChatService) Chat(ctx context.Context, in ChatInput) (*ChatResult, error) {
	if err := s.validateChatInput(in); err != nil {
		return nil, err
	}
	if err := s.history.Create(ctx, domainmodel.Message{ConvID: in.ConvID, Role: domainmodel.MessageRoleUser, Content: in.Question}); err != nil {
		return nil, err
	}
	history, err := s.history.ListByConversation(ctx, in.ConvID)
	if err != nil {
		return nil, err
	}
	references, err := s.retriever.Retrieve(ctx, RetrieveInput{Question: in.Question, TopK: in.TopK, Score: in.Score, KnowledgeName: in.KnowledgeName})
	if err != nil {
		return nil, err
	}
	answer, err := s.model.Generate(ctx, domainservice.ChatGenerateInput{ConvID: in.ConvID, Question: in.Question, History: history, References: references})
	if err != nil {
		return nil, err
	}
	if err = s.gradeAnswer(ctx, in.Question, answer, references); err != nil {
		return nil, err
	}
	if err = s.history.Create(ctx, domainmodel.Message{ConvID: in.ConvID, Role: domainmodel.MessageRoleAssistant, Content: answer}); err != nil {
		return nil, err
	}
	return &ChatResult{Answer: answer, References: references}, nil
}

func (s *ChatService) ChatStream(ctx context.Context, in ChatInput) (*ChatStreamResult, error) {
	if err := s.validateChatInput(in); err != nil {
		return nil, err
	}
	if s.streamModel == nil {
		return nil, errors.New("当前模型不支持流式输出")
	}
	if err := s.history.Create(ctx, domainmodel.Message{ConvID: in.ConvID, Role: domainmodel.MessageRoleUser, Content: in.Question}); err != nil {
		return nil, err
	}
	history, err := s.history.ListByConversation(ctx, in.ConvID)
	if err != nil {
		return nil, err
	}
	references, err := s.retriever.Retrieve(ctx, RetrieveInput{Question: in.Question, TopK: in.TopK, Score: in.Score, KnowledgeName: in.KnowledgeName})
	if err != nil {
		return nil, err
	}
	chunks, err := s.streamModel.GenerateStream(ctx, domainservice.ChatGenerateInput{ConvID: in.ConvID, Question: in.Question, History: history, References: references})
	if err != nil {
		return nil, err
	}
	answer := strings.Join(chunks, "")
	if err = s.gradeAnswer(ctx, in.Question, answer, references); err != nil {
		return nil, err
	}
	if err = s.history.Create(ctx, domainmodel.Message{ConvID: in.ConvID, Role: domainmodel.MessageRoleAssistant, Content: answer}); err != nil {
		return nil, err
	}
	return &ChatStreamResult{Answer: answer, References: references, Chunks: chunks}, nil
}

func (s *ChatService) validateChatInput(in ChatInput) error {
	if strings.TrimSpace(in.ConvID) == "" {
		return errors.New("会话 ID 不能为空")
	}
	if strings.TrimSpace(in.Question) == "" {
		return errors.New("问题不能为空")
	}
	if strings.TrimSpace(in.KnowledgeName) == "" {
		return errors.New("知识库名称不能为空")
	}
	return nil
}

func (s *ChatService) gradeAnswer(
	ctx context.Context,
	question string,
	answer string,
	references []domainmodel.RetrievedChunk,
) error {
	pass, err := s.rag.Grade(ctx, domainservice.GradeInput{
		Question:   question,
		Answer:     answer,
		References: references,
	})
	if err != nil {
		return err
	}
	if !pass {
		return errors.New("回答未通过质量检查")
	}
	return nil
}
