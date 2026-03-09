package repository

import (
	"context"

	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
)

type MessageRepository interface {
	Create(ctx context.Context, message domainmodel.Message) error
	ListByConversation(ctx context.Context, convID string) ([]domainmodel.Message, error)
}
