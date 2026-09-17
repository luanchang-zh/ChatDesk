# 日志

输出格式用 console，不用 JSON。

## 初始化

进程启动时 `logger.InitConsole()`。未初始化时是 Nop，不会 panic。

## 推荐写法

handler / 中间件里优先：

```go
ctx := middleware.NewContextWithGin(c)
logger.WithCtx(ctx).Error("会话列表失败", logger.ErrorField("error", err))
```

包级函数同样可用：

```go
logger.Error(ctx, "会话列表失败", logger.ErrorField("error", err))
```

两种都自动带上 `trace_id`。不要业务包直接 `import go.uber.org/zap`。

## 字段

用封装函数，避免各处 zap API 不一致：

- `logger.String` / `logger.Int` / `logger.Duration` / `logger.Any`
- 错误字段：`logger.ErrorField("error", err)`

## 什么时候打

| 级别 | 场景 |
|---|---|
| Info | 普通请求结束（GinLogger） |
| Warn | 慢请求、客户端断开 |
| Error | HTTP 5xx、业务码 >= 30000、panic |
| 不打 | 参数错误、4xx 业务失败 |

`/health` 成功不打访问日志，避免探活刷屏。
