package dto

import "time"

type ChunkListRequest struct {
	KnowledgeDocID int64 `form:"knowledge_doc_id" binding:"required"`
	Page           int   `form:"page"`
	Size           int   `form:"size"`
}

type ChunkView struct {
	ID             int64     `json:"id"`
	KnowledgeDocID int64     `json:"knowledgeDocId"`
	ChunkID        string    `json:"chunkId"`
	Content        string    `json:"content"`
	Ext            string    `json:"ext"`
	Status         int       `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

type ChunkListResponse struct {
	Data  []ChunkView `json:"data"`
	Total int64       `json:"total"`
	Page  int         `json:"page"`
	Size  int         `json:"size"`
}

type ChunkDeleteRequest struct {
	ID int64 `form:"id" binding:"required"`
}

type ChunkStatusUpdateRequest struct {
	IDs    []int64 `json:"ids" binding:"required"`
	Status int     `json:"status" binding:"required"`
}

type ChunkContentUpdateRequest struct {
	ID      int64  `json:"id" binding:"required"`
	Content string `json:"content" binding:"required"`
}
