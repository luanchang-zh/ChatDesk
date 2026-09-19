package router

import (
	"github.com/gin-gonic/gin"
	"github.com/luanchang-zh/ChatDesk/internal/middleware"
	v1 "github.com/luanchang-zh/ChatDesk/internal/router/v1"
)

// NewEngine 组装 Gin 引擎并注册路由。
//
// 中间件顺序有意固定，不要调换：
//  1. Recovery：panic 不能把整个进程打死，必须最先挂上；
//  2. Trace：后面日志和响应信封都要 trace_id，必须在 Logger / handle 之前；
//  3. Logger：必须在业务 handler 之后看到最终状态码，所以放 Use 链里、在 c.Next() 后打日志。
//
// devUserID 为空时，业务接口保持未认证；健康检查始终开放。
func NewEngine(healthHandler *v1.HealthHandler, conversationHandler *v1.ConversationHandler, devUserID string) *gin.Engine {
	// ReleaseMode 关掉 Gin 自带的 debug 路由打印，访问日志只走我们的 console/zap 封装。
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// 不信任 X-Forwarded-For。后面若放到 Nginx 后面，再显式配置可信网段。
	// 现在信任全部转发头的话，客户端可以伪造 IP，访问日志就脏了。
	_ = r.SetTrustedProxies(nil)

	r.Use(middleware.GinRecovery())
	r.Use(middleware.Trace())
	r.Use(middleware.GinLogger())

	// 探活不进 /api/v1，编排系统和本地 curl 都简单。
	r.GET("/health", healthHandler.Check)

	api := r.Group("/api/v1")
	api.Use(middleware.DevIdentity(devUserID))
	{
		conversations := api.Group("/conversations")
		{
			conversations.GET("", conversationHandler.List)
			conversations.POST("", conversationHandler.Create)
			conversations.GET("/:id", conversationHandler.Get)
			conversations.PATCH("/:id", conversationHandler.Rename)
			conversations.DELETE("/:id", conversationHandler.Delete)
		}
	}

	return r
}
