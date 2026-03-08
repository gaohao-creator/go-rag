package repository

import (
	"context"
	"fmt"
	"strings"
	"testing"

	domainmodel "github.com/gaohao-creator/go-rag/internal/domain/model"
	internaldao "github.com/gaohao-creator/go-rag/internal/repository/dao"
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
	if err = internaldao.AutoMigrate(db); err != nil {
		t.Fatalf("AutoMigrate returned error: %v", err)
	}
	return db
}

func newTestKnowledgeBaseRepository(t *testing.T) *KnowledgeBaseRepository {
	t.Helper()
	db := newTestDB(t)
	return NewKnowledgeBaseRepository(internaldao.NewKnowledgeBaseDAO(db))
}

func TestKnowledgeBaseRepository_Create(t *testing.T) {
	repo := newTestKnowledgeBaseRepository(t)
	id, err := repo.Create(context.Background(), domainmodel.KnowledgeBase{
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
