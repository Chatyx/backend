package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/go-redis/redis/v8"
)

type userChatCacheRepositoryDecorator struct {
	UserChatRepository
	redisClient *redis.Client
	logger      logging.Logger
}

func NewUserChatCacheRepository(repo UserChatRepository, redisClient *redis.Client) UserChatRepository {
	return &userChatCacheRepositoryDecorator{
		UserChatRepository: repo,
		redisClient:        redisClient,
		logger:             logging.GetLogger(),
	}
}

func (r *userChatCacheRepositoryDecorator) IsUserBelongToChat(ctx context.Context, userID, chatID string) (bool, error) {
	chatUsersKey := fmt.Sprintf("chat:%s:user_ids", chatID)
	logger := r.logger.WithFields(logging.Fields{
		"user_id":   userID,
		"chat_id":   chatID,
		"redis_key": chatUsersKey,
	})

	val, err := r.redisClient.Exists(ctx, chatUsersKey).Result()
	if err != nil {
		logger.WithError(err).Errorf("An error occurred while checking existence the key")
		return false, err
	}

	if val == 0 {
		if err = r.cacheChatUserIDs(ctx, chatID); err != nil {
			return false, err
		}
	}

	isBelong, err := r.redisClient.SIsMember(ctx, chatUsersKey, userID).Result()
	if err != nil {
		logger.WithError(err).Error("An error occurred while checking if user belongs to the chat")
		return false, err
	}

	return isBelong, nil
}

func (r *userChatCacheRepositoryDecorator) cacheChatUserIDs(ctx context.Context, chatID string) error {
	chatUsersKey := fmt.Sprintf("chat:%s:user_ids", chatID)
	logger := r.logger.WithFields(logging.Fields{
		"redis_key": chatUsersKey,
	})

	users, err := r.ListUsersWhoBelongToChat(ctx, chatID)
	if err != nil {
		return err
	}

	if len(users) == 0 {
		return nil
	}

	userIDs := make([]interface{}, 0, len(users))
	for _, user := range users {
		userIDs = append(userIDs, user.ID)
	}

	if err = r.redisClient.SAdd(ctx, chatUsersKey, userIDs...).Err(); err != nil {
		logger.WithError(err).Error("An error occurred while setting user ids")
		return err
	}

	if err = r.redisClient.Expire(ctx, chatUsersKey, 15*time.Minute).Err(); err != nil {
		logger.WithError(err).Error("An error occurred while setting TTL")
		return err
	}

	return nil
}
