package result

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/luanchang-zh/ChatDesk/consts"
)

// Response 统一响应信封。所有 JSON 接口都走这五个字段，前端不要另认一套。
//
// 前端看 body.code，不要看 HTTP 状态码来区分「参数错误」和「成功」：
// 参数错误也是 HTTP 200，只有 3xxxx 才抬成 500。
type Response struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	TraceId   string      `json:"trace_id"`
	Timestamp int64       `json:"timestamp"`
}

// Result 写出信封。handle 一般不要直接调它，走 Success / Fail / FailServer。
//
// HTTP 状态码策略：
//   - 业务成功、参数错误、未登录：HTTP 200，用 code 区分；
//   - 系统内部错误（code 落在 30000–39999）：HTTP 500。
//     探活和网关只看状态码，不能把「库还没接上」当成 200。
func Result(c *gin.Context, data interface{}, message string, code int) {
	// 键名必须和 ctxmeta.KeyTraceID 一致，Trace 中间件已经 Set 过。
	traceId := c.GetString("trace_id")
	if message == "" {
		message = consts.GetMessage(code)
	}

	httpStatus := http.StatusOK
	if code >= 30000 && code < 40000 {
		httpStatus = http.StatusInternalServerError
	}

	// 访问日志中间件会读这个键，用来判断要 Info 还是 Error。
	c.Set("business_code", code)
	c.JSON(httpStatus, Response{
		Code:      code,
		Message:   message,
		Data:      data,
		TraceId:   traceId,
		Timestamp: time.Now().Unix(),
	})
}

// Success 成功。data 可以是 nil，信封里会是 "data": null。
func Success(c *gin.Context, data interface{}) {
	Result(c, data, "", consts.CodeSuccess)
}

// Fail 业务失败（1xxxx / 2xxxx）。不要把数据库错误原文塞进 data。
// 这条路径不进 gin.Errors，访问日志保持 Info。
func Fail(c *gin.Context, data interface{}, code int) {
	Result(c, data, "", code)
}

// FailServer 服务端错误（3xxxx）：先把 upstreamErr 挂进 gin.Errors，再写信封。
// GinLogger 在请求结束时会把这条错误打进 Error 日志。
// 用户输错参数请用 Fail，不要走这里，否则访问日志会被参数错误刷爆。
func FailServer(c *gin.Context, upstreamErr error, code int) {
	if upstreamErr != nil {
		_ = c.Error(upstreamErr)
	}
	Fail(c, nil, code)
}
