package service

import (
	"context"
	"time"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/internal/repository"
	"github.com/Mort4lis/scht-backend/pkg/logging"
)

type messageService struct {
	chatMemberRepo repository.ChatMemberRepository
	messageRepo    repository.MessageRepository
	pubSub         repository.MessagePubSub
	logger         logging.Logger
}

func NewMessageService(chatMemberRepo repository.ChatMemberRepository, messageRepo repository.MessageRepository, pubSub repository.MessagePubSub) MessageService {
	return &messageService{
		chatMemberRepo: chatMemberRepo,
		messageRepo:    messageRepo,
		pubSub:         pubSub,
		logger:         logging.GetLogger(),
	}
}

func (s *messageService) NewServeSession(ctx context.Context, userID string) (chan<- domain.CreateMessageDTO, <-chan domain.Message, error) {
	members, err := s.chatMemberRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, nil, err
	}

	chatIDs := make([]string, 0, len(members))

	for _, member := range members {
		if member.IsInChat() {
			chatIDs = append(chatIDs, member.ChatID)
		}
	}

	inCh := make(chan domain.CreateMessageDTO)
	outCh := make(chan domain.Message)

	session := &messageServeSession{
		userID:         userID,
		messageService: s,
		chatMemberRepo: s.chatMemberRepo,
		subscriber:     s.pubSub.Subscribe(ctx, chatIDs...),
		inCh:           inCh,
		outCh:          outCh,
		logger:         s.logger,
	}
	go session.serve(ctx)

	return inCh, outCh, nil
}

func (s *messageService) Create(ctx context.Context, dto domain.CreateMessageDTO) (domain.Message, error) {
	ok, err := s.chatMemberRepo.IsInChat(ctx, domain.ChatMemberIdentity{
		UserID: dto.SenderID,
		ChatID: dto.ChatID,
	})
	if err != nil {
		return domain.Message{}, err
	}

	if !ok {
		s.logger.WithFields(logging.Fields{
			"user_id": dto.SenderID,
			"chat_id": dto.ChatID,
		}).Debug("member isn't in this chat")

		return domain.Message{}, domain.ErrChatNotFound
	}

	message, err := s.messageRepo.Create(ctx, dto)
	if err != nil {
		return domain.Message{}, err
	}

	if err = s.pubSub.Publish(ctx, message); err != nil {
		return domain.Message{}, err
	}

	return message, nil
}

func (s *messageService) List(ctx context.Context, memberKey domain.ChatMemberIdentity, timestamp time.Time) ([]domain.Message, error) {
	ok, err := s.chatMemberRepo.IsInChat(ctx, memberKey)
	if err != nil {
		return nil, err
	}

	if !ok {
		s.logger.WithFields(logging.Fields{
			"user_id": memberKey.UserID,
			"chat_id": memberKey.ChatID,
		}).Debug("member isn't in this chat")

		return nil, domain.ErrChatNotFound
	}

	return s.messageRepo.List(ctx, memberKey.ChatID, timestamp)
}

type messageServeSession struct {
	userID         string
	messageService MessageService
	chatMemberRepo repository.ChatMemberRepository
	subscriber     repository.MessageSubscriber
	logger         logging.Logger

	inCh  chan domain.CreateMessageDTO
	outCh chan domain.Message
}

func (s *messageServeSession) serve(ctx context.Context) {
	defer close(s.outCh)
	defer s.subscriber.Close()

	subCh := s.subscriber.MessageChannel()

	for {
		select {
		case dto, ok := <-s.inCh:
			if !ok {
				return
			}

			if _, err := s.messageService.Create(ctx, dto); err != nil {
				return
			}
		case message, ok := <-subCh:
			if !ok {
				return
			}

			ok, err := s.handleMessage(ctx, message)
			if err != nil {
				return
			}

			if !ok {
				continue
			}

			s.outCh <- message
		}
	}
}

func (s *messageServeSession) handleMessage(ctx context.Context, message domain.Message) (ok bool, err error) {
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
			return false, err
		}

		if !ok {
			return false, nil
		}

		if message.SenderID == s.userID {
			if err = s.subscriber.Subscribe(ctx, message.ChatID); err != nil {
				return false, err
			}
		}
	case domain.MessageLeaveAction, domain.MessageKickAction:
		if message.SenderID == s.userID {
			if err = s.subscriber.Unsubscribe(ctx, message.ChatID); err != nil {
				return false, err
			}
		}
	default:
		s.logger.WithFields(logging.Fields{
			"action_id": message.ActionID,
		}).Error("Unknown action_id for handling the message")

		return false, nil
	}

	return true, nil
}
