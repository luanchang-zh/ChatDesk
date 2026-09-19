package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/luanchang-zh/ChatDesk/pkg/ctxmeta"
)

// Trace 为每个请求准备 trace_id。
// 后续 logger.WithCtx、result.Response.TraceId、响应头 X-Request-ID 都靠它对齐。
// 必须挂在 Logger 和业务 handle 之前，否则它们拿到的 context 是空的。
func Trace() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 优先用上游（浏览器、反向代理）带来的 ID，便于整条请求对账；没有再生成。
		traceID := c.GetHeader(ctxmeta.HeaderRequestID)
		if traceID == "" {
			traceID = uuid.NewString()
		}

		// 2. 放入 Gin 上下文。result.Result 用 c.GetString("trace_id") 填信封。
		traceID = ctxmeta.SetTraceID(c, traceID)

		// 3. 放入标准 request.Context，service / logger 才能拿到，而不必传 *gin.Context。
		reqCtx := ctxmeta.WithTraceID(c.Request.Context(), traceID)
		c.Request = c.Request.WithContext(reqCtx)

		// 4. 回写响应头，前端或 curl -D - 可以直接对着日志搜。
		c.Header(ctxmeta.HeaderRequestID, traceID)

		c.Next()
	}
}
