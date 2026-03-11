package dto

import "time"

type DocumentListRequest struct {
	KnowledgeName string `form:"knowledge_name" binding:"required"`
	Page          int    `form:"page"`
	Size          int    `form:"size"`
}

type DocumentView struct {
	ID                int64     `json:"id"`
	KnowledgeBaseName string    `json:"knowledgeBaseName"`
	FileName          string    `json:"fileName"`
	Status            int       `json:"status"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type DocumentListResponse struct {
	Data  []DocumentView `json:"data"`
	Total int64          `json:"total"`
	Page  int            `json:"page"`
	Size  int            `json:"size"`
}

type DocumentDeleteRequest struct {
	DocumentID int64 `form:"document_id" binding:"required"`
}
