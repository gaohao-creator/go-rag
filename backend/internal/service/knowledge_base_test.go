package service

import (
	"context"
	"errors"
	"testing"

	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
)

type fakeKnowledgeBaseRepository struct{}

func (f *fakeKnowledgeBaseRepository) Create(_ context.Context, kb domainmodel.KnowledgeBase) (int64, error) {
	if kb.Status != 1 {
		return 0, errors.New("unexpected status")
	}
	return 1, nil
}

func (f *fakeKnowledgeBaseRepository) GetByID(_ context.Context, _ int64) (*domainmodel.KnowledgeBase, error) {
	return nil, nil
}

func (f *fakeKnowledgeBaseRepository) List(_ context.Context, _ domainmodel.KnowledgeBaseFilter) ([]domainmodel.KnowledgeBase, error) {
	return nil, nil
}

func (f *fakeKnowledgeBaseRepository) Update(_ context.Context, _ int64, _ domainmodel.KnowledgeBasePatch) error {
	return nil
}

func (f *fakeKnowledgeBaseRepository) Delete(_ context.Context, _ int64) error {
	return nil
}

func TestKnowledgeBaseService_Create(t *testing.T) {
	svc := NewKnowledgeBaseService(&fakeKnowledgeBaseRepository{})
	id, err := svc.Create(context.Background(), CreateKnowledgeBaseInput{
		Name:        "demo",
		Description: "desc",
		Category:    "general",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if id == 0 {
		t.Fatal("expected inserted id")
	}
}
