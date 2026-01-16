package mysql

import (
	"fmt"
	"time"

	"github.com/wavespeedai/waverless-portal/pkg/config"
	"github.com/wavespeedai/waverless-portal/pkg/logger"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type Repository struct {
	DB *gorm.DB
}

func NewRepository() (*Repository, error) {
	cfg := config.GlobalConfig.MySQL

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=UTC",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database)

	logLevel := gormlogger.Silent
	if config.GlobalConfig.Server.Mode == "debug" {
		logLevel = gormlogger.Info
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	logger.Infof("MySQL connected: %s:%d/%s", cfg.Host, cfg.Port, cfg.Database)

	return &Repository{DB: db}, nil
}

func (r *Repository) Close() error {
	sqlDB, err := r.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
