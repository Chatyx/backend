package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Chatyx/backend/internal/dto"
	"github.com/Chatyx/backend/internal/entity"
	"github.com/Chatyx/backend/pkg/ctxutil"
)

type MessageRepository interface {
	List(ctx context.Context, obj dto.MessageList) ([]entity.Message, error)
	Create(ctx context.Context, message *entity.Message) error
}

type InChatChecker interface {
	Check(ctx context.Context, chatID entity.ChatID, userID int) error
}

type MessagePublisher interface {
	Publish(ctx context.Context, message entity.Message) error
}

type Message struct {
	repo      MessageRepository
	publisher MessagePublisher
	checker   InChatChecker
}

func NewMessage(repo MessageRepository, publisher MessagePublisher, checker InChatChecker) *Message {
	return &Message{
		repo:      repo,
		publisher: publisher,
		checker:   checker,
	}
}

func (s *Message) List(ctx context.Context, obj dto.MessageList) ([]entity.Message, error) {
	userID := ctxutil.UserIDFromContext(ctx).ToInt()
	if err := s.checker.Check(ctx, obj.ChatID, userID); err != nil {
		return nil, fmt.Errorf("check whether the current user is in the chat or not: %w", err)
	}

	messages, err := s.repo.List(ctx, obj)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	return messages, nil
}

func (s *Message) Create(ctx context.Context, obj dto.MessageCreate) (entity.Message, error) {
	userID := ctxutil.UserIDFromContext(ctx).ToInt()
	if err := s.checker.Check(ctx, obj.ChatID, userID); err != nil {
		return entity.Message{}, fmt.Errorf("check whether the current user is in the chat or not: %w", err)
	}

	message := entity.Message{
		ChatID:      obj.ChatID,
		SenderID:    userID,
		Content:     obj.Content,
		ContentType: obj.ContentType,
		SentAt:      time.Now(),
	}
	if err := s.repo.Create(ctx, &message); err != nil {
		return entity.Message{}, fmt.Errorf("create message: %w", err)
	}

	if err := s.publisher.Publish(ctx, message); err != nil {
		return entity.Message{}, fmt.Errorf("publish message: %w", err)
	}
	return message, nil
}
