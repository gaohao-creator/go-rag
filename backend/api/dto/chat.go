package dto

import domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"

type ChatRequest struct {
	ConvID        string  `json:"conv_id" binding:"required"`
	Question      string  `json:"question" binding:"required"`
	KnowledgeName string  `json:"knowledge_name" binding:"required"`
	TopK          int     `json:"top_k"`
	Score         float64 `json:"score"`
}

type ChatResponse struct {
	Answer     string                       `json:"answer"`
	References []domainmodel.RetrievedChunk `json:"references"`
}
