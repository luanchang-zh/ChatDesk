package middleware

import (
	"errors"
	"net"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/luanchang-zh/ChatDesk/consts"
	"github.com/luanchang-zh/ChatDesk/pkg/apperr"
	"github.com/luanchang-zh/ChatDesk/pkg/logger"
	"github.com/luanchang-zh/ChatDesk/pkg/result"
)

// GinRecovery 拦住 handler / 中间件里的 panic，避免整个进程退出。
//
// 必须挂在 Use 链最前面：后面 Trace / Logger / handle 任何一层炸了都要能接住。
//
// 两种情况：
//  1. 客户端已经断开（broken pipe / connection reset）：再写响应也没用，只记 Warn 然后 Abort；
//  2. 其它 panic：记 Error（带 top_frame 方便定位），再返回内部错误信封。
//
// 不要把 panic 原文写进信封 message，里面可能有内部路径。
func GinRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			recovered := recover()
			if recovered == nil {
				return
			}

			ctx := NewContextWithGin(c)

			if isBrokenPipe(recovered) {
				logger.WithCtx(ctx).Warn("客户端断开连接",
					logger.Any("error", recovered),
					logger.String("method", c.Request.Method),
					logger.String("path", c.Request.URL.Path),
				)
				c.Abort()
				return
			}

			panicErr := apperr.NewFromPanic(recovered)
			logger.WithCtx(ctx).Error("捕获到 panic",
				logger.Any("error", recovered),
				logger.String("method", c.Request.Method),
				logger.String("path", c.Request.URL.Path),
				logger.String("top_frame", apperr.TopFrame(panicErr)),
			)
			result.Fail(c, nil, consts.CodeInternalError)
		}()
		c.Next()
	}
}

// isBrokenPipe 判断 panic 是否只是对端把 TCP 掐了。
// 浏览器刷新、客户端超时取消，写响应时经常炸这个，不是我们的 bug，不能当 Error。
func isBrokenPipe(recovered any) bool {
	ne, ok := recovered.(*net.OpError)
	if !ok {
		return false
	}
	var se *os.SyscallError
	if !errors.As(ne.Err, &se) {
		return false
	}
	errStr := strings.ToLower(se.Error())
	return strings.Contains(errStr, "broken pipe") || strings.Contains(errStr, "connection reset by peer")
}
