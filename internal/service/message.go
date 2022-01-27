package service

import (
	"context"
	"fmt"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/internal/repository"
	pkgErrors "github.com/Mort4lis/scht-backend/pkg/errors"
)

type messageService struct {
	chatMemberRepo repository.ChatMemberRepository
	messageRepo    repository.MessageRepository
	pubSub         repository.MessagePubSub
}

func NewMessageService(chatMemberRepo repository.ChatMemberRepository, messageRepo repository.MessageRepository, pubSub repository.MessagePubSub) MessageService {
	return &messageService{
		chatMemberRepo: chatMemberRepo,
		messageRepo:    messageRepo,
		pubSub:         pubSub,
	}
}

func (s *messageService) NewServeSession(ctx context.Context, userID string) (chan<- domain.CreateMessageDTO, <-chan domain.Message, <-chan error, error) {
	members, err := s.chatMemberRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get list of chat members by user id: %w", err)
	}

	chatIDs := make([]string, 0, len(members))

	for _, member := range members {
		if member.IsInChat() {
			chatIDs = append(chatIDs, member.ChatID)
		}
	}

	inCh := make(chan domain.CreateMessageDTO)
	outCh := make(chan domain.Message)
	errCh := make(chan error)

	session := &messageServeSession{
		userID:         userID,
		messageService: s,
		chatMemberRepo: s.chatMemberRepo,
		subscriber:     s.pubSub.Subscribe(ctx, chatIDs...),
		inCh:           inCh,
		outCh:          outCh,
		errCh:          errCh,
	}
	go session.serve(ctx)

	return inCh, outCh, errCh, nil
}

func (s *messageService) Create(ctx context.Context, dto domain.CreateMessageDTO) (domain.Message, error) {
	ok, err := s.chatMemberRepo.IsInChat(ctx, domain.ChatMemberIdentity{
		UserID: dto.SenderID,
		ChatID: dto.ChatID,
	})
	if err != nil {
		return domain.Message{}, fmt.Errorf("failed to check if sender is in the chat: %w", err)
	}

	if !ok {
		return domain.Message{}, fmt.Errorf("member isn't in this chat: %w", domain.ErrChatNotFound)
	}

	message, err := s.messageRepo.Create(ctx, dto)
	if err != nil {
		return domain.Message{}, fmt.Errorf("failed to create message: %w", err)
	}

	if err = s.pubSub.Publish(ctx, message); err != nil {
		return domain.Message{}, fmt.Errorf("failed to publish message: %w", err)
	}

	return message, nil
}

func (s *messageService) List(ctx context.Context, memberKey domain.ChatMemberIdentity, dto domain.MessageListDTO) (domain.MessageList, error) {
	ok, err := s.chatMemberRepo.IsInChat(ctx, memberKey)
	if err != nil {
		return domain.MessageList{}, fmt.Errorf("failed to check if member is in the chat: %w", err)
	}

	if !ok {
		return domain.MessageList{}, fmt.Errorf("member isn't in this chat: %w", domain.ErrChatNotFound)
	}

	messages, err := s.messageRepo.List(ctx, memberKey.ChatID, dto)
	if err != nil {
		return domain.MessageList{}, fmt.Errorf("failed to get list of messages: %w", err)
	}

	return messages, nil
}

type messageServeSession struct {
	userID         string
	messageService MessageService
	chatMemberRepo repository.ChatMemberRepository
	subscriber     repository.MessageSubscriber

	inCh  chan domain.CreateMessageDTO
	outCh chan domain.Message
	errCh chan error
}

func (s *messageServeSession) serve(ctx context.Context) {
	defer close(s.errCh)
	defer close(s.outCh)
	defer s.subscriber.Close()

	subCh, subErrCh := s.subscriber.MessageChannel()

	for {
		select {
		case dto, ok := <-s.inCh:
			if !ok {
				return
			}

			if _, err := s.messageService.Create(ctx, dto); err != nil {
				s.errCh <- fmt.Errorf("failed to create message: %w", err)
				return
			}
		case message, ok := <-subCh:
			if !ok {
				return
			}

			ok, err := s.handleMessage(ctx, message)
			if err != nil {
				s.errCh <- fmt.Errorf("failed to handle message: %w", err)
				return
			}

			if !ok {
				continue
			}

			s.outCh <- message
		case err := <-subErrCh:
			s.errCh <- fmt.Errorf("failed to receive message from subscriber: %w", err)
			return
		}
	}
}

func (s *messageServeSession) handleMessage(ctx context.Context, message domain.Message) (ok bool, err error) {
	ctxFields := pkgErrors.ContextFields{
		"message_id": message.ID,
		"action_id":  message.ActionID,
		"chat_id":    message.ChatID,
		"sender_id":  message.SenderID,
	}

	switch message.ActionID {
	case domain.MessageSendAction:
		if message.SenderID == s.userID {
			return false, nil
		}
	case domain.MessageJoinAction:
		ok, err = s.chatMemberRepo.IsInChat(ctx, domain.ChatMemberIdentity{
			UserID: s.userID,
			ChatID: message.ChatID,
		})
		if err != nil {
			return false, pkgErrors.WrapInContextError(err, "failed to check if authenticated user is in the chat", ctxFields)
		}

		if !ok {
			return false, nil
		}

		if message.SenderID == s.userID {
			if err = s.subscriber.Subscribe(ctx, message.ChatID); err != nil {
				return false, pkgErrors.WrapInContextError(err, "failed to subscribe to chat", ctxFields)
			}
		}
	case domain.MessageLeaveAction, domain.MessageKickAction:
		if message.SenderID == s.userID {
			if err = s.subscriber.Unsubscribe(ctx, message.ChatID); err != nil {
				return false, pkgErrors.WrapInContextError(err, "failed to unsubscribe from chat", ctxFields)
			}
		}
	default:
		return false, pkgErrors.ContextError{
			Message: "unknown action_id for handling the message",
			Fields:  ctxFields,
		}
	}

	return true, nil
}
