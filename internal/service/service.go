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
	// List gets a list of chats in which the authenticated user is a member.
	List(ctx context.Context, user domain.AuthUser) ([]domain.Chat, error)
	Create(ctx context.Context, dto domain.CreateChatDTO) (domain.Chat, error)
	// GetByID gets chat by id in which the authenticated user is a member.
	GetByID(ctx context.Context, chatID string, user domain.AuthUser) (domain.Chat, error)
	Update(ctx context.Context, dto domain.UpdateChatDTO) (domain.Chat, error)
	// Delete deletes chat by id in which the authenticated user is a creator.
	Delete(ctx context.Context, chatID string, user domain.AuthUser) error
}

type ChatMemberService interface {
	ListByUserID(ctx context.Context, userID string) ([]domain.ChatMember, error)
	ListByChatID(ctx context.Context, chatID, userID string) ([]domain.ChatMember, error)
	IsInChat(ctx context.Context, userID, chatID string) (bool, error)
	Get(ctx context.Context, chatID, userID string) (domain.ChatMember, error)
	GetAnother(ctx context.Context, authUserID, chatID, userID string) (domain.ChatMember, error)
	JoinToChat(ctx context.Context, chatID, creatorID, userID string) error
	UpdateStatus(ctx context.Context, dto domain.UpdateChatMemberDTO) error
	UpdateStatusByCreator(ctx context.Context, creatorID string, dto domain.UpdateChatMemberDTO) error
}

type MessageService interface {
	NewServeSession(ctx context.Context, userID string) (inCh chan<- domain.CreateMessageDTO, outCh <-chan domain.Message, err error)
	Create(ctx context.Context, dto domain.CreateMessageDTO) (domain.Message, error)
	List(ctx context.Context, chatID, userID string, timestamp time.Time) ([]domain.Message, error)
}

type ServiceContainer struct {
	User       UserService
	Chat       ChatService
	ChatMember ChatMemberService
	Message    MessageService
	Auth       AuthService
}
