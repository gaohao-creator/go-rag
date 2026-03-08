package model

import "time"

type Document struct {
	ID                int64     `json:"id"`
	KnowledgeBaseName string    `json:"knowledge_base_name"`
	FileName          string    `json:"file_name"`
	Status            int       `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
