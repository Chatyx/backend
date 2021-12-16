package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/internal/encoding"
	"github.com/go-redis/redis/v8"
)

type messageRedisRepository struct {
	redisClient *redis.Client
}

func NewMessageRedisRepository(redisClient *redis.Client) MessageRepository {
	return &messageRedisRepository{redisClient: redisClient}
}

func (r *messageRedisRepository) Create(ctx context.Context, dto domain.CreateMessageDTO) (domain.Message, error) {
	createdAt := time.Now()
	message := domain.Message{
		ID:        uuid.New().String(),
		ActionID:  dto.ActionID,
		Text:      dto.Text,
		ChatID:    dto.ChatID,
		SenderID:  dto.SenderID,
		CreatedAt: &createdAt,
	}

	payload, err := encoding.NewProtobufMessageMarshaler(message).Marshal()
	if err != nil {
		return domain.Message{}, fmt.Errorf("an error occurred while marshaling the message: %v", err)
	}

	key := fmt.Sprintf("chat:%s:messages", message.ChatID)
	if err = r.redisClient.ZAdd(ctx, key, &redis.Z{
		Score:  float64(message.CreatedAt.UnixNano()),
		Member: payload,
	}).Err(); err != nil {
		return domain.Message{}, fmt.Errorf("an error occurred while creating the message into the cache: %v", err)
	}

	return message, nil
}

func (r *messageRedisRepository) List(ctx context.Context, chatID string, timestamp time.Time) ([]domain.Message, error) {
	key := fmt.Sprintf("chat:%s:messages", chatID)

	payloads, err := r.redisClient.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min: fmt.Sprintf("(%d", timestamp.UnixNano()),
		Max: "+inf",
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("an error occurred while getting list of messages: %v", err)
	}

	messages := make([]domain.Message, 0, len(payloads))

	for _, payload := range payloads {
		var message domain.Message

		if err = encoding.NewProtobufMessageUnmarshaler(&message).Unmarshal([]byte(payload)); err != nil {
			return nil, fmt.Errorf("an error occurred while unmarshaling the message: %v", err)
		}

		messages = append(messages, message)
	}

	return messages, nil
}
