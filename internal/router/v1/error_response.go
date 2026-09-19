package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/luanchang-zh/ChatDesk/consts"
	"github.com/luanchang-zh/ChatDesk/pkg/apperr"
	"github.com/luanchang-zh/ChatDesk/pkg/result"
)

// handleServiceError 把下游错误转成统一 HTTP 信封。各 handle 不要自己 switch 业务码。
//
// 分支：
//  1. 1xxxx / 2xxxx：客户端或鉴权问题 → Fail，HTTP 200，不进错误日志；
//  2. 已有的 3xxxx（例如 30002 服务暂不可用）→ FailServer，HTTP 500，原样返回该码。
//     不能把 30002 折叠成 30001，否则前端和探活分不清「还没接库」和「未知崩溃」；
//  3. 解析不出业务码，或 code==0：当成内部错误。
//     0 在信封里表示成功，失败请求绝不能把成功码回出去。
func handleServiceError(c *gin.Context, err error) {
	code := apperr.Code(err)
	if consts.IsNonServerError(code) {
		result.Fail(c, nil, code)
		return
	}
	if code == consts.CodeSuccess {
		code = consts.CodeInternalError
	}
	result.FailServer(c, err, code)
}
