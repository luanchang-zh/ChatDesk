# 聊天基础层：数据库、可信身份与会话持久化

用户于 2026-09-19 在查看聊天路线后明确要求先实施这一工作包。完整后续路线见 docs/engineering/chat-core-roadmap.md，不属于本次交付。

执行约束：用户已允许使用 Docker 提供真实 PostgreSQL。集成测试不自动启动容器；配置 `CHATDESK_TEST_ADMIN_DSN` 后对隔离库执行迁移、CRUD 与双身份验证。未配置时明确 skip，验收必须区分代码检查通过和数据库运行未验证。

## 目标行为

系统用 PostgreSQL 保存用户归属和会话；HTTP 接口通过可信上下文获取用户身份。新建、列表、详情、重命名与删除结果在服务重启后保持一致。

## 数据库与迁移

- 使用 goose 的版本化 SQL 迁移；提供单独的 Go 迁移命令，支持 up / down / status。down 只在用户显式执行时发生，测试使用隔离数据库。
- users 保存 UUID 与创建时间；conversations 保存 UUID、user_id 外键、title、创建与更新时间，并有按用户和时间查询的索引。UUID 由服务端生成；所有时间使用 timestamptz。
- SQL 查询通过固定版本 sqlc 生成 pgx/v5 代码；提交配置和生成代码，不要求使用者全局安装 sqlc。
- 独立迁移命令读取 POSTGRES_DSN，不把 DSN 当作命令行参数。服务启动只连接、检查依赖与已有表；不自动迁移，不创建数据库。
- 只建立当前会话功能需要的表。消息、Run、事件、项目表在相应能力实现时添加，不能因此承诺本轮可保存聊天内容。

## 可信身份与开发模式

- 默认不启用开发身份。未认证的 /api/v1 业务请求返回 CodeUnauthorized；/health 保持无需身份。
- DEV_AUTH_ENABLED=true 且 DEV_USER_ID 为有效非零 UUID 时，显式启用本机单用户开发身份。服务启动幂等初始化该 users 记录，身份只来自服务端环境配置。
- 开发身份模式默认监听 127.0.0.1:8080；如指定 HTTP_ADDR，必须使用回环 IP。拒绝通配地址和非回环 IP，不能信任转发头绕过限制。
- 开发身份请求必须来自回环连接、回环 Host；有 Origin 时必须是回环来源且与服务入口同源；跨站 Fetch Metadata 拒绝。无浏览器来源 Header 的本机命令行请求允许。
- 用户 ID 不从请求体、查询参数、Authorization 的未验证内容或 X-User-* Header 读取。中间件注入 Gin 用户元数据，再由现有 Context 转换传入 service。
- service 始终从 Context 获取身份，拒绝空或非法身份；即使调用方填入其他 Input.UserID 也不能越权。未来真实认证可以复用这个边界，但本轮不实现登录注册。

## 会话接口

路径统一位于 /api/v1。沿用项目既有响应信封，成功及业务失败使用 HTTP 200；服务端错误使用 HTTP 500。未授权与不存在不得返回驱动细节。

| 方法 | 路径 | 行为 |
|---|---|---|
| POST | /conversations | 创建会话；title 可省略或空白，默认“新会话” |
| GET | /conversations | 当前用户会话列表，按更新时间与 ID 稳定倒序；空结果 items=[] |
| GET | /conversations/:id | 当前用户会话详情 |
| PATCH | /conversations/:id | 修改 title，不能修改归属 |
| DELETE | /conversations/:id | 删除当前用户会话；不存在返回资源不存在 |

- 标题去除首尾空白后最多 200 个 Unicode 字符；重命名为空或超长返回参数错误。创建超长标题同样拒绝。
- 路径 ID 必须是合法非零 UUID。未知/他人 ID 返回相同的 CodeResourceNotFound，不泄漏是否属于别人。
- ProjectID 仍保留既有领域概念；本轮不支持非空 project_id，明确返回参数错误，不静默忽略。
- 所有 select / update / delete 均同时限定会话 ID 和当前 user_id；列表限定 user_id；创建的 user_id 只由服务端确定。
- 返回 DTO 沿用 id / project_id / title / created_at / updated_at，project_id 当前为空，不返回 owner。时间为 UTC RFC3339。
- JSON 请求体有合理大小上限；格式或校验失败返回既有业务错误。

## 运行和工程边界

Config 仍只从环境变量读取。POSTGRES_DSN 现在是服务启动前置条件；配置错误、连接失败或迁移未就绪时退出，日志只说明类别，不输出 DSN、密码或底层连接串。

main 初始化连接池、检查已有表、初始化开发用户，构造真实 repository / service / handler；收到退出信号先关闭 HTTP，随后释放连接池。没有数据库时不再假装可用地注入 UnavailableConversations。

实现继续沿用 internal/router/v1、internal/service、internal/domain；存储放 internal/repository，sqlc 生成代码单独放置。不引入 Wire、通用 Repository 基类或与本轮无关的框架。

## 验收和验证

正式验收使用 brief.md 的 A1–A8。真实 PostgreSQL 测试创建专用临时数据库，拒绝对未知现有数据库执行 down；测试结束只删除本次创建的数据库。测试覆盖二次 up、down/up、服务重建后读取、创建/列表/详情/重命名/删除、双用户隔离与参数行为。

配置与身份入口测试覆盖 disabled、无效 UUID、非回环监听、伪造 Host / Origin / 转发头。所有请求沿用现有信封，数据库异常不得把原始驱动内容返回给客户端。没有真实数据库时集成测试明确 skip，不作为已通过的证据；本轮交付应实际运行一次数据库集成测试。
