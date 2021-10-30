package service

import (
	"context"
	"time"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/internal/repository"
	"github.com/Mort4lis/scht-backend/pkg/logging"
)

type messageService struct {
	chatService       ChatService
	chatMemberService ChatMemberService
	messageRepo       repository.MessageRepository
	pubSub            repository.MessagePubSub
	logger            logging.Logger
}

func NewMessageService(chatService ChatService, chatMemberService ChatMemberService, messageRepo repository.MessageRepository, pubSub repository.MessagePubSub) MessageService {
	return &messageService{
		chatService:       chatService,
		chatMemberService: chatMemberService,
		messageRepo:       messageRepo,
		pubSub:            pubSub,
		logger:            logging.GetLogger(),
	}
}

func (s *messageService) NewServeSession(ctx context.Context, userID string) (chan<- domain.CreateMessageDTO, <-chan domain.Message) {
	inCh := make(chan domain.CreateMessageDTO)
	outCh := make(chan domain.Message)

	go func() {
		defer close(outCh)

		userChats, err := s.chatService.List(ctx, userID)
		if err != nil {
			return
		}

		chatIDs := make([]string, 0, len(userChats))
		for _, userChat := range userChats {
			chatIDs = append(chatIDs, userChat.ID)
		}

		subscriber := s.pubSub.Subscribe(ctx, chatIDs...)
		defer subscriber.Close()

		subCh := subscriber.MessageChannel()

		for {
			select {
			case dto, ok := <-inCh:
				if !ok {
					return
				}

				if _, err = s.Create(ctx, dto); err != nil {
					return
				}
			case message, ok := <-subCh:
				if !ok {
					return
				}

				switch message.ActionID {
				case domain.MessageSendAction:
					if message.SenderID == userID {
						continue
					}
				case domain.MessageJoinAction:
					ok, err = s.chatMemberService.IsMemberInChat(ctx, userID, message.ChatID)
					if err != nil {
						return
					}

					if !ok {
						continue
					}

					if message.SenderID == userID {
						if err = subscriber.Subscribe(ctx, message.ChatID); err != nil {
							return
						}
					}
				case domain.MessageLeaveAction, domain.MessageBlockAction:
					if message.SenderID == userID {
						if err = subscriber.Unsubscribe(ctx, message.ChatID); err != nil {
							return
						}
					}
				default:
					s.logger.WithFields(logging.Fields{
						"action_id": message.ActionID,
					}).Error("Unknown action_id for handling the message")

					return
				}

				outCh <- message
			}
		}
	}()

	return inCh, outCh
}

func (s *messageService) Create(ctx context.Context, dto domain.CreateMessageDTO) (domain.Message, error) {
	ok, err := s.chatMemberService.IsMemberInChat(ctx, dto.SenderID, dto.ChatID)
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

func (s *messageService) List(ctx context.Context, chatID, userID string, timestamp time.Time) ([]domain.Message, error) {
	ok, err := s.chatMemberService.IsMemberInChat(ctx, userID, chatID)
	if err != nil {
		return nil, err
	}

	if !ok {
		s.logger.WithFields(logging.Fields{
			"user_id": userID,
			"chat_id": chatID,
		}).Debug("member isn't in this chat")

		return nil, domain.ErrChatNotFound
	}

	return s.messageRepo.List(ctx, chatID, timestamp)
}
