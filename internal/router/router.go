package router

import (
	"github.com/gin-gonic/gin"
	"github.com/luanchang-zh/ChatDesk/internal/middleware"
	v1 "github.com/luanchang-zh/ChatDesk/internal/router/v1"
)

// NewEngine 组装 Gin。中间件顺序：
// Recovery → Trace → Logger → 业务路由。
// 本轮不挂鉴权、限流、超时表、Prometheus。
func NewEngine(healthHandler *v1.HealthHandler, conversationHandler *v1.ConversationHandler) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// 本地开发不信任 X-Forwarded-For，避免 ClientIP 被伪造。
	_ = r.SetTrustedProxies(nil)

	r.Use(middleware.GinRecovery())
	r.Use(middleware.Trace())
	r.Use(middleware.GinLogger())

	r.GET("/health", healthHandler.Check)

	api := r.Group("/api/v1")
	{
		conversations := api.Group("/conversations")
		{
			conversations.GET("", conversationHandler.List)
		}
	}

	return r
}
