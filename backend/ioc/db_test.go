package ioc

import (
	"testing"

	appconfig "github.com/gaohao-creator/go-rag/config"
)

func TestNewDB_OpensSQLiteConnection(t *testing.T) {
	conf := &appconfig.Config{}
	conf.Database.Driver = "sqlite"
	conf.Database.DSN = "file::memory:?cache=shared"

	db, err := NewDB(conf)
	if err != nil {
		t.Fatalf("NewDB returned error: %v", err)
	}
	if db == nil {
		t.Fatal("expected non-nil db")
	}
}
