package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/luanchang-zh/ChatDesk/pkg/logger"
	"go.uber.org/zap"
)

// GinLogger 请求结束后打一条访问日志。
// /health 成功不打，避免探活刷屏；5xx 或业务码 >= 30000 走 Error。
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
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

		businessCode := 0
		if code, exists := c.Get("business_code"); exists {
			if bc, ok := code.(int); ok && bc > 0 {
				businessCode = bc
				fields = append(fields, logger.Int("business_code", bc))
			}
		}

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
