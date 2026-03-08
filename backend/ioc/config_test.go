package ioc

import (
	"path/filepath"
	"testing"
)

func TestNewConfig_LoadsDefaultConfig(t *testing.T) {
	configPath := filepath.Join("..", "config", "config.yaml")
	conf, err := NewConfig(configPath)
	if err != nil {
		t.Fatalf("NewConfig returned error: %v", err)
	}
	if conf.HTTP.Port == "" {
		t.Fatal("expected HTTP port")
	}
	if conf.Database.Driver == "" {
		t.Fatal("expected database driver")
	}
}
