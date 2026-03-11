package dto

import "github.com/cloudwego/eino/schema"

type RetrieveRequest struct {
	Question      string  `json:"question" binding:"required"`
	TopK          int     `json:"top_k"`
	Score         float64 `json:"score"`
	KnowledgeName string  `json:"knowledge_name" binding:"required"`
}

type RetrieveResponse struct {
	Document []*schema.Document `json:"document"`
}

type RetrieveDifyRequest struct {
	KnowledgeID      string             `json:"knowledge_id" binding:"required"`
	Query            string             `json:"query" binding:"required"`
	RetrievalSetting *RetrievalSettings `json:"retrieval_setting" binding:"required"`
}

type RetrievalSettings struct {
	TopK           int     `json:"top_k"`
	ScoreThreshold float64 `json:"score_threshold"`
}

type RetrieveDifyResponse struct {
	Records []RetrieveDifyRecord `json:"records"`
}

type RetrieveDifyRecord struct {
	Metadata *RetrieveDifyMetadata `json:"metadata,omitempty"`
	Score    float64               `json:"score"`
	Title    string                `json:"title"`
	Content  string                `json:"content"`
}

type RetrieveDifyMetadata struct {
	Path        string `json:"path"`
	Description string `json:"description"`
}
