package service

import (
	"context"
	"errors"
	"strings"

	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
	domainrepo "github.com/gaohao-creator/go-rag/internal/domain/repository"
)

type CreateKnowledgeBaseInput struct {
	Name        string
	Description string
	Category    string
}

type UpdateKnowledgeBaseInput struct {
	ID          int64
	Name        *string
	Description *string
	Category    *string
	Status      *int
}

type ListKnowledgeBasesInput struct {
	Name     *string
	Status   *int
	Category *string
}

type KnowledgeBaseService struct {
	repo domainrepo.KnowledgeBaseRepository
}

func NewKnowledgeBaseService(repo domainrepo.KnowledgeBaseRepository) *KnowledgeBaseService {
	return &KnowledgeBaseService{repo: repo}
}

func (s *KnowledgeBaseService) Create(ctx context.Context, in CreateKnowledgeBaseInput) (int64, error) {
	if strings.TrimSpace(in.Name) == "" {
		return 0, errors.New("知识库名称不能为空")
	}
	if strings.TrimSpace(in.Description) == "" {
		return 0, errors.New("知识库描述不能为空")
	}
	return s.repo.Create(ctx, domainmodel.KnowledgeBase{Name: in.Name, Description: in.Description, Category: in.Category, Status: 1})
}

func (s *KnowledgeBaseService) List(ctx context.Context, in ListKnowledgeBasesInput) ([]domainmodel.KnowledgeBase, error) {
	return s.repo.List(ctx, domainmodel.KnowledgeBaseFilter{Name: in.Name, Status: in.Status, Category: in.Category})
}

func (s *KnowledgeBaseService) GetByID(ctx context.Context, id int64) (*domainmodel.KnowledgeBase, error) {
	if id <= 0 {
		return nil, errors.New("知识库 ID 非法")
	}
	return s.repo.GetByID(ctx, id)
}

func (s *KnowledgeBaseService) Update(ctx context.Context, in UpdateKnowledgeBaseInput) error {
	if in.ID <= 0 {
		return errors.New("知识库 ID 非法")
	}
	return s.repo.Update(ctx, in.ID, domainmodel.KnowledgeBasePatch{Name: in.Name, Description: in.Description, Category: in.Category, Status: in.Status})
}

func (s *KnowledgeBaseService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("知识库 ID 非法")
	}
	return s.repo.Delete(ctx, id)
}
