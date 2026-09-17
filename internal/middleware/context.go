package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/luanchang-zh/ChatDesk/pkg/ctxmeta"
)

// NewContextWithGin 从 gin.Context 创建带 trace_id 的 context.Context。
// handler 里统一：ctx := middleware.NewContextWithGin(c)
func NewContextWithGin(c *gin.Context) context.Context {
	return ctxmeta.BuildContextFromGin(c)
}

func IsPlatformPath(path string) bool {
	return path == "/health"
}
