package dto

import "time"

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

type KnowledgeBaseView struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Status      int       `json:"status"`
	CreateTime  time.Time `json:"createTime"`
	UpdateTime  time.Time `json:"updateTime"`
}

type KnowledgeBaseListResponse struct {
	List []KnowledgeBaseView `json:"list"`
}

type KnowledgeBaseUpdateRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Category    *string `json:"category"`
	Status      *int    `json:"status"`
}
