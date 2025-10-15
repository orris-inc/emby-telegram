package storage

import (
	"fmt"

	"gorm.io/gorm"

	"emby-telegram/internal/account"
	"emby-telegram/internal/storage/mysql"
	"emby-telegram/internal/storage/sqlite"
	"emby-telegram/internal/user"
)

type Stores struct {
	UserStore    user.Store
	AccountStore account.Store
	DB           *gorm.DB
}

func NewStores(driver, dsn string, debug bool) (*Stores, error) {
	var db *gorm.DB
	var err error

	switch driver {
	case "sqlite":
		db, err = sqlite.Open(dsn, debug)
		if err != nil {
			return nil, fmt.Errorf("open sqlite database: %w", err)
		}
		if err := sqlite.AutoMigrate(db); err != nil {
			return nil, fmt.Errorf("migrate sqlite database: %w", err)
		}
		return &Stores{
			UserStore:    sqlite.NewUserStore(db),
			AccountStore: sqlite.NewAccountStore(db),
			DB:           db,
		}, nil

	case "mysql":
		db, err = mysql.Open(dsn, debug)
		if err != nil {
			return nil, fmt.Errorf("open mysql database: %w", err)
		}
		if err := mysql.AutoMigrate(db); err != nil {
			return nil, fmt.Errorf("migrate mysql database: %w", err)
		}
		return &Stores{
			UserStore:    mysql.NewUserStore(db),
			AccountStore: mysql.NewAccountStore(db),
			DB:           db,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported database driver: %s", driver)
	}
}

func (s *Stores) Close() error {
	sqlDB, err := s.DB.DB()
	if err != nil {
		return fmt.Errorf("get sql.DB: %w", err)
	}
	return sqlDB.Close()
}
