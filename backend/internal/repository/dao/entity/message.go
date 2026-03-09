package entity

import "time"

type Message struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement"`
	ConvID    string    `gorm:"column:conv_id;size:64;not null;index"`
	Role      string    `gorm:"column:role;size:16;not null"`
	Content   string    `gorm:"column:content;type:text;not null"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (Message) TableName() string {
	return "chat_messages"
}
