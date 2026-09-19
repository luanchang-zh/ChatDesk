# 配置

进程配置只来自环境变量，由 `internal/config.Load()` 在启动时读一次。

| 变量 | 默认 | 含义 |
|---|---|---|
| `HTTP_ADDR` | `:8080` | 监听地址 |
| `POSTGRES_DSN` | 空 | 下一轮才真正连接；空表示尚未配置 |
| `LOG_LEVEL` | `info` | zap 级别；非法值回退 info |

不要把 DSN 打进日志。启动时只用 `postgres_dsn_set=true/false`。

还没有配置文件、还没有热更新。要加的话先改 `docs/engineering/decisions.md`。
