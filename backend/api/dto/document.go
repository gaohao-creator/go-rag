package dto

import domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"

type DocumentListRequest struct {
	KnowledgeName string `form:"knowledge_name" binding:"required"`
	Page          int    `form:"page"`
	Size          int    `form:"size"`
}

type DocumentListResponse struct {
	Data  []domainmodel.Document `json:"data"`
	Total int64                  `json:"total"`
	Page  int                    `json:"page"`
	Size  int                    `json:"size"`
}

type DocumentDeleteRequest struct {
	DocumentID int64 `form:"document_id" binding:"required"`
}
