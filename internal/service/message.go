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

type ParticipantEventConsumer interface {
	BeginConsume(ctx context.Context, userID int) (<-chan entity.ParticipantEvent, <-chan error)
}

type MessageConsumer interface {
	BeginConsume(ctx context.Context) (<-chan entity.Message, <-chan error)
	Subscribe(ctx context.Context, chatIDs ...entity.ChatID) error
	Unsubscribe(ctx context.Context, chatIDs ...entity.ChatID) error
	Close() error
}

type MessageSubscriber interface {
	Subscribe(ctx context.Context, chatIDs ...entity.ChatID) MessageConsumer
}

type MessageServeManagerConfig struct {
	Service          *Message
	EventConsumer    ParticipantEventConsumer
	Subscriber       MessageSubscriber
	GroupRepository  GroupRepository
	DialogRepository DialogRepository
}

type MessageServeManager struct {
	msgSrv     *Message
	eventCons  ParticipantEventConsumer
	subscriber MessageSubscriber
	groupRepo  GroupRepository
	dialogRepo DialogRepository
}

func NewMessageServeManager(conf MessageServeManagerConfig) *MessageServeManager {
	return &MessageServeManager{
		msgSrv:     conf.Service,
		eventCons:  conf.EventConsumer,
		subscriber: conf.Subscriber,
		groupRepo:  conf.GroupRepository,
		dialogRepo: conf.DialogRepository,
	}
}

func (sm *MessageServeManager) BeginServe(ctx context.Context, inCh <-chan dto.MessageCreate) (chan<- entity.Message, <-chan error, error) {
	chatIDs, err := sm.listActiveChatIDs(ctx)
	if err != nil {
		return nil, nil, err
	}

	outCh, errCh := sm.serve(ctx, chatIDs, inCh)
	return outCh, errCh, nil
}

//nolint:lll // too long naming
func (sm *MessageServeManager) serve(ctx context.Context, chatIDs []entity.ChatID, inCh <-chan dto.MessageCreate) (chan entity.Message, chan error) {
	curUserID := ctxutil.UserIDFromContext(ctx).ToInt()
	outCh, errCh := make(chan entity.Message), make(chan error)

	go func() {
		defer close(outCh)
		defer close(errCh)

		msgCons := sm.subscriber.Subscribe(ctx, chatIDs...)
		defer msgCons.Close()

		msgCh, msgErrCh := msgCons.BeginConsume(ctx)
		eventCh, eventErrCh := sm.eventCons.BeginConsume(ctx, curUserID)

		for {
			select {
			case obj, ok := <-inCh:
				if !ok {
					return
				}

				if _, err := sm.msgSrv.Create(ctx, obj); err != nil {
					errCh <- err
				}
			case msg, ok := <-msgCh:
				if !ok {
					return
				}

				outCh <- msg
			case event, ok := <-eventCh:
				if !ok {
					return
				}

				if err := sm.applyEvent(ctx, msgCons, event); err != nil {
					errCh <- err
				}
			case err := <-msgErrCh:
				errCh <- err
			case err := <-eventErrCh:
				errCh <- err
			}
		}
	}()

	return outCh, errCh
}

func (sm *MessageServeManager) listActiveChatIDs(ctx context.Context) ([]entity.ChatID, error) {
	groupsCh, errCh := make(chan []entity.Group), make(chan error)
	go func() {
		groups, groupsErr := sm.groupRepo.List(ctx)
		if groupsErr != nil {
			errCh <- fmt.Errorf("list of groups: %w", groupsErr)
			return
		}

		groupsCh <- groups
	}()

	dialogs, err := sm.dialogRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list of dialogs: %w", err)
	}

	var groups []entity.Group
	select {
	case groups = <-groupsCh:
	case err = <-errCh:
		return nil, err
	}

	chatIDs := make([]entity.ChatID, 0, len(dialogs)+len(groups))
	for _, dialog := range dialogs {
		chatIDs = append(chatIDs, entity.ChatID{ID: dialog.ID, Type: entity.DialogChatType})
	}
	for _, group := range groups {
		chatIDs = append(chatIDs, entity.ChatID{ID: group.ID, Type: entity.GroupChatType})
	}

	return chatIDs, nil
}

func (sm *MessageServeManager) applyEvent(ctx context.Context, cons MessageConsumer, event entity.ParticipantEvent) error {
	switch event.Type {
	case entity.AddedParticipant:
		if err := cons.Subscribe(ctx, event.ChatID); err != nil {
			return fmt.Errorf("subscribe to chat: %w", err)
		}
	case entity.RemovedParticipant:
		if err := cons.Unsubscribe(ctx, event.ChatID); err != nil {
			return fmt.Errorf("subscribe from chat: %w", err)
		}
	}

	return nil
}
