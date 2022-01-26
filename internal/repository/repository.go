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
	GetByID(ctx context.Context, userID string) (domain.User, error)
	GetByUsername(ctx context.Context, username string) (domain.User, error)
	Update(ctx context.Context, dto domain.UpdateUserDTO) (domain.User, error)
	UpdatePassword(ctx context.Context, userID, password string) error
	Delete(ctx context.Context, userID string) error
}

type SessionRepository interface {
	Get(ctx context.Context, refreshToken string) (domain.Session, error)
	Set(ctx context.Context, session domain.Session, ttl time.Duration) error
	Delete(ctx context.Context, refreshToken, userID string) error
	DeleteAllByUserID(ctx context.Context, userID string) error
}

type ChatRepository interface {
	List(ctx context.Context, userID string) ([]domain.Chat, error)
	Create(ctx context.Context, dto domain.CreateChatDTO) (domain.Chat, error)
	Get(ctx context.Context, memberKey domain.ChatMemberIdentity) (domain.Chat, error)
	Update(ctx context.Context, dto domain.UpdateChatDTO) (domain.Chat, error)
	Delete(ctx context.Context, memberKey domain.ChatMemberIdentity) error
}

type ChatMemberRepository interface {
	ListByChatID(ctx context.Context, chatID string) ([]domain.ChatMember, error)
	ListByUserID(ctx context.Context, userID string) ([]domain.ChatMember, error)
	GetByKey(ctx context.Context, memberKey domain.ChatMemberIdentity) (domain.ChatMember, error)
	IsInChat(ctx context.Context, memberKey domain.ChatMemberIdentity) (bool, error)
	IsChatCreator(ctx context.Context, memberKey domain.ChatMemberIdentity) (bool, error)
	Create(ctx context.Context, memberKey domain.ChatMemberIdentity) error
	Update(ctx context.Context, dto domain.UpdateChatMemberDTO) error
}

type MessageRepository interface {
	Create(ctx context.Context, dto domain.CreateMessageDTO) (domain.Message, error)
	List(ctx context.Context, chatID string, dto domain.MessageListDTO) (domain.MessageList, error)
}

type MessagePubSub interface {
	Publish(ctx context.Context, message domain.Message) error
	Subscribe(ctx context.Context, chatIDs ...string) MessageSubscriber
}

type MessageSubscriber interface {
	Subscribe(ctx context.Context, chatIDs ...string) error
	Unsubscribe(ctx context.Context, chatIDs ...string) error
	ReceiveMessage(ctx context.Context) (domain.Message, error)
	MessageChannel() (<-chan domain.Message, <-chan error)
	Close() error
}
