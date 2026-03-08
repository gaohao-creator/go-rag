package entity

import "time"

type Chunk struct {
	ID             int64     `gorm:"column:id;primaryKey;autoIncrement"`
	KnowledgeDocID int64     `gorm:"column:knowledge_doc_id;not null;index"`
	ChunkID        string    `gorm:"column:chunk_id;size:128;not null;index"`
	Content        string    `gorm:"column:content;type:text;not null"`
	Ext            string    `gorm:"column:ext;type:text"`
	Status         int       `gorm:"column:status;not null"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Chunk) TableName() string {
	return "knowledge_chunks"
}
