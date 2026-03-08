package dao

import (
	"context"
	"fmt"
	"strings"
	"testing"

	daoentity "github.com/gaohao-creator/go-rag/internal/repository/dao/entity"
	sqlite "github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(strings.ToLower(t.Name()), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("gorm.Open returned error: %v", err)
	}
	if err = AutoMigrate(db); err != nil {
		t.Fatalf("AutoMigrate returned error: %v", err)
	}
	return db
}

func TestKnowledgeBaseDAO_Create(t *testing.T) {
	db := newTestDB(t)
	dao := NewKnowledgeBaseDAO(db)
	id, err := dao.Create(context.Background(), &daoentity.KnowledgeBase{
		Name:        "demo",
		Description: "desc",
		Category:    "general",
		Status:      1,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if id == 0 {
		t.Fatal("expected inserted id")
	}
}
