package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Chatyx/backend/internal/entity"
	"github.com/Chatyx/backend/internal/service"

	"github.com/redis/go-redis/v9"
)

type MessagePublishSubscriber struct {
	cli *redis.Client
}

func NewMessagePublishSubscriber(cli *redis.Client) *MessagePublishSubscriber {
	return &MessagePublishSubscriber{cli: cli}
}

func (ps *MessagePublishSubscriber) Publish(ctx context.Context, message entity.Message) error {
	model := newMessageModel(message)

	bytes, err := json.Marshal(model)
	if err != nil {
		return fmt.Errorf("marshal message: %v", err)
	}

	channel := chatChannelName(message.ChatID)
	if err = ps.cli.Publish(ctx, channel, bytes).Err(); err != nil {
		return fmt.Errorf("publish message to channel: %v", err)
	}
	return nil
}

//nolint:ireturn // that's a factory
func (ps *MessagePublishSubscriber) Subscribe(ctx context.Context, chatIDs ...entity.ChatID) service.MessageConsumer {
	channels := chatChannelNames(chatIDs...)
	return &MessageConsumer{
		pubSub: ps.cli.Subscribe(ctx, channels...),
	}
}

type MessageConsumer struct {
	pubSub *redis.PubSub
}

func (c *MessageConsumer) BeginConsume(ctx context.Context) (<-chan entity.Message, <-chan error) {
	outCh, errCh := make(chan entity.Message), make(chan error)

	go func() {
		defer close(outCh)
		defer close(errCh)

		ch := c.pubSub.Channel()

		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}

				var model messageModel
				if err := json.Unmarshal([]byte(msg.Payload), &model); err != nil {
					errCh <- fmt.Errorf("unmarshal message: %v", err)
					continue
				}

				outCh <- model.ToEntity()
			}
		}
	}()

	return outCh, errCh
}

func (c *MessageConsumer) Subscribe(ctx context.Context, chatIDs ...entity.ChatID) error {
	channels := chatChannelNames(chatIDs...)
	if err := c.pubSub.Subscribe(ctx, channels...); err != nil {
		return fmt.Errorf("subscribe to chat channels: %v", err)
	}
	return nil
}

func (c *MessageConsumer) Unsubscribe(ctx context.Context, chatIDs ...entity.ChatID) error {
	channels := chatChannelNames(chatIDs...)
	if err := c.pubSub.Unsubscribe(ctx, channels...); err != nil {
		return fmt.Errorf("unsubscribe from chat channels: %v", err)
	}
	return nil
}

func (c *MessageConsumer) Close() error {
	if err := c.pubSub.Close(); err != nil {
		return fmt.Errorf("close pubsub: %v", err)
	}
	return nil
}

type messageModel struct {
	ID          int             `json:"id"`
	ChatID      int             `json:"chat_id"`
	ChatType    entity.ChatType `json:"chat_type"`
	SenderID    int             `json:"sender_id"`
	Content     string          `json:"content"`
	ContentType string          `json:"content_type"`
	IsService   bool            `json:"is_service"`
	SentAt      time.Time       `json:"sent_at"`
	DeliveredAt *time.Time      `json:"delivered_at,omitempty"`
}

func newMessageModel(message entity.Message) messageModel {
	return messageModel{
		ID:          message.ID,
		ChatID:      message.ChatID.ID,
		ChatType:    message.ChatID.Type,
		SenderID:    message.SenderID,
		Content:     message.Content,
		ContentType: message.ContentType,
		IsService:   message.IsService,
		SentAt:      message.SentAt,
		DeliveredAt: message.DeliveredAt,
	}
}

func (m messageModel) ToEntity() entity.Message {
	return entity.Message{
		ID: m.ID,
		ChatID: entity.ChatID{
			ID:   m.ChatID,
			Type: m.ChatType,
		},
		SenderID:    m.SenderID,
		Content:     m.Content,
		ContentType: m.ContentType,
		IsService:   m.IsService,
		SentAt:      m.SentAt,
		DeliveredAt: m.DeliveredAt,
	}
}

func chatChannelName(chatID entity.ChatID) string {
	return fmt.Sprintf("%s:%d", chatID.Type, chatID.ID)
}

func chatChannelNames(chatIDs ...entity.ChatID) []string {
	channels := make([]string, len(chatIDs))
	for i, chatID := range chatIDs {
		channels[i] = chatChannelName(chatID)
	}
	return channels
}
