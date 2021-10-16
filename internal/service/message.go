package service

import (
	"context"
	"time"

	"github.com/Mort4lis/scht-backend/pkg/logging"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/internal/repository"
)

type messageService struct {
	chatService ChatService
	pubSub      repository.MessagePubSub
	logger      logging.Logger
}

func NewMessageService(chatService ChatService, pubSub repository.MessagePubSub) MessageService {
	return &messageService{
		chatService: chatService,
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

	/*
		TODO: at the beginning of the function:
		1.  check if user has access to this chat
		2.  store this message (redis)
	*/

	if err := s.pubSub.Publish(ctx, message, "chat:"+message.ChatID); err != nil {
		s.logger.WithError(err).Errorf("An error occurred while publishing the message to the chat (id=%s)", message.ChatID)
		return domain.Message{}, err
	}

	return message, nil
}
