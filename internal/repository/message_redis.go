package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/internal/encoding"
	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/go-redis/redis/v8"
)

type messageRedisRepository struct {
	redisClient *redis.Client
	logger      logging.Logger
}

func NewMessageRedisRepository(redisClient *redis.Client) MessageRepository {
	return &messageRedisRepository{
		redisClient: redisClient,
		logger:      logging.GetLogger(),
	}
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
		r.logger.WithError(err).Error("An error occurred while marshaling the message")
		return domain.Message{}, err
	}

	key := fmt.Sprintf("chat:%s:messages", message.ChatID)
	if err = r.redisClient.ZAdd(ctx, key, &redis.Z{
		Score:  float64(message.CreatedAt.UnixNano()),
		Member: payload,
	}).Err(); err != nil {
		r.logger.WithError(err).Error("An error occurred while creating the message into the redis")
		return domain.Message{}, err
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
		r.logger.WithError(err).Error("An error occurred while getting list of messages")
		return nil, err
	}

	messages := make([]domain.Message, 0, len(payloads))

	for _, payload := range payloads {
		var message domain.Message

		if err = encoding.NewProtobufMessageUnmarshaler(&message).Unmarshal([]byte(payload)); err != nil {
			r.logger.WithError(err).Error("An error occurred while unmarshal the message")
			return nil, err
		}

		messages = append(messages, message)
	}

	return messages, nil
}
