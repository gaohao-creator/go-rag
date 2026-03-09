package repository

import (
	"context"

	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
	internaldao "github.com/gaohao-creator/go-rag/internal/repository/dao"
	daoentity "github.com/gaohao-creator/go-rag/internal/repository/dao/entity"
)

type MessageRepository struct {
	dao *internaldao.MessageDAO
}

func NewMessageRepository(dao *internaldao.MessageDAO) *MessageRepository {
	return &MessageRepository{dao: dao}
}

func (r *MessageRepository) Create(ctx context.Context, message domainmodel.Message) error {
	return r.dao.Create(ctx, &daoentity.Message{
		ConvID:  message.ConvID,
		Role:    message.Role,
		Content: message.Content,
	})
}

func (r *MessageRepository) ListByConversation(ctx context.Context, convID string) ([]domainmodel.Message, error) {
	list, err := r.dao.ListByConversation(ctx, convID)
	if err != nil {
		return nil, err
	}
	result := make([]domainmodel.Message, 0, len(list))
	for _, item := range list {
		result = append(result, domainmodel.Message{
			ID:        item.ID,
			ConvID:    item.ConvID,
			Role:      item.Role,
			Content:   item.Content,
			CreatedAt: item.CreatedAt,
		})
	}
	return result, nil
}
