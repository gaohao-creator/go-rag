package ioc

import (
	"fmt"
	"os"
	"path/filepath"

	appconfig "github.com/gaohao-creator/go-rag/config"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

func NewConfig(path string) (*appconfig.Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	conf := &appconfig.Config{}
	if err = yaml.Unmarshal(content, conf); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	dotenvPath := resolveDotEnvPath(path)
	dotenvValues, err := loadDotEnv(dotenvPath)
	if err != nil {
		return nil, fmt.Errorf("读取 .env 失败: %w", err)
	}
	conf.ApplyEnvLookup(func(key string) (string, bool) {
		value, ok := dotenvValues[key]
		return value, ok
	})
	conf.ApplyEnvLookup(os.LookupEnv)
	conf.ApplyDefaults()
	return conf, nil
}

func resolveDotEnvPath(configPath string) string {
	configDir := filepath.Dir(configPath)
	if filepath.Base(configDir) == "config" {
		return filepath.Join(filepath.Dir(configDir), ".env")
	}
	return filepath.Join(configDir, ".env")
}

func loadDotEnv(path string) (map[string]string, error) {
	values, err := godotenv.Read(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, err
	}
	return values, nil
}
