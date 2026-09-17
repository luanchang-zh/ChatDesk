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

// GinRecovery 拦住 panic，避免进程退出。
// 客户端断开（broken pipe）只记 Warn 并中止；其它 panic 记 Error 后返回内部错误。
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
