package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/luanchang-zh/ChatDesk/consts"
	"github.com/luanchang-zh/ChatDesk/internal/middleware"
	"github.com/luanchang-zh/ChatDesk/pkg/apperr"
	"github.com/luanchang-zh/ChatDesk/pkg/result"
)

// ConversationHandler 会话 HTTP 入口。本轮只定 handle 骨架，不接真实存储。
type ConversationHandler struct{}

func NewConversationHandler() *ConversationHandler {
	return &ConversationHandler{}
}

// List 列出会话
// @Router GET /api/v1/conversations
func (h *ConversationHandler) List(c *gin.Context) {
	// 1. 绑定查询参数
	var req ListConversationsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		// 参数错误由客户端输入导致，属于正常业务流程，不记录日志
		result.Fail(c, nil, consts.CodeParamError)
		return
	}

	// 2. 抽出带 trace_id 的 context。下一轮会传给 ConversationService.List(ctx, req)。
	ctx := middleware.NewContextWithGin(c)
	_ = ctx
	_ = req

	// 3. 本轮尚未接入会话存储。直接 FailServer，保留「服务暂不可用」业务码；
	//    未知错误才走 handleServiceError 收成 CodeInternalError。
	err := apperr.New(consts.CodeServiceUnavailable)
	result.FailServer(c, err, consts.CodeServiceUnavailable)
}
