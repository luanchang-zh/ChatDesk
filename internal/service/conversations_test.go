package service

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/luanchang-zh/ChatDesk/consts"
	"github.com/luanchang-zh/ChatDesk/internal/domain"
	"github.com/luanchang-zh/ChatDesk/pkg/apperr"
	"github.com/luanchang-zh/ChatDesk/pkg/ctxmeta"
)

type conversationStoreSpy struct {
	owner uuid.UUID
	id    uuid.UUID
	title string
}

func (s *conversationStoreSpy) Create(_ context.Context, owner uuid.UUID, title string) (*domain.Conversation, error) {
	s.owner, s.title = owner, title
	return &domain.Conversation{UserID: owner.String(), Title: title}, nil
}
func (s *conversationStoreSpy) Get(_ context.Context, owner, id uuid.UUID) (*domain.Conversation, error) {
	s.owner, s.id = owner, id
	return &domain.Conversation{ID: id.String(), UserID: owner.String()}, nil
}
func (s *conversationStoreSpy) List(_ context.Context, owner uuid.UUID) ([]domain.Conversation, error) {
	s.owner = owner
	return []domain.Conversation{}, nil
}
func (s *conversationStoreSpy) Rename(_ context.Context, owner, id uuid.UUID, title string) (*domain.Conversation, error) {
	s.owner, s.id, s.title = owner, id, title
	return &domain.Conversation{ID: id.String(), UserID: owner.String(), Title: title}, nil
}
func (s *conversationStoreSpy) Delete(_ context.Context, owner, id uuid.UUID) error {
	s.owner, s.id = owner, id
	return nil
}

func TestConversationOwnershipIsRequiredForEveryOperation(t *testing.T) {
	owner, id := uuid.New(), uuid.New()
	operations := map[string]func(*Conversations, context.Context) error{
		"create": func(s *Conversations, ctx context.Context) error {
			_, err := s.Create(ctx, domain.CreateConversationInput{})
			return err
		},
		"list": func(s *Conversations, ctx context.Context) error {
			_, err := s.List(ctx, domain.ListConversationsInput{})
			return err
		},
		"get": func(s *Conversations, ctx context.Context) error { _, err := s.Get(ctx, id.String()); return err },
		"rename": func(s *Conversations, ctx context.Context) error {
			_, err := s.Rename(ctx, domain.RenameConversationInput{ID: id.String(), Title: "改名"})
			return err
		},
		"delete": func(s *Conversations, ctx context.Context) error { return s.Delete(ctx, id.String()) },
	}
	for name, operation := range operations {
		t.Run(name, func(t *testing.T) {
			for _, identity := range []string{"", "forged", uuid.Nil.String()} {
				ctx := ctxmeta.WithUserUUID(context.Background(), identity)
				if err := operation(NewConversations(nil), ctx); apperr.Code(err) != consts.CodeUnauthorized {
					t.Fatalf("unauthenticated operation returned %v", err)
				}
			}
			spy := &conversationStoreSpy{}
			if err := operation(NewConversations(spy), ctxmeta.WithUserUUID(context.Background(), owner.String())); err != nil {
				t.Fatal(err)
			}
			if spy.owner != owner {
				t.Fatalf("owner not passed to storage: %s", spy.owner)
			}
		})
	}
}

func TestConversationInputValidation(t *testing.T) {
	ctx := ctxmeta.WithUserUUID(context.Background(), uuid.NewString())
	svc := NewConversations(nil)
	for _, id := range []string{"", "not-a-uuid", uuid.Nil.String()} {
		if _, err := svc.Get(ctx, id); apperr.Code(err) != consts.CodeParamError {
			t.Fatalf("Get(%q): %v", id, err)
		}
		if err := svc.Delete(ctx, id); apperr.Code(err) != consts.CodeParamError {
			t.Fatalf("Delete(%q): %v", id, err)
		}
	}
	for _, title := range []string{" ", strings.Repeat("中", 201), "contains\x00nul"} {
		if _, err := svc.Rename(ctx, domain.RenameConversationInput{ID: uuid.NewString(), Title: title}); apperr.Code(err) != consts.CodeParamError {
			t.Fatalf("Rename title: %v", err)
		}
	}
	if _, err := svc.Create(ctx, domain.CreateConversationInput{Title: strings.Repeat("中", 201)}); apperr.Code(err) != consts.CodeParamError {
		t.Fatalf("Create title: %v", err)
	}
	if _, err := svc.Create(ctx, domain.CreateConversationInput{ProjectID: uuid.NewString()}); apperr.Code(err) != consts.CodeParamError {
		t.Fatalf("Create project: %v", err)
	}
	if _, err := svc.List(ctx, domain.ListConversationsInput{ProjectID: uuid.NewString()}); apperr.Code(err) != consts.CodeParamError {
		t.Fatalf("List project: %v", err)
	}
	spy := &conversationStoreSpy{}
	for _, tc := range []struct{ input, want string }{{"  ", "新会话"}, {"  标题  ", "标题"}, {strings.Repeat("中", 200), strings.Repeat("中", 200)}} {
		item, err := NewConversations(spy).Create(ctx, domain.CreateConversationInput{Title: tc.input})
		if err != nil || item.Title != tc.want {
			t.Fatalf("create: item=%v error=%v", item, err)
		}
	}
}
