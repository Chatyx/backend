package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/go-redis/redis/v8"
)

type chatMemberCacheRepositoryDecorator struct {
	ChatMemberRepository
	redisClient *redis.Client
	logger      logging.Logger
}

func NewChatMemberCacheRepository(repo ChatMemberRepository, redisClient *redis.Client) ChatMemberRepository {
	return &chatMemberCacheRepositoryDecorator{
		ChatMemberRepository: repo,
		redisClient:          redisClient,
		logger:               logging.GetLogger(),
	}
}

func (r *chatMemberCacheRepositoryDecorator) IsMemberInChat(ctx context.Context, userID, chatID string) (bool, error) {
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

	isIn, err := r.redisClient.SIsMember(ctx, chatUsersKey, userID).Result()
	if err != nil {
		logger.WithError(err).Error("An error occurred while checking if member is in the chat")
		return false, err
	}

	return isIn, nil
}

func (r *chatMemberCacheRepositoryDecorator) cacheChatUserIDs(ctx context.Context, chatID string) error {
	chatUsersKey := fmt.Sprintf("chat:%s:user_ids", chatID)
	logger := r.logger.WithFields(logging.Fields{
		"redis_key": chatUsersKey,
	})

	members, err := r.ListMembersInChat(ctx, chatID)
	if err != nil {
		return err
	}

	if len(members) == 0 {
		return nil
	}

	userIDs := make([]interface{}, 0, len(members))
	for _, member := range members {
		userIDs = append(userIDs, member.UserID)
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

func (r *chatMemberCacheRepositoryDecorator) Create(ctx context.Context, userID string, chatID string) error {
	if err := r.ChatMemberRepository.Create(ctx, userID, chatID); err != nil {
		return err
	}

	chatUsersKey := fmt.Sprintf("chat:%s:user_ids", chatID)
	logger := r.logger.WithFields(logging.Fields{
		"user_id":   userID,
		"chat_id":   chatID,
		"redis_key": chatUsersKey,
	})

	val, err := r.redisClient.Exists(ctx, chatUsersKey).Result()
	if err != nil {
		logger.WithError(err).Errorf("An error occurred while checking existence the key")
		return err
	}

	if val == 0 {
		if err = r.cacheChatUserIDs(ctx, chatID); err != nil {
			return err
		}
	}

	if err = r.redisClient.SAdd(ctx, chatUsersKey, userID).Err(); err != nil {
		logger.WithError(err).Error("An error occurred while setting user ids")
		return err
	}

	return nil
}
