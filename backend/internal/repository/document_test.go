package repository

import (
	"context"
	"testing"

	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
	internaldao "github.com/gaohao-creator/go-rag/internal/repository/dao"
)

func newTestDocumentRepository(t *testing.T) *DocumentRepository {
	t.Helper()
	db := newTestDB(t)
	return NewDocumentRepository(internaldao.NewDocumentDAO(db))
}

func TestDocumentRepository_Create(t *testing.T) {
	repo := newTestDocumentRepository(t)
	id, err := repo.Create(context.Background(), domainmodel.Document{
		KnowledgeBaseName: "demo",
		FileName:          "readme.md",
		Status:            1,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if id == 0 {
		t.Fatal("expected inserted id")
	}
}

func TestDocumentRepository_UpdateStatus(t *testing.T) {
	repo := newTestDocumentRepository(t)
	id, err := repo.Create(context.Background(), domainmodel.Document{
		KnowledgeBaseName: "demo",
		FileName:          "readme.md",
		Status:            0,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if err = repo.UpdateStatus(context.Background(), id, 2); err != nil {
		t.Fatalf("UpdateStatus returned error: %v", err)
	}
	documents, _, err := repo.ListByKnowledgeBase(context.Background(), "demo", 1, 10)
	if err != nil {
		t.Fatalf("ListByKnowledgeBase returned error: %v", err)
	}
	if documents[0].Status != 2 {
		t.Fatalf("expected status 2, got %d", documents[0].Status)
	}
}
