package dto

import domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"

type RetrieveRequest struct {
	Question      string  `json:"question" binding:"required"`
	TopK          int     `json:"top_k"`
	Score         float64 `json:"score"`
	KnowledgeName string  `json:"knowledge_name" binding:"required"`
}

type RetrieveResponse struct {
	Document []domainmodel.RetrievedChunk `json:"document"`
}
