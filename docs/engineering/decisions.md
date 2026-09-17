# 决策日志

按时间追加。新决定如果否定旧决定，必须在本文件写明「替代哪一条」，不要只改代码。

## 2026-09-16 产品与运行时

- **对前端只有一个后端。** 浏览器只打 Go。Python 是内部知识工人，不是第二个产品 API。
- **一个 Go + 一个 Python 进程。** 解析、RAG、图谱都在同一个 Python 知识工人里，不再拆三个 FastAPI。
- **Neo4j 是图存储，不是后端。** 图谱是知识工人内部模块。
- **不按 M0～M4 分期。** 完整首版一次性覆盖四路检索。P0 / P1 / P2 只表示优先级：Agent 是 P1，沙箱是 P2。
- **HTTP 框架用 Gin，不用 Chi。**
- **DOCX 是 P1。** 首版文件是 TXT / Markdown / 可提取 PDF。
- **沙箱是 P2，不进当前主体架构。**

## 2026-09-17 HTTP / 错误 / 日志

- **Handle 层：** `internal/router` + `internal/router/v1/*_handle.go`，不把 `*gin.Context` 传入业务。
- **响应信封：** `code/message/data/trace_id/timestamp`。业务失败 HTTP 200；`code >= 30000` 才 HTTP 500。
- **不要改成「用 HTTP 状态码当业务码」。** 那会和现有 handle、前端约定打架。
- **参数错误用 `result.Fail(c, nil, consts.CodeParamError)`，不打错误日志。** 用户输错不是事故。
- **服务端错误用 `result.FailServer`，由 GinLogger 统一 `logger.WithCtx(ctx).Error`。** handler 里不要再打一遍同样的错误。
- **日志两种写法都保留：** `logger.Error(ctx, msg, fields...)` 与 `logger.WithCtx(ctx).Error(msg, fields...)`。字段用 `logger.String` / `logger.ErrorField`，业务代码不直接 import zap。
- **日志编码用 console，不用 JSON。** 本地可读优先；字段仍是结构化的。
- **本轮只做 handle。** 不写 Postgres、sqlc、Eino、Python 客户端、鉴权、限流、超时表、Prometheus。未实现接口返回 `CodeServiceUnavailable`。
- **不引入 Wire。** 单二进制、依赖很少，`cmd/server/main.go` 手工组装即可。
- **不信任反向代理头。** `SetTrustedProxies(nil)`。后面若上 Nginx，再单独改，并记入本文件。
