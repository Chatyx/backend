package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Mort4lis/scht-backend/internal/domain"

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

func (r *chatMemberCacheRepositoryDecorator) IsInChat(ctx context.Context, memberKey domain.ChatMemberIdentity) (bool, error) {
	if err := r.populateCacheIfNotExist(ctx, memberKey.ChatID); err != nil {
		return false, err
	}

	chatUsersKey := r.getChatUsersKey(memberKey.ChatID)
	logger := r.logger.WithFields(logging.Fields{
		"user_id":   memberKey.UserID,
		"redis_key": chatUsersKey,
	})

	isIn, err := r.redisClient.SIsMember(ctx, chatUsersKey, memberKey.UserID).Result()
	if err != nil {
		logger.WithError(err).Error("An error occurred while checking if member is in the chat")
		return false, err
	}

	return isIn, nil
}

func (r *chatMemberCacheRepositoryDecorator) Create(ctx context.Context, memberKey domain.ChatMemberIdentity) error {
	if err := r.ChatMemberRepository.Create(ctx, memberKey); err != nil {
		return err
	}

	if err := r.populateCacheIfNotExist(ctx, memberKey.ChatID); err != nil {
		return err
	}

	chatUsersKey := r.getChatUsersKey(memberKey.ChatID)
	logger := r.logger.WithFields(logging.Fields{
		"user_id":   memberKey.UserID,
		"redis_key": chatUsersKey,
	})

	if err := r.redisClient.SAdd(ctx, chatUsersKey, memberKey.UserID).Err(); err != nil {
		logger.WithError(err).Error("An error occurred while setting user ids")
		return err
	}

	return nil
}

func (r *chatMemberCacheRepositoryDecorator) Update(ctx context.Context, dto domain.UpdateChatMemberDTO) error {
	if err := r.ChatMemberRepository.Update(ctx, dto); err != nil {
		return err
	}

	if err := r.populateCacheIfNotExist(ctx, dto.ChatID); err != nil {
		return err
	}

	chatUsersKey := r.getChatUsersKey(dto.ChatID)
	logger := r.logger.WithFields(logging.Fields{
		"user_id":   dto.UserID,
		"redis_key": chatUsersKey,
	})

	switch dto.StatusID {
	case domain.InChat:
		if err := r.redisClient.SAdd(ctx, chatUsersKey, dto.UserID).Err(); err != nil {
			logger.WithError(err).Error("An error occurred while setting user ids")
			return err
		}
	case domain.Left, domain.Kicked:
		if err := r.redisClient.SRem(ctx, chatUsersKey, dto.UserID).Err(); err != nil {
			logger.WithError(err).Error("An error occurred while removing user ids")
			return err
		}
	default:
		return domain.ErrChatMemberUnknownStatus
	}

	return nil
}

func (r *chatMemberCacheRepositoryDecorator) populateCacheIfNotExist(ctx context.Context, chatID string) error {
	chatUsersKey := r.getChatUsersKey(chatID)
	logger := r.logger.WithFields(logging.Fields{
		"redis_key": chatUsersKey,
	})

	val, err := r.redisClient.Exists(ctx, chatUsersKey).Result()
	if err != nil {
		logger.WithError(err).Error("An error occurred while checking existence the key")
		return err
	}

	if val == 0 {
		if err = r.cacheChatUserIDs(ctx, chatID); err != nil {
			return err
		}
	}

	return nil
}

func (r *chatMemberCacheRepositoryDecorator) cacheChatUserIDs(ctx context.Context, chatID string) error {
	chatUsersKey := r.getChatUsersKey(chatID)
	logger := r.logger.WithFields(logging.Fields{
		"redis_key": chatUsersKey,
	})

	members, err := r.ListByChatID(ctx, chatID)
	if err != nil {
		return err
	}

	userIDs := make([]interface{}, 0, len(members))

	for _, member := range members {
		if member.IsInChat() {
			userIDs = append(userIDs, member.UserID)
		}
	}

	if len(userIDs) == 0 {
		return nil
	}

	if err = r.redisClient.SAdd(ctx, chatUsersKey, userIDs...).Err(); err != nil {
		logger.WithFields(logging.Fields{
			"error":    err,
			"user_ids": userIDs,
		}).Error("An error occurred while setting user ids")

		return err
	}

	if err = r.redisClient.Expire(ctx, chatUsersKey, 15*time.Minute).Err(); err != nil {
		logger.WithError(err).Error("An error occurred while setting TTL")
		return err
	}

	return nil
}

func (r *chatMemberCacheRepositoryDecorator) getChatUsersKey(chatID string) string {
	return fmt.Sprintf("chat:%s:user_ids", chatID)
}
