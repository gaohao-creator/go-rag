package service

import (
	"context"
	"errors"
	"testing"

	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
)

type fakeDocumentRepository struct{}

func (f *fakeDocumentRepository) Create(_ context.Context, document domainmodel.Document) (int64, error) {
	if document.FileName == "" {
		return 0, errors.New("unexpected document")
	}
	return 1, nil
}

func (f *fakeDocumentRepository) ListByKnowledgeBase(_ context.Context, knowledgeBaseName string, _ int, _ int) ([]domainmodel.Document, int64, error) {
	return []domainmodel.Document{{KnowledgeBaseName: knowledgeBaseName, FileName: "readme.md", Status: 1}}, 1, nil
}

func (f *fakeDocumentRepository) UpdateStatus(_ context.Context, _ int64, _ int) error { return nil }
func (f *fakeDocumentRepository) Delete(_ context.Context, _ int64) error              { return nil }

func TestDocumentService_Create(t *testing.T) {
	svc := NewDocumentService(&fakeDocumentRepository{})
	id, err := svc.Create(context.Background(), CreateDocumentInput{
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
