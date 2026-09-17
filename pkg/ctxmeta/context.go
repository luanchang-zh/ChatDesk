package ctxmeta

import (
	"context"
	"strings"
)

func normalize(v string) string {
	return strings.TrimSpace(v)
}

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

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return with(ctx, KeyTraceID, traceID)
}

func WithUserUUID(ctx context.Context, userUUID string) context.Context {
	return with(ctx, KeyUserUUID, userUUID)
}

func WithClientIP(ctx context.Context, clientIP string) context.Context {
	return with(ctx, KeyClientIP, clientIP)
}

func TraceID(ctx context.Context) string {
	return get(ctx, KeyTraceID)
}

func UserUUID(ctx context.Context) string {
	return get(ctx, KeyUserUUID)
}

func ClientIP(ctx context.Context) string {
	return get(ctx, KeyClientIP)
}
