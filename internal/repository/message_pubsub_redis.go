package repository

import (
	"context"
	"fmt"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/internal/encoding"
	"github.com/go-redis/redis/v8"
)

const broadcastTopic = "broadcast"

type messageRedisSubscriber struct {
	pubSub *redis.PubSub
}

func (s *messageRedisSubscriber) ReceiveMessage(ctx context.Context) (domain.Message, error) {
	msg, err := s.pubSub.ReceiveMessage(ctx)
	if err != nil {
		return domain.Message{}, fmt.Errorf("an error occurred while receiving message from pubSub: %v", err)
	}

	var message domain.Message
	if err = encoding.NewProtobufMessageUnmarshaler(&message).Unmarshal([]byte(msg.Payload)); err != nil {
		return domain.Message{}, fmt.Errorf("an error occurred while unmarshalling the message: %v", err)
	}

	return message, nil
}

func (s *messageRedisSubscriber) MessageChannel() (<-chan domain.Message, <-chan error) {
	ch := s.pubSub.Channel()
	msgCh := make(chan domain.Message)
	errCh := make(chan error)

	go func() {
		defer close(errCh)
		defer close(msgCh)

		for {
			msg, ok := <-ch
			if !ok {
				return
			}

			var message domain.Message
			if err := encoding.NewProtobufMessageUnmarshaler(&message).Unmarshal([]byte(msg.Payload)); err != nil {
				errCh <- fmt.Errorf("an error occurred while unmarshalling the message: %v", err)
				return
			}

			msgCh <- message
		}
	}()

	return msgCh, errCh
}

func (s *messageRedisSubscriber) Unsubscribe(ctx context.Context, chatIDs ...string) error {
	topics := getPubSubTopicsFromChatIDs(chatIDs...)

	if err := s.pubSub.Unsubscribe(ctx, topics...); err != nil {
		return fmt.Errorf("an error occurred while unsubscribing to topics: %v", err)
	}

	return nil
}

func (s *messageRedisSubscriber) Subscribe(ctx context.Context, chatIDs ...string) error {
	topics := getPubSubTopicsFromChatIDs(chatIDs...)

	if err := s.pubSub.Subscribe(ctx, topics...); err != nil {
		return fmt.Errorf("an error occurred while subscribing to topics: %v", err)
	}

	return nil
}

func (s *messageRedisSubscriber) Close() error {
	if err := s.pubSub.Close(); err != nil {
		return fmt.Errorf("an error occurred while closing pubSub subscriber: %v", err)
	}

	return nil
}

type messageRedisPubSub struct {
	redisClient *redis.Client
}

func NewMessagePubSub(redisClient *redis.Client) MessagePubSub {
	return &messageRedisPubSub{redisClient: redisClient}
}

func (ps *messageRedisPubSub) Publish(ctx context.Context, message domain.Message) error {
	payload, err := encoding.NewProtobufMessageMarshaler(message).Marshal()
	if err != nil {
		return fmt.Errorf("an error occurred while marshaling the message: %v", err)
	}

	var topic string
	if message.ActionID == domain.MessageJoinAction {
		topic = broadcastTopic
	} else {
		topic = getPubSubTopicFromChatID(message.ChatID)
	}

	if err = ps.redisClient.Publish(ctx, topic, payload).Err(); err != nil {
		return fmt.Errorf("an error occurred while publishing the message: %v", err)
	}

	return nil
}

func (ps *messageRedisPubSub) Subscribe(ctx context.Context, chatIDs ...string) MessageSubscriber {
	topics := getPubSubTopicsFromChatIDs(chatIDs...)
	topics = append(topics, broadcastTopic)

	return &messageRedisSubscriber{
		pubSub: ps.redisClient.Subscribe(ctx, topics...),
	}
}

func getPubSubTopicsFromChatIDs(chatIDs ...string) []string {
	topics := make([]string, 0, len(chatIDs)+1)
	for _, chatID := range chatIDs {
		topics = append(topics, getPubSubTopicFromChatID(chatID))
	}

	return topics
}

func getPubSubTopicFromChatID(chatID string) string {
	return "chat:" + chatID
}
