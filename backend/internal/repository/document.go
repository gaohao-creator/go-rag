package repository

import (
	"context"

	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
	internaldao "github.com/gaohao-creator/go-rag/internal/repository/dao"
	daoentity "github.com/gaohao-creator/go-rag/internal/repository/dao/entity"
)

type DocumentRepository struct {
	dao      *internaldao.DocumentDAO
	chunkDAO *internaldao.ChunkDAO
}

func NewDocumentRepository(dao *internaldao.DocumentDAO, chunkDAO ...*internaldao.ChunkDAO) *DocumentRepository {
	var chunk *internaldao.ChunkDAO
	if len(chunkDAO) > 0 {
		chunk = chunkDAO[0]
	}
	return &DocumentRepository{dao: dao, chunkDAO: chunk}
}

func (r *DocumentRepository) Create(ctx context.Context, document domainmodel.Document) (int64, error) {
	return r.dao.Create(ctx, &daoentity.Document{KnowledgeBaseName: document.KnowledgeBaseName, FileName: document.FileName, Status: document.Status})
}

func (r *DocumentRepository) ListByKnowledgeBase(ctx context.Context, knowledgeBaseName string, page int, size int) ([]domainmodel.Document, int64, error) {
	documents, total, err := r.dao.ListByKnowledgeBase(ctx, knowledgeBaseName, page, size)
	if err != nil {
		return nil, 0, err
	}
	result := make([]domainmodel.Document, 0, len(documents))
	for _, document := range documents {
		result = append(result, domainmodel.Document{ID: document.ID, KnowledgeBaseName: document.KnowledgeBaseName, FileName: document.FileName, Status: document.Status, CreatedAt: document.CreatedAt, UpdatedAt: document.UpdatedAt})
	}
	return result, total, nil
}

func (r *DocumentRepository) UpdateStatus(ctx context.Context, id int64, status int) error {
	return r.dao.UpdateStatus(ctx, id, status)
}

func (r *DocumentRepository) Delete(ctx context.Context, id int64) error {
	if r.chunkDAO != nil {
		if err := r.chunkDAO.DeleteByDocumentID(ctx, id); err != nil {
			return err
		}
	}
	return r.dao.Delete(ctx, id)
}
