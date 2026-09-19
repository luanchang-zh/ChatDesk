package router

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/luanchang-zh/ChatDesk/consts"
	"github.com/luanchang-zh/ChatDesk/internal/domain"
	v1 "github.com/luanchang-zh/ChatDesk/internal/router/v1"
	"github.com/luanchang-zh/ChatDesk/internal/service"
	"github.com/luanchang-zh/ChatDesk/pkg/apperr"
	"github.com/luanchang-zh/ChatDesk/pkg/ctxmeta"
	"github.com/luanchang-zh/ChatDesk/pkg/result"
)

const testOwner = "a525e8f9-1f6a-47b4-be50-c7ce5afde696"
const testConversationID = "395ee768-23fd-47c6-9c3f-b37e1d58fa63"

type conversationHandlerSpy struct {
	service.UnavailableConversations
	operation string
	owner     string
	title     string
	err       error
}

func (s *conversationHandlerSpy) record(ctx context.Context, operation string) (*domain.Conversation, error) {
	s.operation, s.owner = operation, ctxmeta.UserUUID(ctx)
	return &domain.Conversation{ID: testConversationID, UserID: s.owner, Title: s.title, CreatedAt: time.Now(), UpdatedAt: time.Now()}, s.err
}
func (s *conversationHandlerSpy) Create(ctx context.Context, in domain.CreateConversationInput) (*domain.Conversation, error) {
	s.title = in.Title
	return s.record(ctx, "create")
}
func (s *conversationHandlerSpy) Get(ctx context.Context, _ string) (*domain.Conversation, error) {
	return s.record(ctx, "get")
}
func (s *conversationHandlerSpy) Rename(ctx context.Context, in domain.RenameConversationInput) (*domain.Conversation, error) {
	s.title = in.Title
	return s.record(ctx, "rename")
}
func (s *conversationHandlerSpy) Delete(ctx context.Context, _ string) error {
	_, err := s.record(ctx, "delete")
	return err
}
func (s *conversationHandlerSpy) List(ctx context.Context, _ domain.ListConversationsInput) ([]domain.Conversation, error) {
	_, err := s.record(ctx, "list")
	return nil, err
}

func TestConversationRoutes(t *testing.T) {
	cases := []struct{ method, path, body, operation string }{
		{"POST", "/api/v1/conversations", `{"title":"标题","user_id":"forged"}`, "create"},
		{"GET", "/api/v1/conversations?user_id=forged", "", "list"},
		{"GET", "/api/v1/conversations/" + testConversationID, "", "get"},
		{"PATCH", "/api/v1/conversations/" + testConversationID, `{"title":"新标题"}`, "rename"},
		{"DELETE", "/api/v1/conversations/" + testConversationID, "", "delete"},
	}
	for _, tc := range cases {
		t.Run(tc.operation, func(t *testing.T) {
			spy := &conversationHandlerSpy{}
			engine := NewEngine(v1.NewHealthHandler(), v1.NewConversationHandler(spy), testOwner)
			req := localRequest(tc.method, tc.path, tc.body)
			req.Header.Set("X-User-ID", "forged")
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, req)
			var response result.Response
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			if w.Code != 200 || response.Code != 0 || response.TraceId == "" || response.Timestamp <= 0 || spy.owner != testOwner || spy.operation != tc.operation {
				t.Fatalf("unexpected response %s, owner=%s operation=%s", w.Body.String(), spy.owner, spy.operation)
			}
			if strings.Contains(w.Body.String(), testOwner) {
				t.Fatal("owner leaked into DTO")
			}
			if tc.operation == "list" && !strings.Contains(w.Body.String(), `"items":[]`) {
				t.Fatalf("empty list is not []: %s", w.Body.String())
			}
		})
	}
}

func TestConversationErrorsAndDisabledIdentity(t *testing.T) {
	cases := []struct {
		name, method, path, body, identity string
		err                                error
		wantCode, wantHTTP                 int
	}{
		{name: "health stays public", method: "GET", path: "/health", wantHTTP: 200},
		{name: "identity disabled", method: "GET", path: "/api/v1/conversations", wantHTTP: 200, wantCode: consts.CodeUnauthorized},
		{name: "malformed JSON", method: "POST", path: "/api/v1/conversations", body: "{", identity: testOwner, wantHTTP: 200, wantCode: consts.CodeParamError},
		{name: "oversized body", method: "POST", path: "/api/v1/conversations", body: `{"title":"` + strings.Repeat("x", 70<<10) + `"}`, identity: testOwner, wantHTTP: 200, wantCode: consts.CodeParamError},
		{name: "hidden resource", method: "GET", path: "/api/v1/conversations/" + testConversationID, identity: testOwner, err: apperr.New(consts.CodeResourceNotFound), wantHTTP: 200, wantCode: consts.CodeResourceNotFound},
		{name: "server error is sanitized", method: "GET", path: "/api/v1/conversations", identity: testOwner, err: apperr.New(consts.CodeInternalError), wantHTTP: 500, wantCode: consts.CodeInternalError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			spy := &conversationHandlerSpy{err: tc.err}
			engine := NewEngine(v1.NewHealthHandler(), v1.NewConversationHandler(spy), tc.identity)
			w := httptest.NewRecorder()
			engine.ServeHTTP(w, localRequest(tc.method, tc.path, tc.body))
			var response result.Response
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			if w.Code != tc.wantHTTP || response.Code != tc.wantCode {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
		})
	}
}

func localRequest(method, path, body string) *http.Request {
	req := httptest.NewRequest(method, "http://127.0.0.1:8080"+path, strings.NewReader(body))
	req.RemoteAddr = "127.0.0.1:9000"
	req.Header.Set("Content-Type", "application/json")
	return req
}
