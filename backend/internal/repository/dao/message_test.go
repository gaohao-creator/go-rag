package dao

import (
	"context"
	"testing"

	daoentity "github.com/gaohao-creator/go-rag/internal/repository/dao/entity"
)

func TestMessageDAO_CreateAndListByConversation(t *testing.T) {
	db := newTestDB(t)
	dao := NewMessageDAO(db)

	err := dao.Create(context.Background(), &daoentity.Message{
		ConvID:  "conv-1",
		Role:    "user",
		Content: "hello",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	messages, err := dao.ListByConversation(context.Background(), "conv-1")
	if err != nil {
		t.Fatalf("ListByConversation returned error: %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}
	if messages[0].Role != "user" || messages[0].Content != "hello" {
		t.Fatalf("unexpected message: %+v", messages[0])
	}
}
