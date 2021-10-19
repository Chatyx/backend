package repository

import (
	"context"
	"fmt"

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

func (r *messageRedisRepository) Store(ctx context.Context, message domain.Message) error {
	payload, err := encoding.NewProtobufMessageMarshaler(message).Marshal()
	if err != nil {
		r.logger.WithError(err).Error("An error occurred while marshaling the message")
		return err
	}

	key := fmt.Sprintf("chat:%s:messages", message.ChatID)
	if err = r.redisClient.ZAdd(ctx, key, &redis.Z{
		Score:  float64(message.CreatedAt.Unix()),
		Member: payload,
	}).Err(); err != nil {
		r.logger.WithError(err).Error("An error occurred while storing the message into the redis")
		return err
	}

	return nil
}
