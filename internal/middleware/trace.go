package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/luanchang-zh/ChatDesk/pkg/ctxmeta"
)

// Trace 生成或透传 X-Request-ID，写入 Gin 与 request.Context，并回写响应头。
func Trace() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 优先用上游（网关 / 浏览器）带来的 ID
		traceID := c.GetHeader(ctxmeta.HeaderRequestID)
		if traceID == "" {
			traceID = uuid.NewString()
		}

		// 2. 放入 Gin 上下文，result.Response.TraceId 会读这个 key
		traceID = ctxmeta.SetTraceID(c, traceID)

		// 3. 放入标准 context，供 logger.WithCtx 自动带上 trace_id
		reqCtx := ctxmeta.WithTraceID(c.Request.Context(), traceID)
		c.Request = c.Request.WithContext(reqCtx)

		// 4. 回写响应头，方便对照日志
		c.Header(ctxmeta.HeaderRequestID, traceID)

		c.Next()
	}
}
