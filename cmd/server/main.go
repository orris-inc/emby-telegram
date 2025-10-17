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
	"emby-telegram/internal/storage"
	"emby-telegram/internal/user"
)

// userGetterAdapter adapts user.Service to account.UserGetter interface
type userGetterAdapter struct {
	userService *user.Service
}

func (a *userGetterAdapter) Get(ctx context.Context, id uint) (account.User, error) {
	u, err := a.userService.Get(ctx, id)
	if err != nil {
		return account.User{}, err
	}
	return account.User{
		ID:           u.ID,
		IsAdmin:      u.IsAdmin(),
		AccountQuota: u.AccountQuota,
	}, nil
}

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

	stores, err := storage.NewStores(cfg.Database.Driver, cfg.Database.DSN, cfg.App.Debug)
	if err != nil {
		logger.Fatalf("failed to initialize database: %v", err)
	}
	logger.Infof("✓ database connected (driver: %s)", cfg.Database.Driver)
	logger.Info("✓ database migrated")

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

	userService := user.NewService(stores.UserStore)

	// 创建 UserGetter 适配器
	userGetter := &userGetterAdapter{userService: userService}

	accountService := account.NewService(
		stores.AccountStore,
		userGetter,
		embyClient,
		cfg.Account.UsernamePrefix,
		cfg.Account.DefaultExpireDays,
		cfg.Account.DefaultMaxDevices,
		cfg.Account.PasswordLength,
		cfg.Account.MaxAccountsPerUser,
		cfg.Account.MaxAccountsPerAdmin,
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

	logger.Info("shutting down bot...")
	telegramBot.Stop()

	if err := stores.Close(); err != nil {
		logger.Errorf("failed to close database connection: %v", err)
	} else {
		logger.Info("✓ database connection closed")
	}

	logger.Info("===== bot stopped =====")
}
