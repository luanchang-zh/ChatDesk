# 决策日志

按时间追加。新决定如果否定旧决定，必须在本文件写明「替代哪一条」，不要只改代码。

## 2026-09-16 产品与运行时

- **非基础设施进程只有 Go 和 Python，不是微服务。** 浏览器只打 Go。Python 是配套进程，不对浏览器暴露，也不按解析 / RAG / 图谱拆成多个进程。
- **PostgreSQL、Neo4j、文件目录、模型供应商、Tavily 是基础设施或外部依赖。** 不算第三个产品进程。图谱是 Python 进程内部模块。
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
- **本轮已接配置与领域服务接口。** 仍不写 Postgres、sqlc、Eino、对 Python 进程的调用、鉴权、限流、超时表、Prometheus。未实现的 service 返回 `CodeServiceUnavailable`。
- **不引入 Wire。** 单二进制、依赖很少，`cmd/server/main.go` 手工组装即可。
- **不信任反向代理头。** `SetTrustedProxies(nil)`。后面若上 Nginx，再单独改，并记入本文件。

## 2026-09-17 配置与领域服务

- **配置只从环境变量读：** `HTTP_ADDR`（默认 `:8080`）、`POSTGRES_DSN`、`LOG_LEVEL`（默认 `info`）。启动日志只打 DSN 是否已配置，不打印连接串。
- **领域模型放 `internal/domain`，服务接口放 `internal/service`。** 四个接口：Project / Conversation / Message / Run。handle 只依赖接口，领域结构体不带 json tag。
- **当前注入 `Unavailable*` 占位实现。** 所有方法返回 `CodeServiceUnavailable`。下一轮用 sqlc 替换实现，不改 handle 签名。
- **已知服务端业务码（如 30002）经 `handleServiceError` 原样返回。** 未知错误才收成 `CodeInternalError`。

## 2026-09-19 进程边界措辞

- **替代 2026-09-16「知识工人 / 第二个产品 API」的说法。** 统一写成：非基础设施进程只有 Go 和 Python，不是微服务；PostgreSQL / Neo4j / 文件目录 / 模型供应商是基础设施或外部依赖。文档与注释里不再用「跨服务」「微服务」描述这两个进程。

## 2026-09-19 数据库、开发身份与会话持久化

- **替代 2026-09-17 的会话占位与暂不接库决定。** `main` 注入 pgx / sqlc 真实仓储和会话 service；POSTGRES_DSN 成为启动前置条件，迁移未执行时启动失败。
- **迁移使用 goose，由独立 `cmd/migrate` 执行。** HTTP 启动不自动改变结构；当前只创建 users / conversations，消息、Run 和事件随对应工作包增加。
- **身份仅由服务端确定。** 开发模式必须显式开启、配置用户 UUID 并仅监听回环 IP；中间件校验直接连接、Host 与 Origin。默认业务接口未认证，健康检查公开。客户端自报用户身份没有效力。
- **所有会话 SQL 同时限定 user_id。** 不存在与属于其他用户返回同一业务码；service 从 Context 读取用户，不再通过会话 Input 传 UserID。
- **会话接口补齐新建、列表、详情、重命名、删除。** 空标题创建为“新会话”，重命名不能为空，标题最多 200 个 Unicode 字符。项目关联暂未实现，非空 project_id 明确拒绝。
- **沿用现有 HTTP 信封与日志封装。** 普通业务错误 HTTP 200；服务端错误 HTTP 500。未变更为基于 HTTP 状态码的业务协议。
- **替代上一条「本轮不运行 Docker」。** 用户已允许用 Docker 提供测试库。自动测试仍不自己启动容器；配置 `CHATDESK_TEST_ADMIN_DSN` 后运行隔离库验收，未配置则 SKIP 且不声称实库通过。
