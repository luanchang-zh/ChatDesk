package service

import (
	"context"

	"github.com/luanchang-zh/ChatDesk/consts"
	"github.com/luanchang-zh/ChatDesk/internal/domain"
	"github.com/luanchang-zh/ChatDesk/pkg/apperr"
)

// unavailable 统一返回「服务暂不可用」。
//
// 必须用业务码，不能返回空切片：
// 前端拿到 items: [] 会以为存储已经接上、只是当前没有会话；
// 拿到 code=30002 才知道这是骨架占位，还不能当正式列表渲染。
func unavailable() error {
	return apperr.New(consts.CodeServiceUnavailable)
}

// 编译期检查：占位类型必须满足接口。
// sqlc 实现替换时，同样用这组断言——漏实现某个方法会在编译期而不是上线后暴露。
var (
	_ ProjectService      = UnavailableProjects{}
	_ ConversationService = UnavailableConversations{}
	_ MessageService      = UnavailableMessages{}
	_ RunService          = UnavailableRuns{}
)

// UnavailableProjects 项目服务占位。还没有表，所有读写都失败。
type UnavailableProjects struct{}

func (UnavailableProjects) Create(context.Context, domain.CreateProjectInput) (*domain.Project, error) {
	return nil, unavailable()
}

func (UnavailableProjects) Get(context.Context, string) (*domain.Project, error) {
	return nil, unavailable()
}

func (UnavailableProjects) List(context.Context, string) ([]domain.Project, error) {
	return nil, unavailable()
}

func (UnavailableProjects) Update(context.Context, domain.UpdateProjectInput) (*domain.Project, error) {
	return nil, unavailable()
}

func (UnavailableProjects) Delete(context.Context, string) error {
	return unavailable()
}

// UnavailableConversations 会话服务占位。
// GET /api/v1/conversations 当前会打到这里，handle 再把它映射成 HTTP 500 + code 30002。
type UnavailableConversations struct{}

func (UnavailableConversations) Create(context.Context, domain.CreateConversationInput) (*domain.Conversation, error) {
	return nil, unavailable()
}

func (UnavailableConversations) Get(context.Context, string) (*domain.Conversation, error) {
	return nil, unavailable()
}

func (UnavailableConversations) List(context.Context, domain.ListConversationsInput) ([]domain.Conversation, error) {
	return nil, unavailable()
}

func (UnavailableConversations) Rename(context.Context, domain.RenameConversationInput) (*domain.Conversation, error) {
	return nil, unavailable()
}

func (UnavailableConversations) Delete(context.Context, string) error {
	return unavailable()
}

// UnavailableMessages 消息服务占位。打开会话拉历史时会用到。
type UnavailableMessages struct{}

func (UnavailableMessages) Get(context.Context, string) (*domain.Message, error) {
	return nil, unavailable()
}

func (UnavailableMessages) List(context.Context, domain.ListMessagesInput) ([]domain.Message, error) {
	return nil, unavailable()
}

// UnavailableRuns 运行服务占位。运行详情、停止生成会用到。
type UnavailableRuns struct{}

func (UnavailableRuns) Get(context.Context, string) (*domain.Run, error) {
	return nil, unavailable()
}

func (UnavailableRuns) List(context.Context, domain.ListRunsInput) ([]domain.Run, error) {
	return nil, unavailable()
}
