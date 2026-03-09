package dao

import (
	"context"

	daoentity "github.com/gaohao-creator/go-rag/internal/repository/dao/entity"
	"gorm.io/gorm"
)

type MessageDAO struct {
	db *gorm.DB
}

func NewMessageDAO(db *gorm.DB) *MessageDAO {
	return &MessageDAO{db: db}
}

func (d *MessageDAO) Create(ctx context.Context, message *daoentity.Message) error {
	return d.db.WithContext(ctx).Create(message).Error
}

func (d *MessageDAO) ListByConversation(ctx context.Context, convID string) ([]daoentity.Message, error) {
	var list []daoentity.Message
	if err := d.db.WithContext(ctx).
		Where("conv_id = ?", convID).
		Order("id asc").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}
