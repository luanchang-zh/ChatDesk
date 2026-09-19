# ChatDesk 工程文档

本目录记录 **已经拍板的选型与挑择**。后面改代码前先读这里，避免和 `docs/total` 产品方案或已落地的 HTTP 习惯打架。

| 文件 | 内容 |
|---|---|
| [decisions.md](./decisions.md) | 决策日志（含产品架构收缩、HTTP/日志习惯） |
| [http-gin.md](./http-gin.md) | Gin 编排、handle 职责、中间件顺序 |
| [error-handling.md](./error-handling.md) | 业务码、result.Fail / FailServer |
| [logging.md](./logging.md) | logger.WithCtx / logger.Error 用法 |
| [config.md](./config.md) | 环境变量：HTTP / DSN / 日志级别 |

产品需求、技术栈、系统架构仍在 `docs/total/`。这里只补「代码怎么写、为什么不能改回去」。
