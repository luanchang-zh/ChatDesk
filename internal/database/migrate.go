package database

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"io/fs"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrations embed.FS

// NewMigrator 创建 goose 迁移器。调用方必须 Close，否则会泄漏数据库连接。
// DSN 只从环境读入，失败时只返回类别信息，不回传连接串。
func NewMigrator(ctx context.Context, dsn string) (*goose.Provider, error) {
	if strings.TrimSpace(dsn) == "" {
		return nil, errors.New("POSTGRES_DSN 未配置")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, errors.New("PostgreSQL 配置无效")
	}
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, errors.New("PostgreSQL 连接失败，请检查数据库与 POSTGRES_DSN")
	}
	source, err := fs.Sub(migrations, "migrations")
	if err != nil {
		db.Close()
		return nil, err
	}
	provider, err := goose.NewProvider(goose.DialectPostgres, db, source)
	if err != nil {
		db.Close()
		return nil, errors.New("初始化数据库迁移失败")
	}
	return provider, nil
}
