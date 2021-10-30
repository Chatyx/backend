package service

import (
	"context"
	"fmt"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/internal/repository"
	"github.com/Mort4lis/scht-backend/pkg/logging"
)

type chatMemberService struct {
	userService    UserService
	chatService    ChatService
	chatMemberRepo repository.ChatMemberRepository
	messageRepo    repository.MessageRepository
	messagePubSub  repository.MessagePubSub
	logger         logging.Logger
}

type ChatMemberConfig struct {
	UserService    UserService
	ChatService    ChatService
	ChatMemberRepo repository.ChatMemberRepository
	MessageRepo    repository.MessageRepository
	MessagePubSub  repository.MessagePubSub
}

func NewChatMemberService(cfg ChatMemberConfig) ChatMemberService {
	return &chatMemberService{
		userService:    cfg.UserService,
		chatService:    cfg.ChatService,
		chatMemberRepo: cfg.ChatMemberRepo,
		messageRepo:    cfg.MessageRepo,
		messagePubSub:  cfg.MessagePubSub,
		logger:         logging.GetLogger(),
	}
}

func (s *chatMemberService) ListMembersInChat(ctx context.Context, chatID, userID string) ([]domain.ChatMember, error) {
	members, err := s.chatMemberRepo.ListMembersInChat(ctx, chatID)
	if err != nil {
		return nil, err
	}

	isIn := false

	for _, member := range members {
		if member.UserID == userID {
			isIn = true
			break
		}
	}

	if !isIn {
		s.logger.WithFields(logging.Fields{
			"user_id": userID,
			"chat_id": chatID,
		}).Debug("member isn't in this chat")

		return nil, domain.ErrChatNotFound
	}

	return members, nil
}

func (s *chatMemberService) IsMemberInChat(ctx context.Context, userID, chatID string) (bool, error) {
	return s.chatMemberRepo.IsMemberInChat(ctx, userID, chatID)
}

func (s *chatMemberService) JoinMemberToChat(ctx context.Context, chatID, creatorID, userID string) error {
	if _, err := s.chatService.GetOwnByID(ctx, chatID, creatorID); err != nil {
		return err
	}

	user, err := s.userService.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if err = s.chatMemberRepo.Create(ctx, userID, chatID); err != nil {
		return err
	}

	dto := domain.CreateMessageDTO{
		ActionID: domain.MessageJoinAction,
		Text:     fmt.Sprintf("%s successfully joined to the chat", user.Username),
		ChatID:   chatID,
		SenderID: userID,
	}

	message, err := s.messageRepo.Create(ctx, dto)
	if err != nil {
		return err
	}

	if err = s.messagePubSub.Publish(ctx, message); err != nil {
		return err
	}

	return nil
}
