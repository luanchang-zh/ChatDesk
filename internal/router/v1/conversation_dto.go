package v1

// ListConversationsRequest 列会话的查询参数。
//
// 只用 form tag，不用 json tag：这个接口是 GET，参数在 query string 里，
// handle 用 ShouldBindQuery 绑定。绑失败时返回 CodeParamError，不打 Error 日志。
//
// 当前只支持空 ProjectID，项目管理将在后续工作包实现。
// 空值是合法输入，不要在 handle 里把它当成参数错误。
type ListConversationsRequest struct {
	ProjectID string `form:"project_id"`
}

// CreateConversationRequest 新建会话的 JSON 体。
// Title 可省略；非空 project_id 由 service 拒绝，handle 不做静默丢弃。
type CreateConversationRequest struct {
	Title     string `json:"title"`
	ProjectID string `json:"project_id"`
}

// RenameConversationRequest 重命名的 JSON 体。空白标题在 service 里判参数错误。
type RenameConversationRequest struct {
	Title string `json:"title"`
}
