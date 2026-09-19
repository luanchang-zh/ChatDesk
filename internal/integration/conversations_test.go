package integration

import (
	"context"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/luanchang-zh/ChatDesk/consts"
	"github.com/luanchang-zh/ChatDesk/internal/database"
	"github.com/luanchang-zh/ChatDesk/internal/domain"
	"github.com/luanchang-zh/ChatDesk/internal/repository"
	"github.com/luanchang-zh/ChatDesk/internal/service"
	"github.com/luanchang-zh/ChatDesk/pkg/apperr"
	"github.com/luanchang-zh/ChatDesk/pkg/ctxmeta"
)

func TestPostgresConversationPersistence(t *testing.T) {
	adminDSN := os.Getenv("CHATDESK_TEST_ADMIN_DSN")
	if adminDSN == "" {
		t.Skip("CHATDESK_TEST_ADMIN_DSN 未配置；真实 PostgreSQL 测试未执行，不启动 Docker")
	}
	u, err := url.Parse(adminDSN)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") {
		t.Fatal("CHATDESK_TEST_ADMIN_DSN 必须是 PostgreSQL URL")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	admin, err := pgx.Connect(ctx, adminDSN)
	if err != nil {
		t.Fatal("连接 PostgreSQL 测试管理员失败")
	}
	t.Cleanup(func() { _ = admin.Close(context.Background()) })
	dbName := "chatdesk_test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	identifier := pgx.Identifier{dbName}.Sanitize()
	if _, err := admin.Exec(ctx, "CREATE DATABASE "+identifier); err != nil {
		t.Fatal("创建隔离测试数据库失败，需要 CREATEDB 权限")
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()
		if _, err := admin.Exec(cleanupCtx, "DROP DATABASE "+identifier); err != nil {
			t.Errorf("清理本次创建的测试数据库 %s 失败", dbName)
		}
	})
	u.Path = "/" + dbName
	query := u.Query()
	query.Del("dbname")
	query.Del("database")
	u.RawQuery = query.Encode()
	dsn := u.String()
	probe, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal("连接隔离测试数据库失败")
	}
	var actualDB string
	err = probe.QueryRow(ctx, "SELECT current_database()").Scan(&actualDB)
	_ = probe.Close(ctx)
	if err != nil || actualDB != dbName {
		t.Fatal("拒绝在非本次创建的数据库上执行迁移测试")
	}
	provider, err := database.NewMigrator(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = provider.Close() })
	if _, err := provider.Up(ctx); err != nil {
		t.Fatal(err)
	}
	if applied, err := provider.Up(ctx); err != nil || len(applied) != 0 {
		t.Fatalf("second up: count=%d error=%v", len(applied), err)
	}
	if _, err := provider.Down(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Up(ctx); err != nil {
		t.Fatal(err)
	}
	pool, err := database.Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	store := repository.NewConversations(pool)
	a, b := uuid.New(), uuid.New()
	for _, owner := range []uuid.UUID{a, a, b} {
		if err := store.EnsureUser(ctx, owner); err != nil {
			t.Fatal(err)
		}
	}
	ctxA, ctxB := ctxmeta.WithUserUUID(ctx, a.String()), ctxmeta.WithUserUUID(ctx, b.String())
	svc := service.NewConversations(store)
	created, err := svc.Create(ctxA, domain.CreateConversationInput{Title: " 原标题 "})
	if err != nil {
		t.Fatal(err)
	}
	if created.Title != "原标题" || created.UserID != a.String() {
		t.Fatalf("unexpected persisted conversation: %+v", created)
	}
	if _, err := svc.Create(ctxB, domain.CreateConversationInput{}); err != nil {
		t.Fatal(err)
	}
	items, err := svc.List(ctxA, domain.ListConversationsInput{})
	if err != nil || len(items) != 1 || items[0].ID != created.ID {
		t.Fatalf("scoped list: %#v %v", items, err)
	}
	if _, err := svc.Get(ctxB, created.ID); apperr.Code(err) != consts.CodeResourceNotFound {
		t.Fatalf("cross-owner get: %v", err)
	}
	if _, err := svc.Rename(ctxB, domain.RenameConversationInput{ID: created.ID, Title: "攻击"}); apperr.Code(err) != consts.CodeResourceNotFound {
		t.Fatalf("cross-owner rename: %v", err)
	}
	if err := svc.Delete(ctxB, created.ID); apperr.Code(err) != consts.CodeResourceNotFound {
		t.Fatalf("cross-owner delete: %v", err)
	}
	unchanged, err := svc.Get(ctxA, created.ID)
	if err != nil || unchanged.Title != "原标题" {
		t.Fatal("cross-owner write changed owner data")
	}
	if _, err := svc.Rename(ctxA, domain.RenameConversationInput{ID: created.ID, Title: " 新标题 "}); err != nil {
		t.Fatal(err)
	}
	pool.Close()
	reopened, err := database.Open(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(reopened.Close)
	svc = service.NewConversations(repository.NewConversations(reopened))
	restored, err := svc.Get(ctxA, created.ID)
	if err != nil || restored.Title != "新标题" || restored.CreatedAt.IsZero() {
		t.Fatalf("persistence after reopen: %#v %v", restored, err)
	}
	if err := svc.Delete(ctxA, created.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Get(ctxA, created.ID); apperr.Code(err) != consts.CodeResourceNotFound {
		t.Fatalf("deleted conversation: %v", err)
	}
	items, err = svc.List(ctxA, domain.ListConversationsInput{})
	if err != nil || items == nil || len(items) != 0 {
		t.Fatalf("empty list: %#v %v", items, err)
	}
}
