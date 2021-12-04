package service

import (
	"context"
	"time"

	"github.com/Mort4lis/scht-backend/internal/domain"
)

//go:generate mockgen -source=service.go -destination=mocks/mock.go

type UserService interface {
	List(ctx context.Context) ([]domain.User, error)
	Create(ctx context.Context, dto domain.CreateUserDTO) (domain.User, error)
	GetByID(ctx context.Context, userID string) (domain.User, error)
	GetByUsername(ctx context.Context, username string) (domain.User, error)
	Update(ctx context.Context, dto domain.UpdateUserDTO) (domain.User, error)
	UpdatePassword(ctx context.Context, dto domain.UpdateUserPasswordDTO) error
	Delete(ctx context.Context, userID string) error
}

type AuthService interface {
	SignIn(ctx context.Context, dto domain.SignInDTO) (domain.JWTPair, error)
	Refresh(ctx context.Context, dto domain.RefreshSessionDTO) (domain.JWTPair, error)
	Authorize(accessToken string) (domain.Claims, error)
}

type ChatService interface {
	// List gets a list of chats where user is a member.
	List(ctx context.Context, userID string) ([]domain.Chat, error)
	Create(ctx context.Context, dto domain.CreateChatDTO) (domain.Chat, error)
	// Get gets chat where user is a member.
	Get(ctx context.Context, memberKey domain.ChatMemberIdentity) (domain.Chat, error)
	Update(ctx context.Context, dto domain.UpdateChatDTO) (domain.Chat, error)
	// Delete deletes chat where user is a creator.
	Delete(ctx context.Context, memberKey domain.ChatMemberIdentity) error
}

type ChatMemberService interface {
	// List gets a list of chat members if accepted member consists in this chat.
	List(ctx context.Context, memberKey domain.ChatMemberIdentity) ([]domain.ChatMember, error)
	// GetByKey gets a chat member by its key where authenticated user is a member of this chat.
	GetByKey(ctx context.Context, memberKey domain.ChatMemberIdentity, user domain.AuthUser) (domain.ChatMember, error)
	// JoinToChat joins member to chat if authenticated user is a creator of this chat.
	JoinToChat(ctx context.Context, memberKey domain.ChatMemberIdentity, user domain.AuthUser) error
	// UpdateStatus updates status member if authenticated user is a chat creator or updatable member.
	UpdateStatus(ctx context.Context, dto domain.UpdateChatMemberDTO, user domain.AuthUser) error
}

type MessageService interface {
	NewServeSession(ctx context.Context, userID string) (inCh chan<- domain.CreateMessageDTO, outCh <-chan domain.Message, err error)
	Create(ctx context.Context, dto domain.CreateMessageDTO) (domain.Message, error)
	// List gets a list of chat messages with timestamp if accepted member consists in this chat.
	List(ctx context.Context, memberKey domain.ChatMemberIdentity, timestamp time.Time) ([]domain.Message, error)
}

type ServiceContainer struct {
	User       UserService
	Chat       ChatService
	ChatMember ChatMemberService
	Message    MessageService
	Auth       AuthService
}
