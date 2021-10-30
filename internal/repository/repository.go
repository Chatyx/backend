package repository

import (
	"context"
	"time"

	"github.com/Mort4lis/scht-backend/internal/domain"
)

//go:generate mockgen -source=repository.go -destination=mocks/mock.go

type UserRepository interface {
	List(ctx context.Context) ([]domain.User, error)
	Create(ctx context.Context, dto domain.CreateUserDTO) (domain.User, error)
	GetByID(ctx context.Context, id string) (domain.User, error)
	GetByUsername(ctx context.Context, username string) (domain.User, error)
	Update(ctx context.Context, dto domain.UpdateUserDTO) (domain.User, error)
	UpdatePassword(ctx context.Context, id, password string) error
	Delete(ctx context.Context, id string) error
}

type SessionRepository interface {
	Get(ctx context.Context, refreshToken string) (domain.Session, error)
	Set(ctx context.Context, session domain.Session, ttl time.Duration) error
	Delete(ctx context.Context, refreshToken, userID string) error
	DeleteAllByUserID(ctx context.Context, userID string) error
}

type ChatRepository interface {
	List(ctx context.Context, memberID string) ([]domain.Chat, error)
	Create(ctx context.Context, dto domain.CreateChatDTO) (domain.Chat, error)
	GetByID(ctx context.Context, chatID, memberID string) (domain.Chat, error)
	Update(ctx context.Context, dto domain.UpdateChatDTO) (domain.Chat, error)
	Delete(ctx context.Context, chatID, creatorID string) error
}

type ChatMemberRepository interface {
	ListMembersInChat(ctx context.Context, chatID string) ([]domain.ChatMember, error)
	IsMemberInChat(ctx context.Context, userID, chatID string) (bool, error)
}

type MessageRepository interface {
	Create(ctx context.Context, dto domain.CreateMessageDTO) (domain.Message, error)
	List(ctx context.Context, chatID string, timestamp time.Time) ([]domain.Message, error)
}

type MessagePubSub interface {
	Publish(ctx context.Context, message domain.Message) error
	Subscribe(ctx context.Context, chatIDs ...string) MessageSubscriber
}

type MessageSubscriber interface {
	Subscribe(ctx context.Context, chatIDs ...string) error
	Unsubscribe(ctx context.Context, chatIDs ...string) error
	ReceiveMessage(ctx context.Context) (domain.Message, error)
	MessageChannel() <-chan domain.Message
	Close() error
}
