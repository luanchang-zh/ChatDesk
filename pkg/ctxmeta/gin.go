package ctxmeta

import (
	"context"

	"github.com/gin-gonic/gin"
)

// setGinString 往 gin.Context 写字符串。
// c 为 nil 或值为空时不写：中间件单元测试和错误路径经常拿到空指针，这里要扛住。
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

// getGinString 从 gin.Context 读字符串。键不存在或类型不是 string 都当空。
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

// SetTraceID 把链路 ID 放进 Gin。result.Result 用 c.GetString("trace_id") 填信封，键必须是 KeyTraceID。
func SetTraceID(c *gin.Context, traceID string) string {
	return setGinString(c, KeyTraceID, traceID)
}

// SetUserUUID 把当前用户放进 Gin。开发身份中间件写入后再拷到标准 context。
func SetUserUUID(c *gin.Context, userUUID string) string {
	return setGinString(c, KeyUserUUID, userUUID)
}

// SetClientIP 把对端 IP 放进 Gin。
func SetClientIP(c *gin.Context, clientIP string) string {
	return setGinString(c, KeyClientIP, clientIP)
}

// TraceIDFromGin 从 Gin 取出链路 ID。BuildContextFromGin 会再拷到标准 context。
func TraceIDFromGin(c *gin.Context) string {
	return getGinString(c, KeyTraceID)
}

// UserUUIDFromGin 从 Gin 取出当前用户。
func UserUUIDFromGin(c *gin.Context) string {
	return getGinString(c, KeyUserUUID)
}

// ClientIPFromGin 从 Gin 取出对端 IP。
func ClientIPFromGin(c *gin.Context) string {
	return getGinString(c, KeyClientIP)
}

// BuildContextFromGin 把 Gin 里已经放好的链路字段拷进标准 context。
//
// 必须先跑 Trace 中间件，否则 gin.Context 上还没有 trace_id，拷过去也是空的。
//
// 返回值继承 c.Request.Context() 的取消信号：客户端断开、Shutdown 超时时，
// 下游 service 读 ctx.Done() 才能停，而不是一直打到 Postgres。
//
// handle 禁止把 *gin.Context 传给 service，统一走这条转换。
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
