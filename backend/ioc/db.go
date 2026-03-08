package ioc

import (
	"fmt"
	"time"

	sqlite "github.com/glebarez/sqlite"
	appconfig "github.com/gaohao-creator/go-rag/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewDB(conf *appconfig.Config) (*gorm.DB, error) {
	if conf == nil {
		return nil, fmt.Errorf("配置不能为空")
	}
	conf.ApplyDefaults()

	var dialector gorm.Dialector
	switch conf.Database.Driver {
	case "mysql":
		dialector = mysql.Open(conf.Database.DSN)
	case "sqlite":
		dialector = sqlite.Open(conf.Database.DSN)
	default:
		return nil, fmt.Errorf("不支持的数据库驱动: %s", conf.Database.Driver)
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("初始化数据库失败: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("获取底层数据库连接失败: %w", err)
	}
	sqlDB.SetMaxIdleConns(conf.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(conf.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(conf.Database.ConnMaxLifetimeSeconds) * time.Second)

	if err = sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("数据库连通性检查失败: %w", err)
	}
	return db, nil
}
