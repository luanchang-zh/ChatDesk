package service

import (
	"context"

	"github.com/luanchang-zh/ChatDesk/internal/domain"
)

// 本包定义领域服务接口。handle 只依赖这些接口，不依赖 Postgres / sqlc / Python。
//
// 硬约束：
//  1. 第一参数必须是 context.Context，用来透传取消和 trace_id；
//  2. 禁止出现 *gin.Context：service 不能绑定 HTTP、不能写信封、不能读 Header；
//  3. 入参用 domain 的 Input 结构，不要把 HTTP DTO 漏进这一层。
//
// 当前实现是 Unavailable* 占位。下一轮 sqlc 仓储接上后，只换 main 里的注入，handle 不用改。

// ProjectService 项目管理。
// 修改说明只影响后续 Run，不回头改已经完成的回答；删除有关联会话时由实现层决定是否拒绝。
type ProjectService interface {
	// Create 新建项目。Name / Description 的合法性由实现校验。
	Create(ctx context.Context, in domain.CreateProjectInput) (*domain.Project, error)
	// Get 按 ID 取单个项目。不存在应返回业务码，而不是 (nil, nil)。
	Get(ctx context.Context, id string) (*domain.Project, error)
	// List 列出某用户的项目。userID 来自鉴权后的 context，不要信任客户端传入的用户 ID。
	List(ctx context.Context, userID string) ([]domain.Project, error)
	// Update 更新名称和说明。只影响后续 Run。
	Update(ctx context.Context, in domain.UpdateProjectInput) (*domain.Project, error)
	// Delete 删除项目。有关联会话时是否级联，由实现和产品约定，handle 不猜。
	Delete(ctx context.Context, id string) error
}

// ConversationService 会话管理。
// handle 列会话、以后新建 / 打开 / 重命名 / 删除都走这里。
type ConversationService interface {
	// Create 新建会话。Title 可空，由后续首条消息再命名。
	Create(ctx context.Context, in domain.CreateConversationInput) (*domain.Conversation, error)
	// Get 按 ID 打开会话。刷新页面后要能靠这个恢复。
	Get(ctx context.Context, id string) (*domain.Conversation, error)
	// List 列出会话。ProjectID 空表示不按项目过滤。当前占位实现会返回 CodeServiceUnavailable。
	List(ctx context.Context, in domain.ListConversationsInput) ([]domain.Conversation, error)
	// Rename 重命名，不改历史消息。
	Rename(ctx context.Context, in domain.RenameConversationInput) (*domain.Conversation, error)
	// Delete 删除会话。产品层确认后再调；实现层负责级联消息 / 运行记录。
	Delete(ctx context.Context, id string) error
}

// MessageService 会话消息。
// 追加消息会伴随一次 Run，等聊天闭环再补 Append；当前只先定读取，避免 handle 提前依赖还没有的写入路径。
type MessageService interface {
	// Get 按 ID 取单条消息。
	Get(ctx context.Context, id string) (*domain.Message, error)
	// List 按会话拉历史。打开会话、刷新恢复都走这条。
	List(ctx context.Context, in domain.ListMessagesInput) ([]domain.Message, error)
}

// RunService 一次提交对应一次运行。
// 状态流转（排队 → 执行中 → 终态）在实现里做；handle 只展示。同一会话同时只允许一个活跃运行。
type RunService interface {
	// Get 按 ID 取一次运行，给「运行详情」用。
	Get(ctx context.Context, id string) (*domain.Run, error)
	// List 按会话列出运行记录。
	List(ctx context.Context, in domain.ListRunsInput) ([]domain.Run, error)
}
