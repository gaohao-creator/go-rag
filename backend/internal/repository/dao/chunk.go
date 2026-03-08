package dao

import (
	"context"

	daoentity "github.com/gaohao-creator/go-rag/internal/repository/dao/entity"
	"gorm.io/gorm"
)

type ChunkDAO struct {
	db *gorm.DB
}

func NewChunkDAO(db *gorm.DB) *ChunkDAO {
	return &ChunkDAO{db: db}
}

func (d *ChunkDAO) BatchCreate(ctx context.Context, chunks []*daoentity.Chunk) error {
	if len(chunks) == 0 {
		return nil
	}
	return d.db.WithContext(ctx).Create(chunks).Error
}

func (d *ChunkDAO) ListByDocumentID(ctx context.Context, documentID int64, page int, size int) ([]daoentity.Chunk, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}

	query := d.db.WithContext(ctx).Model(&daoentity.Chunk{}).Where("knowledge_doc_id = ?", documentID)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var chunks []daoentity.Chunk
	if err := query.Order("created_at asc").Offset((page - 1) * size).Limit(size).Find(&chunks).Error; err != nil {
		return nil, 0, err
	}
	return chunks, total, nil
}

func (d *ChunkDAO) DeleteByID(ctx context.Context, id int64) error {
	return d.db.WithContext(ctx).Delete(&daoentity.Chunk{}, id).Error
}

func (d *ChunkDAO) DeleteByDocumentID(ctx context.Context, documentID int64) error {
	return d.db.WithContext(ctx).Where("knowledge_doc_id = ?", documentID).Delete(&daoentity.Chunk{}).Error
}

func (d *ChunkDAO) UpdateStatusByIDs(ctx context.Context, ids []int64, status int) error {
	if len(ids) == 0 {
		return nil
	}
	return d.db.WithContext(ctx).Model(&daoentity.Chunk{}).Where("id IN ?", ids).Update("status", status).Error
}

func (d *ChunkDAO) UpdateContentByID(ctx context.Context, id int64, content string) error {
	return d.db.WithContext(ctx).Model(&daoentity.Chunk{}).Where("id = ?", id).Update("content", content).Error
}
