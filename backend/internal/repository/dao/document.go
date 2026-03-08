package dao

import (
	"context"

	daoentity "github.com/gaohao-creator/go-rag/internal/repository/dao/entity"
	"gorm.io/gorm"
)

type DocumentDAO struct {
	db *gorm.DB
}

func NewDocumentDAO(db *gorm.DB) *DocumentDAO {
	return &DocumentDAO{db: db}
}

func (d *DocumentDAO) Create(ctx context.Context, document *daoentity.Document) (int64, error) {
	if err := d.db.WithContext(ctx).Create(document).Error; err != nil {
		return 0, err
	}
	return document.ID, nil
}

func (d *DocumentDAO) ListByKnowledgeBase(ctx context.Context, knowledgeBaseName string, page int, size int) ([]daoentity.Document, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}

	query := d.db.WithContext(ctx).Model(&daoentity.Document{}).Where("knowledge_base_name = ?", knowledgeBaseName)
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var documents []daoentity.Document
	if err := query.Order("created_at desc").Offset((page - 1) * size).Limit(size).Find(&documents).Error; err != nil {
		return nil, 0, err
	}
	return documents, total, nil
}

func (d *DocumentDAO) UpdateStatus(ctx context.Context, id int64, status int) error {
	return d.db.WithContext(ctx).Model(&daoentity.Document{}).Where("id = ?", id).Update("status", status).Error
}

func (d *DocumentDAO) Delete(ctx context.Context, id int64) error {
	return d.db.WithContext(ctx).Delete(&daoentity.Document{}, id).Error
}
