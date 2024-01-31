package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Chatyx/backend/internal/entity"

	"github.com/redis/go-redis/v9"
)

type ParticipantEventProduceConsumer struct {
	cli *redis.Client
}

func NewParticipantEventProduceConsumer(cli *redis.Client) *ParticipantEventProduceConsumer {
	return &ParticipantEventProduceConsumer{cli: cli}
}

func (ps *ParticipantEventProduceConsumer) Produce(ctx context.Context, event entity.ParticipantEvent) error {
	model := newParticipantEventModel(event)

	bytes, err := json.Marshal(model)
	if err != nil {
		return fmt.Errorf("marshal participant event: %v", err)
	}

	channel := participantEventsChannelName(event.UserID)
	if err = ps.cli.Publish(ctx, channel, bytes).Err(); err != nil {
		return fmt.Errorf("produce participant event to channel: %v", err)
	}
	return nil
}

func (ps *ParticipantEventProduceConsumer) BeginConsume(ctx context.Context, userID int) (<-chan entity.ParticipantEvent, <-chan error) {
	outCh, errCh := make(chan entity.ParticipantEvent), make(chan error)

	go func() {
		defer close(outCh)
		defer close(errCh)

		channel := participantEventsChannelName(userID)
		pubSub := ps.cli.Subscribe(ctx, channel)
		defer pubSub.Close()

		ch := pubSub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}

				var model participantEventModel
				if err := json.Unmarshal([]byte(msg.Payload), &model); err != nil {
					errCh <- fmt.Errorf("unmarshal participant event: %v", err)
					continue
				}

				outCh <- model.ToEntity()
			}
		}
	}()

	return outCh, errCh
}

type participantEventModel struct {
	Type     entity.ParticipantEventType `json:"type"`
	ChatID   int                         `json:"chat_id"`
	ChatType entity.ChatType             `json:"chat_type"`
	UserID   int                         `json:"user_id"`
}

func newParticipantEventModel(event entity.ParticipantEvent) participantEventModel {
	return participantEventModel{
		Type:     event.Type,
		ChatID:   event.ChatID.ID,
		ChatType: event.ChatID.Type,
		UserID:   event.UserID,
	}
}

func (e participantEventModel) ToEntity() entity.ParticipantEvent {
	return entity.ParticipantEvent{
		Type: e.Type,
		ChatID: entity.ChatID{
			ID:   e.ChatID,
			Type: e.ChatType,
		},
		UserID: e.UserID,
	}
}

func participantEventsChannelName(userID int) string {
	return fmt.Sprintf("user:%d:participant_events", userID)
}
