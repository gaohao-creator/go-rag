package ioc

import (
	"fmt"
	"os"

	appconfig "github.com/gaohao-creator/go-rag/config"
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
	conf.ApplyDefaults()
	return conf, nil
}
