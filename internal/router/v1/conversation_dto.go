package v1

// ListConversationsRequest 列会话的查询参数。
//
// 只用 form tag，不用 json tag：这个接口是 GET，参数在 query string 里，
// handle 用 ShouldBindQuery 绑定。绑失败时返回 CodeParamError，不打 Error 日志。
//
// ProjectID 为空表示不按项目过滤，列出当前用户全部会话（鉴权接上后才真正按用户切）。
// 空值是合法输入，不要在 handle 里把它当成参数错误。
type ListConversationsRequest struct {
	ProjectID string `form:"project_id"`
}
