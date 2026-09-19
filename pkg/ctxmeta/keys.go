package ctxmeta

// 本包把「一条请求上要跟着走的元数据」收口到同一组键。
//
// 为什么单独成包：
//  1. Gin 的 c.Set、标准 context.Value、日志字段、响应信封里的 trace_id
//     必须用同一字符串，否则日志对不上信封、信封对不上响应头；
//  2. handle / service / logger 都只依赖本包常量，禁止在业务代码里再写 "trace_id"。

const (
	// KeyTraceID 链路 ID。
	// 同时用作：gin.Context 键、context.Value 键、日志字段名、信封 json:"trace_id"。
	// 改名字必须四处一起改，否则访问日志和前端对账会断。
	KeyTraceID = "trace_id"

	// KeyUserUUID 当前登录用户。鉴权接上后再写入；空表示匿名，不要存空串占位。
	KeyUserUUID = "user_uuid"

	// KeyClientIP 对端 IP。访问日志会打一份；业务层一般用不到。
	KeyClientIP = "client_ip"
)

const (
	// HeaderRequestID 请求进出都用这个头透传。
	// 入站：上游（浏览器、反向代理）带来则沿用，便于整条请求对账；
	// 出站：原样回写，curl -D - 或前端就能对着日志搜。
	// 不要再发明 X-Trace-Id / X-Correlation-Id，多一个头等于多一套对不齐的 ID。
	HeaderRequestID = "X-Request-ID"
)
