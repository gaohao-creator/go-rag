package repository

import (
	"context"
	"testing"

	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
	internaldao "github.com/gaohao-creator/go-rag/internal/repository/dao"
)

func newTestMessageRepository(t *testing.T) *MessageRepository {
	t.Helper()
	db := newTestDB(t)
	return NewMessageRepository(internaldao.NewMessageDAO(db))
}

func TestMessageRepository_CreateAndListByConversation(t *testing.T) {
	repo := newTestMessageRepository(t)
	if err := repo.Create(context.Background(), domainmodel.Message{
		ConvID:  "conv-1",
		Role:    domainmodel.MessageRoleUser,
		Content: "hello",
	}); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	messages, err := repo.ListByConversation(context.Background(), "conv-1")
	if err != nil {
		t.Fatalf("ListByConversation returned error: %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}
	if messages[0].Role != domainmodel.MessageRoleUser {
		t.Fatalf("unexpected role: %s", messages[0].Role)
	}
}
