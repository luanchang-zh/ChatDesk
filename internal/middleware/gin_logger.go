package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/luanchang-zh/ChatDesk/pkg/logger"
	"go.uber.org/zap"
)

// GinLogger 在请求结束后打一条聚合访问日志。
//
// 级别规则：
//   - /health 且未 5xx：不打，避免探活刷屏；
//   - HTTP 5xx 或业务码 >= 30000：Error（服务端问题）；
//   - 耗时超过 2s：Warn（还成功，但慢）；
//   - 其余（含参数错误 1xxxx）：Info。参数输错是正常流量，不能记 Error。
//
// 必须放在 Recovery / Trace 之后：
//  1. Recovery 可能改写状态码；
//  2. Trace 已经把 trace_id 放进 context，这条日志才能对上信封。
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		// c.Next 跑完整个 handler 链；日志必须在这之后打，才能拿到最终状态码和业务码。
		c.Next()

		if IsPlatformPath(path) && c.Writer.Status() < 500 {
			return
		}

		ctx := NewContextWithGin(c)
		cost := time.Since(start)
		statusCode := c.Writer.Status()
		fields := []zap.Field{
			logger.String("method", c.Request.Method),
			logger.String("path", path),
			logger.String("ip", c.ClientIP()),
			logger.Int("status", statusCode),
			logger.Duration("cost", cost),
		}

		// result.Fail / FailServer 会把业务码塞进 gin.Context。
		// 拿不到就当 0：成功路径本来就没有 business_code。
		businessCode := 0
		if code, exists := c.Get("business_code"); exists {
			if bc, ok := code.(int); ok && bc > 0 {
				businessCode = bc
				fields = append(fields, logger.Int("business_code", bc))
			}
		}

		// FailServer 把 upstreamErr 挂进 c.Errors。只取第一条，避免一次请求打出一串重复错误。
		if len(c.Errors) > 0 {
			for _, ge := range c.Errors {
				if ge.Err == nil {
					continue
				}
				fields = append(fields, logger.ErrorField("error", ge.Err))
				break
			}
		}

		switch {
		case businessCode >= 30000, statusCode >= 500:
			logger.WithCtx(ctx).Error("HTTP 请求错误", fields...)
		case cost > 2*time.Second:
			logger.WithCtx(ctx).Warn("HTTP 慢请求", fields...)
		default:
			logger.WithCtx(ctx).Info("HTTP 请求成功", fields...)
		}
	}
}
