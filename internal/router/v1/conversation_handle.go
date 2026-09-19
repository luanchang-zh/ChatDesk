package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/luanchang-zh/ChatDesk/consts"
	"github.com/luanchang-zh/ChatDesk/internal/domain"
	"github.com/luanchang-zh/ChatDesk/internal/middleware"
	"github.com/luanchang-zh/ChatDesk/internal/service"
	"github.com/luanchang-zh/ChatDesk/pkg/result"
)

// ConversationHandler 会话 HTTP 入口。
// 只做绑定、调 service、映射 DTO；不直接碰数据库或 Python 进程。
type ConversationHandler struct {
	conversations service.ConversationService
}

// NewConversationHandler 注入会话服务。调用方负责传入占位实现或真实实现。
func NewConversationHandler(conversations service.ConversationService) *ConversationHandler {
	return &ConversationHandler{conversations: conversations}
}

// conversationItem 给前端的会话摘要。
// 时间用 RFC3339 字符串，避免把 time.Time 零值直接序列化成 "0001-01-01T00:00:00Z"。
type conversationItem struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Title     string `json:"title"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// List 列出会话。
// 当前 service 仍是占位，会返回 CodeServiceUnavailable；接上存储后这个函数不用改签名。
// @Router GET /api/v1/conversations
func (h *ConversationHandler) List(c *gin.Context) {
	// 1. 绑定查询参数。用 ShouldBindQuery，失败时由我们自己写信封，而不是让 Gin 直接 400。
	var req ListConversationsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		// 参数错误由客户端输入导致，属于正常业务流程，不记录 Error 日志。
		result.Fail(c, nil, consts.CodeParamError)
		return
	}

	// 2. 从 Gin 抽出带 trace_id 的标准 context。service 禁止接收 *gin.Context。
	ctx := middleware.NewContextWithGin(c)

	// 3. 调用领域服务。UserID 等鉴权接上后再从 ctx 填入，现在先空着。
	items, err := h.conversations.List(ctx, domain.ListConversationsInput{
		ProjectID: req.ProjectID,
	})
	if err != nil {
		handleServiceError(c, err)
		return
	}

	// 4. 领域对象 → HTTP DTO。domain 没有 json tag，避免把内部字段泄漏成线格式。
	out := make([]conversationItem, 0, len(items))
	for _, item := range items {
		out = append(out, toConversationItem(item))
	}
	result.Success(c, gin.H{"items": out})
}

// toConversationItem 时间统一转 UTC RFC3339，前后端时区才不会各说各话。
func toConversationItem(item domain.Conversation) conversationItem {
	return conversationItem{
		ID:        item.ID,
		ProjectID: item.ProjectID,
		Title:     item.Title,
		CreatedAt: item.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: item.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}
}
