package repository

import (
	"context"

	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
	internaldao "github.com/gaohao-creator/go-rag/internal/repository/dao"
	daoentity "github.com/gaohao-creator/go-rag/internal/repository/dao/entity"
)

type KnowledgeBaseRepository struct {
	dao *internaldao.KnowledgeBaseDAO
}

func NewKnowledgeBaseRepository(dao *internaldao.KnowledgeBaseDAO) *KnowledgeBaseRepository {
	return &KnowledgeBaseRepository{dao: dao}
}

func (r *KnowledgeBaseRepository) Create(ctx context.Context, kb domainmodel.KnowledgeBase) (int64, error) {
	return r.dao.Create(ctx, &daoentity.KnowledgeBase{
		Name:        kb.Name,
		Description: kb.Description,
		Category:    kb.Category,
		Status:      kb.Status,
	})
}

func (r *KnowledgeBaseRepository) GetByID(ctx context.Context, id int64) (*domainmodel.KnowledgeBase, error) {
	kb, err := r.dao.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &domainmodel.KnowledgeBase{ID: kb.ID, Name: kb.Name, Description: kb.Description, Category: kb.Category, Status: kb.Status, CreateTime: kb.CreateTime, UpdateTime: kb.UpdateTime}, nil
}

func (r *KnowledgeBaseRepository) List(ctx context.Context, filter domainmodel.KnowledgeBaseFilter) ([]domainmodel.KnowledgeBase, error) {
	list, err := r.dao.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	result := make([]domainmodel.KnowledgeBase, 0, len(list))
	for _, kb := range list {
		result = append(result, domainmodel.KnowledgeBase{ID: kb.ID, Name: kb.Name, Description: kb.Description, Category: kb.Category, Status: kb.Status, CreateTime: kb.CreateTime, UpdateTime: kb.UpdateTime})
	}
	return result, nil
}

func (r *KnowledgeBaseRepository) Update(ctx context.Context, id int64, patch domainmodel.KnowledgeBasePatch) error {
	return r.dao.Update(ctx, id, patch)
}

func (r *KnowledgeBaseRepository) Delete(ctx context.Context, id int64) error {
	return r.dao.Delete(ctx, id)
}
