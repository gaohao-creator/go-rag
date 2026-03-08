package main

import (
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/gaohao-creator/go-rag/ioc"
)

func buildApplication() (*ioc.App, error) {
	return ioc.NewApp(resolveConfigPath())
}

func buildApp() (*gin.Engine, error) {
	application, err := buildApplication()
	if err != nil {
		return nil, err
	}
	return application.Router, nil
}

func resolveConfigPath() string {
	if configPath := os.Getenv("GO_RAG_CONFIG"); configPath != "" {
		return configPath
	}
	return filepath.Join("config", "config.yaml")
}

func main() {
	application, err := buildApplication()
	if err != nil {
		panic(err)
	}
	if err = application.Router.Run(":" + application.Config.HTTP.Port); err != nil {
		panic(err)
	}
}
