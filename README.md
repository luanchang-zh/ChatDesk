# ChatDesk

Go 聊天工作台。当前已实现 PostgreSQL 会话持久化和显式本机开发身份，聊天生成与查找能力仍在后续路线中。

- [启动、迁移与会话接口](docs/engineering/database.md)
- [环境变量](docs/engineering/config.md)
- [工程决策](docs/engineering/decisions.md)
- [后续聊天路线](docs/engineering/chat-core-roadmap.md)

基本检查：`go test ./...`、`go vet ./...`。实库测试需要单独配置 `CHATDESK_TEST_ADMIN_DSN`；没有配置时会明确跳过，不自动启动 Docker。
