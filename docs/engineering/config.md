# 配置

进程配置只来自环境变量，由 `internal/config.Load()` 在启动时读一次。

| 变量 | 默认 | 含义 |
|---|---|---|
| `HTTP_ADDR` | `:8080`；开发身份开启时为 `127.0.0.1:8080` | 开发身份仅允许回环 IP，不接受通配地址 |
| `POSTGRES_DSN` | 无，必填 | PostgreSQL URL 或 pgx 连接参数；启动会检查连接与已迁移的结构 |
| `LOG_LEVEL` | `info` | zap 级别；非法值回退 info |
| `DEV_AUTH_ENABLED` | `false` | 显式开启本机开发身份；不是生产认证 |
| `DEV_USER_ID` | 无 | 开发身份开启时必填，非零 UUID，由服务端固定当前用户 |

不要把 DSN 打进日志。启动时只用 `postgres_dsn_set=true/false`。

还没有配置文件、还没有热更新。要加的话先改 `docs/engineering/decisions.md`。

关闭开发身份时，`/health` 可用，`/api/v1` 业务接口返回 `20001`。开启时会幂等初始化开发用户；客户端传入的 `user_id`、身份 Header 和转发头都不能决定身份。

本机入口拒绝非回环连接、非回环 Host、跨站 Fetch Metadata 和不同 Origin。将来接前端时可使用 Vite 同源代理；不要为绕开这些限制而把开发身份开放到局域网。

数据库迁移独立运行，不会在 HTTP 启动时自动执行。缺少 DSN、连接失败、表未就绪或开发身份配置错误都会让启动明确失败。运行命令见 [数据库与会话](database.md)。
