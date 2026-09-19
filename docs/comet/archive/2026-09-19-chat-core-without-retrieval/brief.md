# Outcome

实现聊天主链路的第一个工作包：数据库迁移、可信身份入口和会话持久化。用户在阅读规划后明确要求实施这三项，本次确认范围仅覆盖基础层。

# Scope

- PostgreSQL 的 users / conversations 迁移，pgx 连接、sqlc 查询生成、goose 迁移入口。
- 显式开启的本机单用户开发身份；服务端配置用户 UUID，业务接口不接受客户端自报身份，未启用身份时拒绝访问。
- 会话新建、列表、详情、重命名、删除，全部限定当前可信用户；重启后数据仍存在。
- 配置、启动与关闭、可重复开发说明、真实 PostgreSQL 集成测试。

# Non-goals

模型、消息写入、Run 执行、SSE、前端、查找工具、项目管理、登录注册和多实例运行不属于本次实施。完整聊天后续路线保留在 docs/engineering/chat-core-roadmap.md；不为尚未实现的能力提前创建表或占位实现。

# Acceptance examples

- A1：空 PostgreSQL 数据库能通过版本化迁移创建 users / conversations；重复 up 不重复建表；隔离测试库中 down 再 up 可成功。
- A2：显式启用开发身份并配置有效 UUID 后，服务端初始化该用户且只允许回环监听；关闭身份时业务接口返回未认证，健康检查仍可用。
- A3：客户端伪造 user_id、身份 Header 或转发头不能切换用户；开发入口拒绝非回环连接及跨站来源，防止网页借本地开发身份操作数据。
- A4：会话可新建、列表、按 ID 打开、重命名、删除；空列表为数组；重新创建服务或数据库连接后仍读到已提交数据。
- A5：两个可信身份只能访问各自会话；对他人会话 Get / Rename / Delete 均返回与不存在一致的业务结果，写入不会改变他人数据。
- A6：非法 UUID、非法标题和未支持的项目关联有明确业务错误；业务失败沿用 HTTP 200 信封，数据库故障沿用服务端错误规则，不向客户端泄漏驱动原文。
- A7：启动使用真实仓储；缺少 DSN、数据库不可达或迁移未执行时明确失败且不打印 DSN；迁移由独立命令执行，不隐式改动数据库；关闭时释放连接池。
- A8：提交可重现的 sqlc 配置、生成代码、启动/迁移示例与测试；真实 PostgreSQL 下迁移、CRUD、隔离测试通过，Go 测试与静态检查通过。

# Constraints and invariants

沿用现有 Gin / domain / service 分层与 zap 日志封装；不把 Gin Context、sqlc 类型或驱动错误暴露到领域和 HTTP 契约。用户归属来自可信 Context，不依赖 Input.UserID。单实例、本机开发身份不代表完成生产认证。普通接口保持 code/message/data/trace_id/timestamp 信封。

# Decisions

- 2026-09-19：用户明确指示“那你做一下 数据库迁移、可信身份入口和会话持久化”，确认此前规划的第一个工作包；不重复请求相同授权。
- 使用既定 pgx / sqlc / goose。只迁移已实现的 users / conversations，后续消息版本与 Run 关系随聊天工作包迁移。
- 开发身份仅显式开启、回环地址可用；默认没有身份时拒绝业务访问。增加回环 Host / 来源校验作为本机开发身份的必要边界。
- 保留当前工作目录与单一 change。三项共享配置、身份和数据契约，按顺序实现，无需拆分子 change。
- 用户随后允许使用 Docker 提供真实 PostgreSQL，以完成本轮迁移、CRUD 与隔离验收。测试仍不自动拉起容器；只在显式配置 `CHATDESK_TEST_ADMIN_DSN` 时连库。未配置时 SKIP，不作为实库通过。

# Open questions

无阻塞问题。用户已确认：范围与 A1–A8 不变，允许用 Docker 提供真实 PostgreSQL 完成本轮验收。

# Verification expectations

真实 PostgreSQL 集成测试覆盖迁移重入与回滚、CRUD 持久化、双身份隔离；HTTP 测试覆盖信封、恶意自报身份、参数边界和本机访问限制。执行 go test ./...、go vet ./...，核对 sqlc 生成结果。候选完成后按 Native Runtime 派发新的只读 Verifier。
