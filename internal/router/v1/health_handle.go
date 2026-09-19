package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/luanchang-zh/ChatDesk/internal/middleware"
	"github.com/luanchang-zh/ChatDesk/pkg/result"
)

// HealthHandler 进程探活。
//
// 路径固定 GET /health，故意不进 /api/v1：
//  1. 编排系统、本地 curl、负载均衡探活都只认这个短路径；
//  2. 成功时 GinLogger 会跳过这条，避免每秒一次的探活把访问日志刷没。
//
// 本轮只回答「进程还活着」，不探 Postgres / Python 进程。
// 连库之后如果要做就绪检查，另开 /ready，不要把失败的依赖塞进 /health，否则发布滚动会把实例摘光。
type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Check 返回 {status: ok}。
// 仍走 NewContextWithGin：信封里的 trace_id 和其它接口同一套，排障时不用两套习惯。
func (h *HealthHandler) Check(c *gin.Context) {
	_ = middleware.NewContextWithGin(c)
	result.Success(c, gin.H{"status": "ok"})
}
