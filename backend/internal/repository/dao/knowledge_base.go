package dao

import (
	"context"

	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
	daoentity "github.com/gaohao-creator/go-rag/internal/repository/dao/entity"
	"gorm.io/gorm"
)

type KnowledgeBaseDAO struct {
	db *gorm.DB
}

func NewKnowledgeBaseDAO(db *gorm.DB) *KnowledgeBaseDAO {
	return &KnowledgeBaseDAO{db: db}
}

func (d *KnowledgeBaseDAO) Create(ctx context.Context, kb *daoentity.KnowledgeBase) (int64, error) {
	if err := d.db.WithContext(ctx).Create(kb).Error; err != nil {
		return 0, err
	}
	return kb.ID, nil
}

func (d *KnowledgeBaseDAO) GetByID(ctx context.Context, id int64) (*daoentity.KnowledgeBase, error) {
	var kb daoentity.KnowledgeBase
	if err := d.db.WithContext(ctx).First(&kb, id).Error; err != nil {
		return nil, err
	}
	return &kb, nil
}

func (d *KnowledgeBaseDAO) List(ctx context.Context, filter domainmodel.KnowledgeBaseFilter) ([]daoentity.KnowledgeBase, error) {
	query := d.db.WithContext(ctx).Model(&daoentity.KnowledgeBase{})
	if filter.Name != nil {
		query = query.Where("name = ?", *filter.Name)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.Category != nil {
		query = query.Where("category = ?", *filter.Category)
	}
	var list []daoentity.KnowledgeBase
	if err := query.Order("id asc").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (d *KnowledgeBaseDAO) Update(ctx context.Context, id int64, patch domainmodel.KnowledgeBasePatch) error {
	updates := map[string]any{}
	if patch.Name != nil {
		updates["name"] = *patch.Name
	}
	if patch.Description != nil {
		updates["description"] = *patch.Description
	}
	if patch.Category != nil {
		updates["category"] = *patch.Category
	}
	if patch.Status != nil {
		updates["status"] = *patch.Status
	}
	if len(updates) == 0 {
		return nil
	}
	return d.db.WithContext(ctx).Model(&daoentity.KnowledgeBase{}).Where("id = ?", id).Updates(updates).Error
}

func (d *KnowledgeBaseDAO) Delete(ctx context.Context, id int64) error {
	return d.db.WithContext(ctx).Delete(&daoentity.KnowledgeBase{}, id).Error
}
