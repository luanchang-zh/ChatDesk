package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/luanchang-zh/ChatDesk/internal/middleware"
	"github.com/luanchang-zh/ChatDesk/pkg/result"
)

// HealthHandler 探活。不走 /api，方便编排和本地 curl。
type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Check GET /health
func (h *HealthHandler) Check(c *gin.Context) {
	_ = middleware.NewContextWithGin(c)
	result.Success(c, gin.H{"status": "ok"})
}
