package dao

import (
	"context"
	"testing"

	daoentity "github.com/gaohao-creator/go-rag/internal/repository/dao/entity"
)

func TestDocumentDAO_CreateAndListByKnowledgeBase(t *testing.T) {
	db := newTestDB(t)
	dao := NewDocumentDAO(db)
	_, err := dao.Create(context.Background(), &daoentity.Document{
		KnowledgeBaseName: "demo",
		FileName:          "readme.md",
		Status:            1,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	documents, total, err := dao.ListByKnowledgeBase(context.Background(), "demo", 1, 10)
	if err != nil {
		t.Fatalf("ListByKnowledgeBase returned error: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total 1, got %d", total)
	}
	if len(documents) != 1 {
		t.Fatalf("expected 1 document, got %d", len(documents))
	}
}

func TestDocumentDAO_UpdateStatus(t *testing.T) {
	db := newTestDB(t)
	dao := NewDocumentDAO(db)
	id, err := dao.Create(context.Background(), &daoentity.Document{
		KnowledgeBaseName: "demo",
		FileName:          "readme.md",
		Status:            0,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if err = dao.UpdateStatus(context.Background(), id, 2); err != nil {
		t.Fatalf("UpdateStatus returned error: %v", err)
	}
	documents, _, err := dao.ListByKnowledgeBase(context.Background(), "demo", 1, 10)
	if err != nil {
		t.Fatalf("ListByKnowledgeBase returned error: %v", err)
	}
	if documents[0].Status != 2 {
		t.Fatalf("expected status 2, got %d", documents[0].Status)
	}
}
