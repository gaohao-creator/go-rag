package dto

import domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"

type KnowledgeBaseCreateRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
	Category    string `json:"category"`
}

type KnowledgeBaseCreateResponse struct {
	ID int64 `json:"id"`
}

type KnowledgeBaseListRequest struct {
	Name     *string `form:"name"`
	Status   *int    `form:"status"`
	Category *string `form:"category"`
}

type KnowledgeBaseListResponse struct {
	List []domainmodel.KnowledgeBase `json:"list"`
}

type KnowledgeBaseUpdateRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Category    *string `json:"category"`
	Status      *int    `json:"status"`
}
