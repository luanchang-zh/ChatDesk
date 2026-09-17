package v1

// ListConversationsRequest 列会话查询参数。本轮只用来定绑定写法。
type ListConversationsRequest struct {
	ProjectID string `form:"project_id"`
}
