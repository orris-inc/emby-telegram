package database

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	"github.com/pressly/goose/v3"
)

func RunMigrations(db *sql.DB, driver, migrationsDir string, logger *slog.Logger) error {
	var dialect goose.Dialect
	var subDir string

	if driver == "sqlite" || driver == "sqlite3" {
		dialect = goose.DialectSQLite3
		subDir = "sqlite"
	} else if driver == "mysql" {
		dialect = goose.DialectMySQL
		subDir = "mysql"
	} else {
		return fmt.Errorf("unsupported driver: %s", driver)
	}

	fullPath := fmt.Sprintf("%s/%s", migrationsDir, subDir)
	fsys := os.DirFS(fullPath)

	provider, err := goose.NewProvider(dialect, db, fsys, goose.WithSlog(logger))
	if err != nil {
		return fmt.Errorf("create goose provider: %w", err)
	}

	_, err = provider.Up(context.Background())
	if err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	return nil
}
