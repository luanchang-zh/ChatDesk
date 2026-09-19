package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/luanchang-zh/ChatDesk/consts"
	"github.com/luanchang-zh/ChatDesk/internal/domain"
	"github.com/luanchang-zh/ChatDesk/internal/repository/dbgen"
	"github.com/luanchang-zh/ChatDesk/pkg/apperr"
)

// Conversations 用 sqlc 生成的查询访问 PostgreSQL。
// 本层只做行映射和错误转换，不解释标题规则或项目关联。
type Conversations struct {
	queries *dbgen.Queries
}

func NewConversations(pool *pgxpool.Pool) *Conversations {
	return &Conversations{queries: dbgen.New(pool)}
}

// EnsureUser 幂等写入 users 行。开发身份启动时调用，重复执行不能报错。
func (r *Conversations) EnsureUser(ctx context.Context, id uuid.UUID) error {
	return r.queries.EnsureUser(ctx, id)
}

// Create 由服务端生成会话 UUID。user_id 必须是调用方已经校验过的当前用户。
func (r *Conversations) Create(ctx context.Context, userID uuid.UUID, title string) (*domain.Conversation, error) {
	row, err := r.queries.CreateConversation(ctx, dbgen.CreateConversationParams{
		ID: uuid.New(), UserID: userID, Title: title,
	})
	if err != nil {
		return nil, conversationError(err)
	}
	return conversation(row), nil
}

// Get 同时按会话 ID 和 user_id 查询。找不到与属于别人都是 pgx.ErrNoRows。
func (r *Conversations) Get(ctx context.Context, userID, id uuid.UUID) (*domain.Conversation, error) {
	row, err := r.queries.GetConversation(ctx, dbgen.GetConversationParams{ID: id, UserID: userID})
	if err != nil {
		return nil, conversationError(err)
	}
	return conversation(row), nil
}

// List 只返回该用户的会话。make(..., 0) 保证空列表不是 nil，JSON 才能编成 []。
func (r *Conversations) List(ctx context.Context, userID uuid.UUID) ([]domain.Conversation, error) {
	rows, err := r.queries.ListConversations(ctx, userID)
	if err != nil {
		return nil, conversationError(err)
	}
	items := make([]domain.Conversation, 0, len(rows))
	for _, row := range rows {
		items = append(items, *conversation(row))
	}
	return items, nil
}

// Rename 条件更新标题。影响行数为 0 时 sqlc 会返回 ErrNoRows。
func (r *Conversations) Rename(ctx context.Context, userID, id uuid.UUID, title string) (*domain.Conversation, error) {
	row, err := r.queries.RenameConversation(ctx, dbgen.RenameConversationParams{
		ID: id, UserID: userID, Title: title,
	})
	if err != nil {
		return nil, conversationError(err)
	}
	return conversation(row), nil
}

// Delete 按 ID 和 user_id 删除。count=0 与 Get 一样当成资源不存在。
func (r *Conversations) Delete(ctx context.Context, userID, id uuid.UUID) error {
	count, err := r.queries.DeleteConversation(ctx, dbgen.DeleteConversationParams{ID: id, UserID: userID})
	if err != nil {
		return conversationError(err)
	}
	if count == 0 {
		return apperr.New(consts.CodeResourceNotFound)
	}
	return nil
}

// conversation 把 sqlc 行收成领域对象。不带 json tag，HTTP 层自己做 DTO。
func conversation(row dbgen.Conversation) *domain.Conversation {
	return &domain.Conversation{
		ID: row.ID.String(), UserID: row.UserID.String(), Title: row.Title,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

// conversationError 把驱动错误收成业务码。
// 原文只进 Wrap 的 cause，信封 message 走 consts，避免把 SQL 泄漏给客户端。
func conversationError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return apperr.New(consts.CodeResourceNotFound)
	}
	return apperr.Wrap(err, consts.CodeInternalError, "")
}
