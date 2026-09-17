# 错误处理

## 信封

```json
{"code":0,"message":"success","data":{},"trace_id":"...","timestamp":0}
```

- `code == 0`：成功，HTTP 200
- `10000 <= code < 30000`：业务失败，HTTP 200，前端读 `code`
- `30000 <= code < 40000`：服务端错误，HTTP 500

## Handle 两条路径

| 场景 | 调用 | 日志 |
|---|---|---|
| 绑定失败、缺参数 | `result.Fail(c, nil, consts.CodeParamError)` | 不记 |
| 可识别业务失败 | `result.Fail(c, nil, code)` | 不记 Error |
| 未知 / 内部错误 | `result.FailServer(c, err, consts.CodeInternalError)` | GinLogger Error |
| panic | Recovery → `result.Fail(..., CodeInternalError)` | Recovery Error |

下游错误不要在每个 handle 里自己 `switch`，走 `handleServiceError`：

- `consts.IsNonServerError(apperr.Code(err))` → `Fail`
- 其它 → `FailServer`，HTTP 500，错误进 `c.Errors`

## 业务码

通用码放 `consts/code.go`。模块专用码（知识库、图谱）以后按 11xxx、12xxx 分段加，不要复用 HTTP 状态码当业务码。
