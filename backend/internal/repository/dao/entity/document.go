package entity

import "time"

type Document struct {
	ID                int64     `gorm:"column:id;primaryKey;autoIncrement"`
	KnowledgeBaseName string    `gorm:"column:knowledge_base_name;size:50;not null;index"`
	FileName          string    `gorm:"column:file_name;size:255;not null"`
	Status            int       `gorm:"column:status;not null"`
	CreatedAt         time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt         time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Document) TableName() string {
	return "knowledge_documents"
}
