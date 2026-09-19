package ctxmeta

import (
	"context"
	"strings"
)

// normalize 去掉首尾空白。空白 ID 和「没有 ID」在日志里同样有害，统一当成空。
func normalize(v string) string {
	return strings.TrimSpace(v)
}

// with 往标准 context 写一个字符串元数据。
//
// 规则：
//  1. ctx 为 nil 时改用 Background，避免后台 goroutine 忘传 context 直接 panic；
//  2. 空值不写：没登录不要留下 user_uuid=""，日志检索会把「匿名」和「空串用户」混在一起；
//  3. 键必须用本包常量，和 Gin / 日志 / 信封对齐。
func with(ctx context.Context, key, value string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	n := normalize(value)
	if n == "" {
		return ctx
	}
	return context.WithValue(ctx, key, n)
}

// get 从标准 context 读字符串。缺键、类型不对、空白，一律返回空串，调用方按「没有」处理。
func get(ctx context.Context, key string) string {
	if ctx == nil {
		return ""
	}
	value, ok := ctx.Value(key).(string)
	if !ok {
		return ""
	}
	return normalize(value)
}

// WithTraceID 把链路 ID 放进 context。Trace 中间件和 handle 转调 service 时都会用。
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return with(ctx, KeyTraceID, traceID)
}

// WithUserUUID 写入当前用户。鉴权中间件接上后再调；现在骨架里通常还是空的。
func WithUserUUID(ctx context.Context, userUUID string) context.Context {
	return with(ctx, KeyUserUUID, userUUID)
}

// WithClientIP 写入对端 IP。访问日志用；业务代码一般不要依赖它做鉴权。
func WithClientIP(ctx context.Context, clientIP string) context.Context {
	return with(ctx, KeyClientIP, clientIP)
}

// TraceID 取出链路 ID。logger.WithCtx 靠它把同一条请求的日志串起来。
func TraceID(ctx context.Context) string {
	return get(ctx, KeyTraceID)
}

// UserUUID 取出当前用户。没登录时返回空串，不要当成错误。
func UserUUID(ctx context.Context) string {
	return get(ctx, KeyUserUUID)
}

// ClientIP 取出对端 IP。
func ClientIP(ctx context.Context) string {
	return get(ctx, KeyClientIP)
}
