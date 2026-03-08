package model

import "time"

type Chunk struct {
	ID             int64     `json:"id"`
	KnowledgeDocID int64     `json:"knowledge_doc_id"`
	ChunkID        string    `json:"chunk_id"`
	Content        string    `json:"content"`
	Ext            string    `json:"ext"`
	Status         int       `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
