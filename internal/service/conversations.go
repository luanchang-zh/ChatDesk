package service

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/luanchang-zh/ChatDesk/consts"
	"github.com/luanchang-zh/ChatDesk/internal/domain"
	"github.com/luanchang-zh/ChatDesk/pkg/apperr"
	"github.com/luanchang-zh/ChatDesk/pkg/ctxmeta"
)

// ConversationStore 会话仓储。
// 每一次读写都必须带当前用户：按 ID 查询也不能省略归属，否则会把别人的会话漏出来。
type ConversationStore interface {
	Create(context.Context, uuid.UUID, string) (*domain.Conversation, error)
	Get(context.Context, uuid.UUID, uuid.UUID) (*domain.Conversation, error)
	List(context.Context, uuid.UUID) ([]domain.Conversation, error)
	Rename(context.Context, uuid.UUID, uuid.UUID, string) (*domain.Conversation, error)
	Delete(context.Context, uuid.UUID, uuid.UUID) error
}

// Conversations 会话领域服务。身份只从 Context 读取，不信任 Input.UserID。
type Conversations struct {
	store ConversationStore
}

func NewConversations(store ConversationStore) *Conversations {
	return &Conversations{store: store}
}

// Create 新建会话。空白标题改成「新会话」；非空 project_id 本轮明确拒绝。
func (s *Conversations) Create(ctx context.Context, in domain.CreateConversationInput) (*domain.Conversation, error) {
	userID, err := actorID(ctx)
	if err != nil {
		return nil, err
	}
	// 项目管理还没做：不能静默丢掉 project_id，否则以后接上时历史数据对不齐。
	if in.ProjectID != "" {
		return nil, apperr.New(consts.CodeParamError)
	}
	title := strings.TrimSpace(in.Title)
	if title == "" {
		title = "新会话"
	}
	if !validTitle(title) {
		return nil, apperr.New(consts.CodeParamError)
	}
	return s.store.Create(ctx, userID, title)
}

// Get 打开当前用户的会话。他人 ID 与不存在走同一套错误，由仓储保证。
func (s *Conversations) Get(ctx context.Context, id string) (*domain.Conversation, error) {
	userID, conversationID, err := conversationIDs(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.store.Get(ctx, userID, conversationID)
}

// List 列出当前用户会话。空结果应是空切片，不是 nil。
func (s *Conversations) List(ctx context.Context, in domain.ListConversationsInput) ([]domain.Conversation, error) {
	userID, err := actorID(ctx)
	if err != nil {
		return nil, err
	}
	if in.ProjectID != "" {
		return nil, apperr.New(consts.CodeParamError)
	}
	return s.store.List(ctx, userID)
}

// Rename 修改标题。去掉首尾空白后不能为空；归属不能改。
func (s *Conversations) Rename(ctx context.Context, in domain.RenameConversationInput) (*domain.Conversation, error) {
	userID, id, err := conversationIDs(ctx, in.ID)
	if err != nil {
		return nil, err
	}
	title := strings.TrimSpace(in.Title)
	if !validTitle(title) {
		return nil, apperr.New(consts.CodeParamError)
	}
	return s.store.Rename(ctx, userID, id, title)
}

// Delete 删除当前用户会话。不存在时返回资源不存在。
func (s *Conversations) Delete(ctx context.Context, id string) error {
	userID, conversationID, err := conversationIDs(ctx, id)
	if err != nil {
		return err
	}
	return s.store.Delete(ctx, userID, conversationID)
}

// actorID 从可信 Context 取当前用户。空或非法一律未认证，不读请求体里的 user_id。
func actorID(ctx context.Context) (uuid.UUID, error) {
	id, err := uuid.Parse(ctxmeta.UserUUID(ctx))
	if err != nil || id == uuid.Nil {
		return uuid.Nil, apperr.New(consts.CodeUnauthorized)
	}
	return id, nil
}

// conversationIDs 同时校验当前用户和路径里的会话 ID。
// 会话 ID 非法是参数错误；没有身份是未认证。两者不能混。
func conversationIDs(ctx context.Context, rawID string) (uuid.UUID, uuid.UUID, error) {
	userID, err := actorID(ctx)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	id, err := uuid.Parse(rawID)
	if err != nil || id == uuid.Nil {
		return uuid.Nil, uuid.Nil, apperr.New(consts.CodeParamError)
	}
	return userID, id, nil
}

// validTitle 标题去掉空白后必须是 1～200 个 Unicode 字符，且不含 NUL。
func validTitle(title string) bool {
	return utf8.ValidString(title) && strings.IndexByte(title, 0) < 0 && utf8.RuneCountInString(title) >= 1 && utf8.RuneCountInString(title) <= 200
}
