// Package main Emby Telegram Bot 主入口
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"emby-telegram/internal/account"
	"emby-telegram/internal/bot"
	"emby-telegram/internal/config"
	"emby-telegram/internal/emby"
	"emby-telegram/internal/logger"
	"emby-telegram/internal/storage/sqlite"
	"emby-telegram/internal/user"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// 初始化日志
	if err := logger.Init(cfg.Log.Level, cfg.Log.Output); err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("===== emby telegram bot starting =====")
	logger.Infof("version: %s", cfg.App.Version)
	logger.Infof("debug mode: %v", cfg.App.Debug)

	// 初始化数据库
	db, err := sqlite.Open(cfg.Database.DSN, cfg.App.Debug)
	if err != nil {
		logger.Fatalf("failed to initialize database: %v", err)
	}
	logger.Info("✓ database connected")

	// 自动迁移
	if err := sqlite.AutoMigrate(db); err != nil {
		logger.Fatalf("failed to migrate database: %v", err)
	}
	logger.Info("✓ database migrated")

	// 创建存储层
	userStore := sqlite.NewUserStore(db)
	accountStore := sqlite.NewAccountStore(db)

	// 初始化 Emby Client
	var embyClient *emby.Client
	if cfg.Emby.EnableSync && cfg.Emby.ServerURL != "" && cfg.Emby.APIKey != "" {
		embyClient = emby.NewClient(
			cfg.Emby.ServerURL,
			cfg.Emby.APIKey,
			cfg.Emby.Timeout,
			cfg.Emby.RetryCount,
			cfg.Emby.EnableSync,
		)

		// 测试连接
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := embyClient.Ping(ctx); err != nil {
			logger.Warnf("emby server connection failed, running in offline mode: %v", err)
			embyClient = nil // 降级为离线模式
		} else {
			logger.Info("✓ emby server connected")
		}
	} else {
		logger.Info("✓ emby sync disabled, running in offline mode")
	}

	// 创建服务层
	userService := user.NewService(userStore)
	accountService := account.NewService(
		accountStore,
		embyClient,
		cfg.Account.UsernamePrefix,
		cfg.Account.DefaultExpireDays,
		cfg.Account.DefaultMaxDevices,
		cfg.Account.PasswordLength,
		cfg.Emby.EnableSync,
		cfg.Emby.SyncOnCreate,
		cfg.Emby.SyncOnDelete,
	)

	// 创建并启动 Telegram Bot
	telegramBot, err := bot.New(
		cfg.Telegram.Token,
		cfg.Telegram.AdminIDs,
		accountService,
		userService,
		embyClient,
	)
	if err != nil {
		logger.Fatalf("failed to initialize bot: %v", err)
	}
	logger.Info("✓ telegram bot initialized")

	// 创建上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动 Bot (在 goroutine 中)
	go func() {
		if err := telegramBot.Start(ctx); err != nil {
			if err != context.Canceled {
				logger.Errorf("bot runtime error: %v", err)
			}
			cancel()
		}
	}()

	logger.Info("===== bot ready =====")

	// 优雅关闭信号处理
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		logger.Infof("received shutdown signal: %v", sig)
	case <-ctx.Done():
		logger.Info("context canceled")
	}

	// 优雅关闭
	logger.Info("shutting down bot...")
	telegramBot.Stop()

	// 关闭数据库连接
	if err := sqlite.Close(db); err != nil {
		logger.Errorf("failed to close database connection: %v", err)
	} else {
		logger.Info("✓ database connection closed")
	}

	logger.Info("===== bot stopped =====")
}
