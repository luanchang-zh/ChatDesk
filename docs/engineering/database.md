# 数据库与会话

当前完成的是会话管理基础层：users / conversations 持久化及本机开发身份。还没有消息生成、Run、SSE、前端和工具接入。

## 启动

需要 Go 1.25+ 和已创建的 PostgreSQL 数据库。此项目不会自动安装或启动 PostgreSQL，也不要求 Docker。生产认证暂未实现，不应把开发身份入口用于对外部署。

PowerShell 示例（替换账号、密码和数据库）：

```powershell
$env:POSTGRES_DSN = 'postgres://chatdesk:replace-me@127.0.0.1:5432/chatdesk?sslmode=disable'
go run ./cmd/migrate up
go run ./cmd/migrate status
$env:DEV_AUTH_ENABLED = 'true'
$env:DEV_USER_ID = 'a525e8f9-1f6a-47b4-be50-c7ce5afde696'
$env:HTTP_ADDR = '127.0.0.1:8080'
go run ./cmd/server
```

服务启动会检查库与表，并幂等创建配置的开发用户。默认关闭开发身份时业务接口返回未认证。DSN 不作为命令行参数或日志输出；`.env.example` 只用于说明变量，程序不会自动读取 `.env`。

迁移文件位于 `internal/database/migrations/`。`go run ./cmd/migrate down` 回滚最近一个迁移；当前首个迁移的回滚会删除会话与用户表，只能对确认可丢弃的数据库执行。HTTP 启动只检查结构，不自动迁移。

## 会话接口

普通接口遵循 `code/message/data/trace_id/timestamp` 信封。参数、未认证、不存在等业务失败返回 HTTP 200；服务端错误为 HTTP 500。

| 方法 | 路径 | 请求或结果 |
|---|---|---|
| POST | /api/v1/conversations | `{"title":"学习笔记"}`；省略或空白标题为“新会话” |
| GET | /api/v1/conversations | 当前用户列表，`data.items`；空列表为 `[]` |
| GET | /api/v1/conversations/:id | 当前用户会话详情 |
| PATCH | /api/v1/conversations/:id | `{"title":"新标题"}` |
| DELETE | /api/v1/conversations/:id | 删除成功后 `data=null` |

```powershell
$conversation = Invoke-RestMethod -Method Post -Uri 'http://127.0.0.1:8080/api/v1/conversations' -ContentType 'application/json' -Body '{"title":"学习笔记"}'
Invoke-RestMethod -Uri "http://127.0.0.1:8080/api/v1/conversations/$($conversation.data.id)"
```

标题去除首尾空白后最多 200 个 Unicode 字符；重命名不能为空。JSON 正文最大 64 KiB。资源 ID 要求非零 UUID；他人的 ID 与不存在的 ID 返回相同 `10003`。非空 `project_id` 返回 `10001`，当前不支持项目关联。

身份来自服务端配置，HTTP 传入 `user_id` 或 `X-User-ID` 不会切换用户。开发入口校验回环连接与 Host，并拒绝跨 Origin/跨站浏览器请求；不信任 `X-Forwarded-*`。

## 生成查询与验证

依赖固定为 pgx v5.11.0、goose v3.26.0；sqlc 生成器固定为 v1.30.0。后两个版本兼容当前 Go 1.25 工具链。生成代码提交在 `internal/repository/dbgen/`，不要手改。

```powershell
# sqlc 使用纯 Go 解析路径，避开 Windows 本机 C 链接器差异
$env:CGO_ENABLED = '0'
go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0 generate
go test ./...
go vet ./...
```

Git Bash 可以使用 `CGO_ENABLED=0 go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0 generate`，只对该命令设置环境变量。生成和编译不需要启动 PostgreSQL。

真实 PostgreSQL 测试另需有 CREATEDB 权限的测试账号：

```powershell
$env:CHATDESK_TEST_ADMIN_DSN = 'postgres://test-admin:replace-me@127.0.0.1:5432/postgres?sslmode=disable'
go test ./internal/integration -v -count=1
```

测试为每次运行创建随机命名的 `chatdesk_test_*` 数据库，先核对实际连接库名，再执行 up / 重复 up / down / up、CRUD、双用户隔离和关闭连接后重新读取，结束时只清理本次创建的数据库。不会对已有业务库执行回滚。

未配置 `CHATDESK_TEST_ADMIN_DSN` 时测试明确 SKIP；普通 `go test ./...` 的成功不能证明实库测试已经执行。可用本机或 Docker 中的 PostgreSQL，账号需有 CREATEDB 权限。
