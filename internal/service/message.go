package service

import (
	"context"
	"time"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/internal/repository"
	"github.com/Mort4lis/scht-backend/pkg/logging"
)

type messageService struct {
	chatService ChatService
	messageRepo repository.MessageRepository
	pubSub      repository.MessagePubSub
	logger      logging.Logger
}

func NewMessageService(chatService ChatService, messageRepo repository.MessageRepository, pubSub repository.MessagePubSub) MessageService {
	return &messageService{
		chatService: chatService,
		messageRepo: messageRepo,
		pubSub:      pubSub,
		logger:      logging.GetLogger(),
	}
}

func (s *messageService) NewServeSession(ctx context.Context, userID string) (chan<- domain.CreateMessageDTO, <-chan domain.Message, <-chan error) {
	errCh := make(chan error)
	inCh := make(chan domain.CreateMessageDTO)
	outCh := make(chan domain.Message)

	go func() {
		defer close(errCh)
		defer close(outCh)

		userChats, err := s.chatService.List(ctx, userID)
		if err != nil {
			errCh <- err
			return
		}

		chatIDs := make([]string, 0, len(userChats))
		for _, userChat := range userChats {
			chatIDs = append(chatIDs, userChat.ID)
		}

		subscriber := s.pubSub.Subscribe(ctx, chatIDs...)
		defer subscriber.Close()

		subCh := subscriber.MessageChannel(ctx)

	LOOP:
		for {
			select {
			case dto, ok := <-inCh:
				if !ok {
					break LOOP
				}

				if _, err = s.Create(ctx, dto); err != nil {
					errCh <- err
					return
				}
			case message, ok := <-subCh:
				if !ok {
					break LOOP
				}

				switch message.Action {
				case domain.MessageSendAction:
					if message.SenderID == userID {
						continue
					}
				case domain.MessageJoinAction:
				case domain.MessageLeaveAction:
				case domain.MessageBlockAction:
				}

				outCh <- message
			}
		}

		errCh <- nil
	}()

	return inCh, outCh, errCh
}

func (s *messageService) Create(ctx context.Context, dto domain.CreateMessageDTO) (domain.Message, error) {
	ok, err := s.chatService.IsBelongToChat(ctx, dto.ChatID, dto.SenderID)
	if err != nil {
		return domain.Message{}, err
	}

	if !ok {
		s.logger.WithFields(logging.Fields{
			"user_id": dto.SenderID,
			"chat_id": dto.ChatID,
		}).Debug("user doesn't belong to this chat")

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

func (s *messageService) List(ctx context.Context, chatID, userID string, timestamp time.Time) ([]domain.Message, error) {
	ok, err := s.chatService.IsBelongToChat(ctx, chatID, userID)
	if err != nil {
		return nil, err
	}

	if !ok {
		s.logger.WithFields(logging.Fields{
			"user_id": userID,
			"chat_id": chatID,
		}).Debug("user doesn't belong to this chat")

		return nil, domain.ErrChatNotFound
	}

	return s.messageRepo.List(ctx, chatID, timestamp)
}
