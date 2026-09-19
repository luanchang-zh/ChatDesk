package consts

// 业务码是信封里的 code，不是 HTTP 状态码。
// handle 只传这些整数；不要在 handle 里写 http.StatusBadRequest 当业务结果。
//
// 分段约定（改分段必须同步改 result.Result 和 handleServiceError）：
//
//	0      成功 → HTTP 200
//	1xxxx  客户端问题（参数、方法、体过大）→ HTTP 200 + Fail，不打 Error 日志
//	2xxxx  认证 / 权限 → 同上
//	3xxxx  服务端问题 → HTTP 500 + FailServer，访问日志记 Error
//
// 3xxxx 必须走 HTTP 500：探活和网关只看状态码，不能把「库还没接上」当成 200 成功。
const (
	CodeSuccess = 0 // 成功；错误对象里不允许出现 0，apperr 会改写成 30001
)

const (
	CodeParamError       = 10001 // 查询参数 / JSON 绑定失败。调用方输错，不记 Error。
	CodeBodyError        = 10002 // 请求体格式错误（JSON 都解不开）。
	CodeResourceNotFound = 10003 // 资源不存在。给「打开已删会话」这类情况，不是 404 HTML。
	CodeMethodNotAllowed = 10004 // 方法不允许。预留；当前路由表还没接到这个码。
)

const (
	CodeUnauthorized   = 20001 // 未认证。鉴权中间件接上后再用。
	CodePermissionDeny = 20004 // 已登录但无权限。跳号是为了给以后的 20002/20003 留位置。
)

const (
	CodeInternalError      = 30001 // 未分类的服务端错误。未知 error 一律落到这里，避免把驱动原文当业务码。
	CodeServiceUnavailable = 30002 // 依赖还没就绪。当前会话列表走占位 service，就会返回这个码。
	CodeTimeoutError       = 30003 // 下游超时。和 30001 分开，方便以后做重试策略。
)

// CodeMessage 给信封 message 用的固定文案。
// 未知码不要把 err.Error() 回给客户端，里面可能有 DSN、SQL、堆栈。
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

// GetMessage 按码取文案。找不到就「未知错误」，不要 fmt.Sprintf("%v", err)。
func GetMessage(code int) string {
	if msg, ok := CodeMessage[code]; ok {
		return msg
	}
	return "未知错误"
}

// IsNonServerError 判断是不是「调用方的问题」。
// 1xxxx / 2xxxx 走 result.Fail：HTTP 200、不进 Gin 错误链、不打 Error 日志。
// 参数输错是正常流量，不能让访问日志被刷成一片红。
func IsNonServerError(code int) bool {
	return code >= 10000 && code < 30000
}
