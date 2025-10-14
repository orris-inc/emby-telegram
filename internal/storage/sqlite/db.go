// Package sqlite 提供 SQLite 数据库存储实现
package sqlite

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"emby-telegram/internal/account"
	"emby-telegram/internal/user"
)

// Open 打开 SQLite 数据库连接
func Open(dsn string, debug bool) (*gorm.DB, error) {
	// 确保目录存在
	dir := filepath.Dir(dsn)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create database directory: %w", err)
		}
	}

	// 配置 logger
	var logLevel logger.LogLevel
	if debug {
		logLevel = logger.Info
	} else {
		logLevel = logger.Silent
	}

	// 打开数据库
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}

	// SQLite 推荐配置：单连接以避免并发问题
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

// AutoMigrate 自动迁移数据库表结构
func AutoMigrate(db *gorm.DB) error {
	// 迁移所有模型
	if err := db.AutoMigrate(
		&user.User{},
		&account.Account{},
	); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}

	return nil
}

// Close 关闭数据库连接
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sql.DB: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("close database: %w", err)
	}

	return nil
}
