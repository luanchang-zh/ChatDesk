# Gin 编排

## 目录

```text
cmd/server/main.go              启动、信号、Shutdown
internal/router/router.go       gin.New、中间件、路由组
internal/router/v1/*_handle.go  一个资源一个 handle 文件
internal/middleware/            Recovery / Trace / 访问日志
```

不按 conversation / knowledge 再切垂直模块。接口变多后再拆，拆之前先改 `decisions.md`。

## 中间件顺序

只保留当前用得到的三层：

1. `GinRecovery`：panic 不能打死进程
2. `Trace`：生成或透传 `X-Request-ID`
3. `GinLogger`：请求结束打一条访问日志

鉴权、限流、超时表、CORS、Prometheus 本轮不上。健康检查 `GET /health` 不进 `/api/v1`。

## Handle 写法

一个方法只做：

1. `ShouldBindJSON` / `ShouldBindQuery`（不要用 `BindJSON`，它会自己写 400）
2. `ctx := middleware.NewContextWithGin(c)`
3. 调服务（本轮没有服务就返回业务码）
4. `result.Success` 或 `handleServiceError`

禁止：

- 把 `*gin.Context` 传入 service
- handler 里直接打数据库或 Python
- 在 handler 里拼 `gin.H{"error": ...}` 绕过 `pkg/result`
