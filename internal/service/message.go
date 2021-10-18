package service

import (
	"context"
	"fmt"
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

		topics := make([]string, 0, len(userChats))
		for _, userChat := range userChats {
			topics = append(topics, "chat:"+userChat.ID)
		}

		subscriber := s.pubSub.Subscribe(ctx, topics...)
		defer subscriber.Close()

		subCh := subscriber.MessageChannel(ctx)

	LOOP:
		for {
			select {
			case dto, ok := <-inCh:
				if !ok {
					break LOOP
				}

				if _, err = s.Create(ctx, userID, dto); err != nil {
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

func (s *messageService) Create(ctx context.Context, senderID string, dto domain.CreateMessageDTO) (domain.Message, error) {
	curTime := time.Now()
	message := domain.Message{
		Text:      dto.Text,
		ChatID:    dto.ChatID,
		SenderID:  senderID,
		CreatedAt: &curTime,
	}

	// TODO: check if user has access to this chat

	key := fmt.Sprintf("chat:%s:messages", dto.ChatID)
	if err := s.messageRepo.Store(ctx, key, message); err != nil {
		return domain.Message{}, err
	}

	if err := s.pubSub.Publish(ctx, message, "chat:"+message.ChatID); err != nil {
		return domain.Message{}, err
	}

	return message, nil
}
