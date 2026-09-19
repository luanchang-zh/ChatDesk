package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Open 建立连接池并检查 users / conversations 是否已迁移。
// 这里不执行 goose：缺表时明确失败，避免 HTTP 启动悄悄改库结构。
func Open(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, errors.New("PostgreSQL 配置无效")
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, errors.New("PostgreSQL 连接失败，请检查数据库与 POSTGRES_DSN")
	}
	// LIMIT 0 只验证列和表存在，不读业务行。错误文案不带 DSN。
	rows, err := pool.Query(ctx, `SELECT u.id, u.created_at, c.id, c.user_id, c.title, c.created_at, c.updated_at
		FROM users u LEFT JOIN conversations c ON c.user_id = u.id LIMIT 0`)
	if err != nil {
		pool.Close()
		return nil, errors.New("会话数据库结构未就绪，请先执行迁移")
	}
	rows.Close()
	if rows.Err() != nil {
		pool.Close()
		return nil, errors.New("检查会话数据库结构失败")
	}
	return pool, nil
}
