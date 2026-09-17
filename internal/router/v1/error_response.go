package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/luanchang-zh/ChatDesk/consts"
	"github.com/luanchang-zh/ChatDesk/pkg/apperr"
	"github.com/luanchang-zh/ChatDesk/pkg/result"
)

// handleServiceError 把下游错误转成 HTTP 响应：
//  1. 可识别的非服务端业务错误：Fail，HTTP 200，body.code 为业务码；
//  2. 未知错误或服务端错误：FailServer，挂入 Gin 错误链给日志中间件。
func handleServiceError(c *gin.Context, err error) {
	code := apperr.Code(err)
	if consts.IsNonServerError(code) {
		result.Fail(c, nil, code)
		return
	}
	result.FailServer(c, err, consts.CodeInternalError)
}
