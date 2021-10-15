package service

import (
	"context"

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

func (s *messageService) NewServeSession(ctx context.Context, userID string) (chan<- domain.Message, <-chan domain.Message, <-chan error) {
	inCh := make(chan domain.Message, 0)
	outCh := make(chan domain.Message, 0)
	errCh := make(chan error, 0)

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

		for {
			select {
			case inMsg, ok := <-inCh:
				if !ok {
					return
				}

				// TODO: check if user has access to this chat
				// TODO: store this message (redis)

				if err = s.pubSub.Publish(ctx, inMsg, "chat:"+inMsg.ChatID); err != nil {
					s.logger.WithError(err).Errorf("An error occurred while publishing the message to the chat (id=%s)", inMsg.ChatID)
					errCh <- err

					return
				}
			case recMsg, ok := <-subscriber.MessageChannel(ctx):
				if !ok {
					return
				}

				switch recMsg.Type {
				case domain.MessageSendType:
					if recMsg.SenderID == userID {
						continue
					}
				case domain.MessageJoinType:
				case domain.MessageLeaveType:
				case domain.MessageBlockType:
				}

				outCh <- recMsg
			}
		}
	}()

	return inCh, outCh, errCh
}
