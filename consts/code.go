package consts

// 业务码分段：
// 0 成功；1xxxx 客户端；2xxxx 认证；3xxxx 服务端。
// HTTP 状态码由 pkg/result 按区间映射，handler 只传业务码。

const (
	CodeSuccess = 0
)

const (
	CodeParamError       = 10001
	CodeBodyError        = 10002
	CodeResourceNotFound = 10003
	CodeMethodNotAllowed = 10004
)

const (
	CodeUnauthorized   = 20001
	CodePermissionDeny = 20004
)

const (
	CodeInternalError      = 30001
	CodeServiceUnavailable = 30002
	CodeTimeoutError       = 30003
)

var CodeMessage = map[int]string{
	CodeSuccess:            "success",
	CodeParamError:         "参数验证失败",
	CodeBodyError:          "请求体格式错误",
	CodeResourceNotFound:   "资源不存在",
	CodeMethodNotAllowed:   "请求方法不允许",
	CodeUnauthorized:       "未认证",
	CodePermissionDeny:     "权限不足",
	CodeInternalError:      "服务器内部错误",
	CodeServiceUnavailable: "服务暂不可用",
	CodeTimeoutError:       "超时错误",
}

// GetMessage 根据业务码取文案；未知码不要把内部实现细节回给客户端。
func GetMessage(code int) string {
	if msg, ok := CodeMessage[code]; ok {
		return msg
	}
	return "未知错误"
}

// IsNonServerError 可识别的业务失败（参数、鉴权等）走 Fail，不进 Gin 错误链。
func IsNonServerError(code int) bool {
	return code >= 10000 && code < 30000
}
