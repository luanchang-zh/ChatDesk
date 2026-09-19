package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/luanchang-zh/ChatDesk/pkg/ctxmeta"
)

// NewContextWithGin 从 gin.Context 抽出带 trace_id / 用户信息的标准 context。
//
// handle 统一写法：
//
//	ctx := middleware.NewContextWithGin(c)
//	items, err := h.svc.List(ctx, ...)
//	logger.WithCtx(ctx).Error("...", ...)
//
// 禁止把 *gin.Context 传给 service：service 不能绑定 HTTP、不能写信封。
// 必须先跑 Trace 中间件，否则 gin.Context 上还没有 trace_id。
func NewContextWithGin(c *gin.Context) context.Context {
	return ctxmeta.BuildContextFromGin(c)
}

// IsPlatformPath 探活、以后的 metrics 走这里。
// GinLogger 对成功的平台路径直接跳过，避免编排探活把访问日志刷没。
// 不要把 /api/v1 下的业务接口加进来。
func IsPlatformPath(path string) bool {
	return path == "/health"
}
