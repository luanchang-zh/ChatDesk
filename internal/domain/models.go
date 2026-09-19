package domain

import "time"

// 本包只放业务含义上的对象，不依赖 Gin、数据库驱动、json tag。
// HTTP 层准备 DTO；仓储层负责映射 sqlc 行。
// 不要在这里加 form / json tag：线格式一变，领域对象不该跟着变。

// Project 一次工作的范围。
// 产品上它绑定知识库、默认模型和一批会话；当前骨架先保住归属和说明字段。
type Project struct {
	ID          string
	UserID      string // 数据归属，鉴权接上后所有查询都要带它
	Name        string
	Description string // 任务背景 / 回答偏好，影响后续 Run，不能改写历史回答
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CreateProjectInput 创建项目时 handle 转进来的用例入参。
// 和 HTTP DTO 分开：handle 负责绑定，这里只描述「创建项目需要什么」。
type CreateProjectInput struct {
	UserID      string
	Name        string
	Description string
}

// UpdateProjectInput 只更新后续 Run 会读到的说明，不回头改已经完成的回答。
type UpdateProjectInput struct {
	ID          string
	Name        string
	Description string
}

// Conversation 一次问答线程。
// 刷新页面后要能按 ID 打开；删除前由产品层确认。当前不含消息分支。
type Conversation struct {
	ID        string
	UserID    string
	ProjectID string // 空表示尚未归入项目，不是非法值
	Title     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CreateConversationInput 新建会话。Title 可空，默认“新会话”；身份来自 Context。
type CreateConversationInput struct {
	ProjectID string
	Title     string
}

// ListConversationsInput 列表过滤。用户归属始终从可信 Context 读取。
type ListConversationsInput struct {
	ProjectID string
}

// RenameConversationInput 重命名，不改历史消息。
type RenameConversationInput struct {
	ID    string
	Title string
}

// Message 会话里的一条消息。
// Role / Content 的合法性由 service 解释；handle 不做业务拆解。
type Message struct {
	ID             string
	ConversationID string
	Role           string
	Content        string
	CreatedAt      time.Time
}

// ListMessagesInput 按会话拉消息。打开会话、刷新恢复都走这条。
type ListMessagesInput struct {
	ConversationID string
}

// RunStatus 一次提交对应一次运行。取值与产品需求里的运行状态对齐，
// 不要在 handle 里另造一套英文单词。
type RunStatus string

const (
	RunQueued     RunStatus = "queued"     // 排队
	RunRunning    RunStatus = "running"    // 执行中
	RunCancelling RunStatus = "cancelling" // 已点停止，下游还在收尾
	RunCompleted  RunStatus = "completed"  // 正常完成
	RunFailed     RunStatus = "failed"     // 失败，已生成片段不能当完整答案
	RunStopped    RunStatus = "stopped"    // 用户停止
	RunTimeout    RunStatus = "timeout"    // 超时
)

// Run 一次模型生成任务。
// 产品约束：同一会话同时只允许一个活跃运行；刷新不等于取消。
type Run struct {
	ID             string
	ConversationID string
	Status         RunStatus
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// ListRunsInput 按会话查看运行记录，给「运行详情」用。
type ListRunsInput struct {
	ConversationID string
}
