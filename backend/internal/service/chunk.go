package service

import (
	"context"
	"errors"
	"strings"

	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
	domainrepo "github.com/gaohao-creator/go-rag/internal/domain/repository"
)

type ChunkService struct {
	repo domainrepo.ChunkRepository
}

func NewChunkService(repo domainrepo.ChunkRepository) *ChunkService {
	return &ChunkService{repo: repo}
}

func (s *ChunkService) BatchCreate(ctx context.Context, chunks []domainmodel.Chunk) error {
	if len(chunks) == 0 {
		return errors.New("chunk 列表不能为空")
	}
	return s.repo.BatchCreate(ctx, chunks)
}

func (s *ChunkService) ListByDocumentID(ctx context.Context, documentID int64, page int, size int) ([]domainmodel.Chunk, int64, error) {
	return s.repo.ListByDocumentID(ctx, documentID, page, size)
}

func (s *ChunkService) DeleteByID(ctx context.Context, id int64) error {
	if id <= 0 {
		return errors.New("chunk ID 非法")
	}
	return s.repo.DeleteByID(ctx, id)
}

func (s *ChunkService) UpdateStatusByIDs(ctx context.Context, ids []int64, status int) error {
	if len(ids) == 0 {
		return errors.New("chunk ID 列表不能为空")
	}
	return s.repo.UpdateStatusByIDs(ctx, ids, status)
}

func (s *ChunkService) UpdateContentByID(ctx context.Context, id int64, content string) error {
	if id <= 0 {
		return errors.New("chunk ID 非法")
	}
	if strings.TrimSpace(content) == "" {
		return errors.New("chunk 内容不能为空")
	}
	return s.repo.UpdateContentByID(ctx, id, content)
}
