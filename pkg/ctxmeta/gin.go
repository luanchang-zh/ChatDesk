package ctxmeta

import (
	"context"

	"github.com/gin-gonic/gin"
)

func setGinString(c *gin.Context, key, value string) string {
	if c == nil {
		return ""
	}
	n := normalize(value)
	if n == "" {
		return ""
	}
	c.Set(key, n)
	return n
}

func getGinString(c *gin.Context, key string) string {
	if c == nil {
		return ""
	}
	value, exists := c.Get(key)
	if !exists {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return normalize(text)
}

func SetTraceID(c *gin.Context, traceID string) string {
	return setGinString(c, KeyTraceID, traceID)
}

func SetUserUUID(c *gin.Context, userUUID string) string {
	return setGinString(c, KeyUserUUID, userUUID)
}

func SetClientIP(c *gin.Context, clientIP string) string {
	return setGinString(c, KeyClientIP, clientIP)
}

func TraceIDFromGin(c *gin.Context) string {
	return getGinString(c, KeyTraceID)
}

func UserUUIDFromGin(c *gin.Context) string {
	return getGinString(c, KeyUserUUID)
}

func ClientIPFromGin(c *gin.Context) string {
	return getGinString(c, KeyClientIP)
}

// BuildContextFromGin 把 Gin 里的 trace / 用户信息拷进标准 context，供 logger.WithCtx 使用。
func BuildContextFromGin(c *gin.Context) context.Context {
	if c == nil || c.Request == nil {
		return context.Background()
	}
	ctx := c.Request.Context()
	if traceID := TraceIDFromGin(c); traceID != "" {
		ctx = WithTraceID(ctx, traceID)
	}
	if userUUID := UserUUIDFromGin(c); userUUID != "" {
		ctx = WithUserUUID(ctx, userUUID)
	}
	if clientIP := ClientIPFromGin(c); clientIP != "" {
		ctx = WithClientIP(ctx, clientIP)
	}
	return ctx
}
