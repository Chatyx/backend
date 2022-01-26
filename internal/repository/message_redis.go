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

func (r *messageRedisRepository) List(ctx context.Context, chatID string, dto domain.MessageListDTO) (domain.MessageList, error) {
	min, max := "-inf", "+inf"
	key := fmt.Sprintf("chat:%s:messages", chatID)
	offsetDateUnixNano := fmt.Sprintf("%d", dto.OffsetDate.UnixNano())

	if dto.Direction == domain.NewerMessages {
		min = offsetDateUnixNano
	} else {
		max = offsetDateUnixNano
	}

	total, err := r.redisClient.ZCount(ctx, key, min, max).Result()
	if err != nil {
		return domain.MessageList{}, fmt.Errorf("an error occurred while calculating total messages: %v", err)
	}

	limit, offset := dto.Limit, dto.Offset
	if dto.Direction == domain.OlderMessages {
		offset = int(total) - dto.Limit - offset
		if offset < 0 {
			offset = 0
			limit = int(total) - dto.Offset
		}

		if limit <= 0 {
			return domain.MessageList{Total: int(total)}, nil
		}
	}

	payloads, err := r.redisClient.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min:    min,
		Max:    max,
		Offset: int64(offset),
		Count:  int64(limit),
	}).Result()
	if err != nil {
		return domain.MessageList{}, fmt.Errorf("an error occurred while getting list of messages: %v", err)
	}

	messages := make([]domain.Message, 0, len(payloads))

	for _, payload := range payloads {
		var message domain.Message

		if err = encoding.NewProtobufMessageUnmarshaler(&message).Unmarshal([]byte(payload)); err != nil {
			return domain.MessageList{}, fmt.Errorf("an error occurred while unmarshaling the message: %v", err)
		}

		messages = append(messages, message)
	}

	return domain.MessageList{
		Total:    int(total),
		Messages: messages,
	}, nil
}
