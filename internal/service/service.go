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
	GetByID(ctx context.Context, id string) (domain.User, error)
	GetByUsername(ctx context.Context, username string) (domain.User, error)
	Update(ctx context.Context, dto domain.UpdateUserDTO) (domain.User, error)
	UpdatePassword(ctx context.Context, dto domain.UpdateUserPasswordDTO) error
	Delete(ctx context.Context, id string) error
}

type AuthService interface {
	SignIn(ctx context.Context, dto domain.SignInDTO) (domain.JWTPair, error)
	Refresh(ctx context.Context, dto domain.RefreshSessionDTO) (domain.JWTPair, error)
	Authorize(accessToken string) (domain.Claims, error)
}

type ChatService interface {
	List(ctx context.Context, memberID string) ([]domain.Chat, error)
	Create(ctx context.Context, dto domain.CreateChatDTO) (domain.Chat, error)
	GetByID(ctx context.Context, chatID, memberID string) (domain.Chat, error)
	Update(ctx context.Context, dto domain.UpdateChatDTO) (domain.Chat, error)
	Delete(ctx context.Context, chatID, creatorID string) error
}

type MessageService interface {
	NewServeSession(ctx context.Context, userID string) (inCh chan<- domain.CreateMessageDTO, outCh <-chan domain.Message, errCh <-chan error)
	Create(ctx context.Context, senderID string, dto domain.CreateMessageDTO) (domain.Message, error)
	List(ctx context.Context, chatID, userID string, timestamp time.Time) ([]domain.Message, error)
}

type ServiceContainer struct {
	User    UserService
	Chat    ChatService
	Message MessageService
	Auth    AuthService
}
