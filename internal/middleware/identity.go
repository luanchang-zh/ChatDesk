package middleware

import (
	"net"
	"net/netip"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/luanchang-zh/ChatDesk/consts"
	"github.com/luanchang-zh/ChatDesk/pkg/ctxmeta"
	"github.com/luanchang-zh/ChatDesk/pkg/result"
)

// DevIdentity 把服务端配置的开发用户注入请求。
//
// userID 为空或非法时，所有 /api/v1 业务接口直接未认证；健康检查不走这条中间件。
// 身份只来自启动配置，不读请求体、查询参数或 X-User-* Header。
func DevIdentity(userID string) gin.HandlerFunc {
	id, parseErr := uuid.Parse(userID)
	return func(c *gin.Context) {
		if parseErr != nil || id == uuid.Nil {
			result.Fail(c, nil, consts.CodeUnauthorized)
			c.Abort()
			return
		}
		// RemoteAddr 是直接连接对端。不信任 X-Forwarded-For，否则网页可以把身份伪装成本机。
		remoteHost, _, err := net.SplitHostPort(c.Request.RemoteAddr)
		remoteIP, ipErr := netip.ParseAddr(remoteHost)
		if err != nil || ipErr != nil || !remoteIP.IsLoopback() || !localHost(c.Request.Host) || !sameOrigin(c) {
			result.Fail(c, nil, consts.CodePermissionDeny)
			c.Abort()
			return
		}
		ctxmeta.SetUserUUID(c, id.String())
		c.Next()
	}
}

// localHost 只接受回环 Host。localhost 与 127.0.0.1 / ::1 可以，带 userinfo 或路径的 Host 不行。
func localHost(authority string) bool {
	u, err := url.Parse("http://" + authority)
	if err != nil || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
		return false
	}
	if strings.EqualFold(u.Hostname(), "localhost") {
		return true
	}
	ip, err := netip.ParseAddr(u.Hostname())
	return err == nil && ip.IsLoopback() && ip.Zone() == ""
}

// sameOrigin 挡住跨站浏览器请求借本机开发身份写数据。
// 没有 Origin 的 curl 允许；有 Origin 时必须与当前 Host 同源。
func sameOrigin(c *gin.Context) bool {
	site := c.GetHeader("Sec-Fetch-Site")
	if site != "" && site != "same-origin" && site != "none" {
		return false
	}
	origins := c.Request.Header.Values("Origin")
	if len(origins) == 0 {
		return true
	}
	if len(origins) != 1 {
		return false
	}
	origin, err := url.Parse(origins[0])
	if err != nil || origin.User != nil || origin.Path != "" || origin.RawQuery != "" || origin.Fragment != "" {
		return false
	}
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	return origin.Scheme == scheme && strings.EqualFold(origin.Host, c.Request.Host)
}
