package storage

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/alanjm0532/agentmesh/internal/cli/config"
	"github.com/alanjm0532/agentmesh/internal/cli/ent"
	_ "modernc.org/sqlite"
)

func Open(ctx context.Context, configPath string, cfg config.DatabaseConfig) (*ent.Client, error) {
	absolute, err := filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("resolve CLI configuration path: %w", err)
	}
	databasePath := cfg.Path
	if !filepath.IsAbs(databasePath) {
		databasePath = filepath.Join(filepath.Dir(absolute), databasePath)
	}
	if err = os.MkdirAll(filepath.Dir(databasePath), 0o700); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(databasePath, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open CLI database file: %w", err)
	}
	if err = file.Chmod(0o600); err != nil {
		_ = file.Close()
		return nil, err
	}
	if err = file.Close(); err != nil {
		return nil, err
	}
	dsn := url.URL{Scheme: "file", Path: filepath.ToSlash(databasePath)}
	if runtime.GOOS == "windows" {
		dsn.Path = "/" + dsn.Path
	}
	parameters := url.Values{
		"_pragma":   {"foreign_keys(1)", "busy_timeout(10000)"},
		"_timezone": {"UTC"},
		// 写事务开始时获取锁，让读取、刷新和保存令牌成为一个串行操作。
		"_txlock": {"immediate"},
	}
	dsn.RawQuery = parameters.Encode()
	db, err := sql.Open("sqlite", dsn.String())
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	client := ent.NewClient(ent.Driver(entsql.OpenDB(dialect.SQLite, db)))
	// CLI 本地表统一交给 Ent 迁移，后续功能复用同一个数据库入口。
	if err = client.Schema.Create(ctx); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("migrate CLI database: %w", err)
	}
	return client, nil
}
