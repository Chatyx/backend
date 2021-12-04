package repository

import (
	"context"
	"fmt"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/go-redis/redis/v8"
)

type chatCacheRepositoryDecorator struct {
	ChatRepository
	redisClient *redis.Client
	logger      logging.Logger
}

func NewChatCacheRepositoryDecorator(repo ChatRepository, redisClient *redis.Client) ChatRepository {
	return &chatCacheRepositoryDecorator{
		ChatRepository: repo,
		redisClient:    redisClient,
		logger:         logging.GetLogger(),
	}
}

func (r *chatCacheRepositoryDecorator) Delete(ctx context.Context, memberKey domain.ChatMemberIdentity) error {
	if err := r.ChatRepository.Delete(ctx, memberKey); err != nil {
		return err
	}

	chatUsersKey := fmt.Sprintf("chat:%s:user_ids", memberKey.ChatID)
	if err := r.redisClient.Del(ctx, chatUsersKey).Err(); err != nil {
		r.logger.WithFields(logging.Fields{
			"error":     err,
			"redis_key": chatUsersKey,
		}).Error("An error occurred while deleting the key from redis")

		return err
	}

	return nil
}
