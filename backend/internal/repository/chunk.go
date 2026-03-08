package repository

import (
	"context"

	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
	internaldao "github.com/gaohao-creator/go-rag/internal/repository/dao"
	daoentity "github.com/gaohao-creator/go-rag/internal/repository/dao/entity"
)

type ChunkRepository struct {
	dao *internaldao.ChunkDAO
}

func NewChunkRepository(dao *internaldao.ChunkDAO) *ChunkRepository {
	return &ChunkRepository{dao: dao}
}

func (r *ChunkRepository) BatchCreate(ctx context.Context, chunks []domainmodel.Chunk) error {
	entities := make([]*daoentity.Chunk, 0, len(chunks))
	for _, chunk := range chunks {
		entities = append(entities, &daoentity.Chunk{KnowledgeDocID: chunk.KnowledgeDocID, ChunkID: chunk.ChunkID, Content: chunk.Content, Ext: chunk.Ext, Status: chunk.Status})
	}
	return r.dao.BatchCreate(ctx, entities)
}

func (r *ChunkRepository) ListByDocumentID(ctx context.Context, documentID int64, page int, size int) ([]domainmodel.Chunk, int64, error) {
	chunks, total, err := r.dao.ListByDocumentID(ctx, documentID, page, size)
	if err != nil {
		return nil, 0, err
	}
	result := make([]domainmodel.Chunk, 0, len(chunks))
	for _, chunk := range chunks {
		result = append(result, domainmodel.Chunk{ID: chunk.ID, KnowledgeDocID: chunk.KnowledgeDocID, ChunkID: chunk.ChunkID, Content: chunk.Content, Ext: chunk.Ext, Status: chunk.Status, CreatedAt: chunk.CreatedAt, UpdatedAt: chunk.UpdatedAt})
	}
	return result, total, nil
}

func (r *ChunkRepository) DeleteByID(ctx context.Context, id int64) error {
	return r.dao.DeleteByID(ctx, id)
}
func (r *ChunkRepository) DeleteByDocumentID(ctx context.Context, documentID int64) error {
	return r.dao.DeleteByDocumentID(ctx, documentID)
}
func (r *ChunkRepository) UpdateStatusByIDs(ctx context.Context, ids []int64, status int) error {
	return r.dao.UpdateStatusByIDs(ctx, ids, status)
}
func (r *ChunkRepository) UpdateContentByID(ctx context.Context, id int64, content string) error {
	return r.dao.UpdateContentByID(ctx, id, content)
}
