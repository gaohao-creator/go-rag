package entity

import "time"

type KnowledgeBase struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement"`
	Name        string    `gorm:"column:name;size:50;not null"`
	Description string    `gorm:"column:description;size:200;not null"`
	Category    string    `gorm:"column:category;size:50"`
	Status      int       `gorm:"column:status;not null"`
	CreateTime  time.Time `gorm:"column:create_time;autoCreateTime"`
	UpdateTime  time.Time `gorm:"column:update_time;autoUpdateTime"`
}

func (KnowledgeBase) TableName() string {
	return "knowledge_base"
}
