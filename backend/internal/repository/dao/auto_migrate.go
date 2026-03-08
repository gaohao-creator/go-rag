package dao

import (
	"fmt"

	daoentity "github.com/gaohao-creator/go-rag/internal/repository/dao/entity"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("数据库连接不能为空")
	}
	return db.AutoMigrate(
		&daoentity.KnowledgeBase{},
		&daoentity.Document{},
		&daoentity.Chunk{},
	)
}
