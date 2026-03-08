package service

import (
	"context"
	"errors"
	"strings"

	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
	domainrepo "github.com/gaohao-creator/go-rag/internal/domain/repository"
)

type CreateDocumentInput struct {
	KnowledgeBaseName string
	FileName          string
	Status            int
}

type DocumentService struct {
	repo domainrepo.DocumentRepository
}

func NewDocumentService(repo domainrepo.DocumentRepository) *DocumentService {
	return &DocumentService{repo: repo}
}

func (s *DocumentService) Create(ctx context.Context, in CreateDocumentInput) (int64, error) {
	if strings.TrimSpace(in.KnowledgeBaseName) == "" {
		return 0, errors.New("知识库名称不能为空")
	}
	if strings.TrimSpace(in.FileName) == "" {
		return 0, errors.New("文件名不能为空")
	}
	return s.repo.Create(ctx, domainmodel.Document{KnowledgeBaseName: in.KnowledgeBaseName, FileName: in.FileName, Status: in.Status})
}

func (s *DocumentService) ListByKnowledgeBase(ctx context.Context, knowledgeBaseName string, page int, size int) ([]domainmodel.Document, int64, error) {
	return s.repo.ListByKnowledgeBase(ctx, knowledgeBaseName, page, size)
}

func (s *DocumentService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("文档 ID 非法")
	}
	return s.repo.Delete(ctx, id)
}
