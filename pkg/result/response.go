package result

import (
	"net/http"
	"time"

	"github.com/luanchang-zh/ChatDesk/consts"
	"github.com/gin-gonic/gin"
)

// Response 业务码在 body，HTTP 状态码另算。
type Response struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	TraceId   string      `json:"trace_id"`
	Timestamp int64       `json:"timestamp"`
}

// Result HTTP 状态码策略：
//   - 业务成功或业务失败：HTTP 200，body.code 区分
//   - 系统内部错误（code >= 30000）：HTTP 500
func Result(c *gin.Context, data interface{}, message string, code int) {
	traceId := c.GetString("trace_id")
	if message == "" {
		message = consts.GetMessage(code)
	}

	httpStatus := http.StatusOK
	if code >= 30000 && code < 40000 {
		httpStatus = http.StatusInternalServerError
	}

	c.Set("business_code", code)
	c.JSON(httpStatus, Response{
		Code:      code,
		Message:   message,
		Data:      data,
		TraceId:   traceId,
		Timestamp: time.Now().Unix(),
	})
}

func Success(c *gin.Context, data interface{}) {
	Result(c, data, "", consts.CodeSuccess)
}

func Fail(c *gin.Context, data interface{}, code int) {
	Result(c, data, "", code)
}

// FailServer 服务端错误：写入响应，并把 upstreamErr 挂进 gin.Errors，交给 GinLogger 打日志。
// 业务错误只用 Fail，不要把用户输入问题当服务器错误。
func FailServer(c *gin.Context, upstreamErr error, code int) {
	if upstreamErr != nil {
		_ = c.Error(upstreamErr)
	}
	Fail(c, nil, code)
}
