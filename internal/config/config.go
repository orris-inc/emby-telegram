// Package config 提供应用配置管理
package config

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	App      AppConfig
	Telegram TelegramConfig
	Database DatabaseConfig
	Account  AccountConfig
	Emby     EmbyConfig
	Log      LogConfig
}

// AppConfig 应用配置
type AppConfig struct {
	Name    string
	Version string
	Debug   bool
}

// TelegramConfig Telegram Bot 配置
type TelegramConfig struct {
	Token    string
	Timeout  int
	AdminIDs []int64 `mapstructure:"admin_ids"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver string
	DSN    string
}

// AccountConfig 账号配置
type AccountConfig struct {
	DefaultExpireDays int    `mapstructure:"default_expire_days"`
	DefaultMaxDevices int    `mapstructure:"default_max_devices"`
	UsernamePrefix    string `mapstructure:"username_prefix"`
	PasswordLength    int    `mapstructure:"password_length"`
}

// EmbyConfig Emby 服务器配置
type EmbyConfig struct {
	ServerURL    string `mapstructure:"server_url"`
	APIKey       string `mapstructure:"api_key"`
	EnableSync   bool   `mapstructure:"enable_sync"`
	SyncOnCreate bool   `mapstructure:"sync_on_create"`
	SyncOnDelete bool   `mapstructure:"sync_on_delete"`
	Timeout      int    `mapstructure:"timeout"`
	RetryCount   int    `mapstructure:"retry_count"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string
	Output string
}

// Load 加载配置文件
// 优先级: 环境变量 > 配置文件 > 默认值
func Load() (*Config, error) {
	v := viper.New()

	// 设置默认值
	setDefaults(v)

	// 配置文件
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("/etc/emby-telegram")
	v.AddConfigPath("./configs")
	v.AddConfigPath(".")

	// 自动读取环境变量
	v.AutomaticEnv()

	// 读取配置文件 (如果不存在则使用默认值)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config file: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	// 环境变量覆盖
	if token := os.Getenv("TELEGRAM_BOT_TOKEN"); token != "" {
		cfg.Telegram.Token = token
	}

	if driver := os.Getenv("DB_DRIVER"); driver != "" {
		cfg.Database.Driver = driver
	}

	if dsn := os.Getenv("DB_DSN"); dsn != "" {
		cfg.Database.DSN = dsn
	}

	if serverURL := os.Getenv("EMBY_SERVER_URL"); serverURL != "" {
		cfg.Emby.ServerURL = serverURL
	}

	if apiKey := os.Getenv("EMBY_API_KEY"); apiKey != "" {
		cfg.Emby.APIKey = apiKey
	}

	// 验证必需配置
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return &cfg, nil
}

// setDefaults 设置默认配置值
func setDefaults(v *viper.Viper) {
	// App 默认值
	v.SetDefault("app.name", "Emby Telegram Bot")
	v.SetDefault("app.version", "1.0.0")
	v.SetDefault("app.debug", false)

	// Telegram 默认值
	v.SetDefault("telegram.timeout", 60)
	v.SetDefault("telegram.admin_ids", []int64{})

	// Database 默认值
	v.SetDefault("database.driver", "sqlite")
	v.SetDefault("database.dsn", "data/emby.db")

	// Account 默认值
	v.SetDefault("account.default_expire_days", 30)
	v.SetDefault("account.default_max_devices", 3)
	v.SetDefault("account.username_prefix", "emby_")
	v.SetDefault("account.password_length", 12)

	// Emby 默认值
	v.SetDefault("emby.server_url", "http://localhost:8096")
	v.SetDefault("emby.api_key", "")
	v.SetDefault("emby.enable_sync", true)
	v.SetDefault("emby.sync_on_create", true)
	v.SetDefault("emby.sync_on_delete", true)
	v.SetDefault("emby.timeout", 30)
	v.SetDefault("emby.retry_count", 3)

	// Log 默认值
	v.SetDefault("log.level", "info")
	v.SetDefault("log.output", "stdout")
}

// validate 验证配置
func (c *Config) validate() error {
	if c.Telegram.Token == "" {
		return fmt.Errorf("telegram.token is required")
	}

	if c.Telegram.Timeout <= 0 {
		c.Telegram.Timeout = 60
	}

	if c.Database.Driver == "" {
		return fmt.Errorf("database.driver is required")
	}

	if c.Database.DSN == "" {
		return fmt.Errorf("database.dsn is required")
	}

	if c.Account.DefaultExpireDays <= 0 {
		c.Account.DefaultExpireDays = 30
	}

	if c.Account.DefaultMaxDevices <= 0 {
		c.Account.DefaultMaxDevices = 3
	}

	if c.Account.PasswordLength < 8 {
		c.Account.PasswordLength = 12
	}

	// Emby 配置验证(仅在启用同步时)
	if c.Emby.EnableSync {
		if c.Emby.ServerURL == "" {
			return fmt.Errorf("emby.server_url is required when sync is enabled")
		}
		// API Key 可以为空，后续可以在运行时提示
		if c.Emby.Timeout <= 0 {
			c.Emby.Timeout = 30
		}
		if c.Emby.RetryCount < 0 {
			c.Emby.RetryCount = 0
		}
	}

	return nil
}

// GetTimeout 获取超时时间
func (c *TelegramConfig) GetTimeout() time.Duration {
	return time.Duration(c.Timeout) * time.Second
}

// IsAdmin 检查用户是否为管理员
func (c *TelegramConfig) IsAdmin(userID int64) bool {
	for _, id := range c.AdminIDs {
		if id == userID {
			return true
		}
	}
	return false
}
