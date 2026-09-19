package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/luanchang-zh/ChatDesk/consts"
	"github.com/luanchang-zh/ChatDesk/pkg/ctxmeta"
	"github.com/luanchang-zh/ChatDesk/pkg/result"
)

func TestDevIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const owner = "a525e8f9-1f6a-47b4-be50-c7ce5afde696"
	cases := []struct {
		name     string
		user     string
		remote   string
		host     string
		origin   string
		site     string
		wantCode int
	}{
		{name: "local command", user: owner},
		{name: "same origin browser", user: owner, origin: "http://127.0.0.1:8080", site: "same-origin"},
		{name: "IPv6", user: owner, remote: "[::1]:9000", host: "[::1]:8080", origin: "http://[::1]:8080"},
		{name: "disabled", wantCode: consts.CodeUnauthorized},
		{name: "invalid configured identity", user: "bad", wantCode: consts.CodeUnauthorized},
		{name: "external connection with forged forwarding", user: owner, remote: "192.0.2.1:9000", wantCode: consts.CodePermissionDeny},
		{name: "DNS rebinding host", user: owner, host: "attacker.example:8080", wantCode: consts.CodePermissionDeny},
		{name: "cross origin", user: owner, origin: "https://attacker.example", wantCode: consts.CodePermissionDeny},
		{name: "opaque origin", user: owner, origin: "null", wantCode: consts.CodePermissionDeny},
		{name: "other local port", user: owner, origin: "http://127.0.0.1:9999", wantCode: consts.CodePermissionDeny},
		{name: "cross site metadata", user: owner, site: "cross-site", wantCode: consts.CodePermissionDeny},
		{name: "same site is not same origin", user: owner, site: "same-site", wantCode: consts.CodePermissionDeny},
		{name: "userinfo host", user: owner, host: "attacker@localhost:8080", wantCode: consts.CodePermissionDeny},
		{name: "origin path", user: owner, origin: "http://127.0.0.1:8080/evil", wantCode: consts.CodePermissionDeny},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()
			router.Use(DevIdentity(tc.user))
			router.GET("/", func(c *gin.Context) {
				result.Success(c, ctxmeta.UserUUID(NewContextWithGin(c)))
			})
			req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:8080/?user_id=forged", nil)
			req.RemoteAddr = "127.0.0.1:9000"
			if tc.remote != "" {
				req.RemoteAddr = tc.remote
			}
			if tc.host != "" {
				req.Host = tc.host
			}
			if tc.origin != "" {
				req.Header.Set("Origin", tc.origin)
			}
			req.Header.Set("Sec-Fetch-Site", tc.site)
			req.Header.Set("X-User-ID", "forged")
			req.Header.Set("X-User-UUID", "forged")
			req.Header.Set("X-Forwarded-For", "127.0.0.1")
			req.Header.Set("X-Forwarded-Host", "127.0.0.1:8080")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			var response result.Response
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			if w.Code != http.StatusOK || response.Code != tc.wantCode {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
			if tc.wantCode == 0 && response.Data != owner {
				t.Fatalf("untrusted identity selected: %#v", response.Data)
			}
		})
	}
}
