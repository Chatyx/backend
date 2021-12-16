package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/go-redis/redis/v8"
)

type chatMemberCacheRepositoryDecorator struct {
	ChatMemberRepository
	redisClient *redis.Client
}

func NewChatMemberCacheRepository(repo ChatMemberRepository, redisClient *redis.Client) ChatMemberRepository {
	return &chatMemberCacheRepositoryDecorator{
		ChatMemberRepository: repo,
		redisClient:          redisClient,
	}
}

func (r *chatMemberCacheRepositoryDecorator) IsInChat(ctx context.Context, memberKey domain.ChatMemberIdentity) (bool, error) {
	if err := r.populateCacheIfNotExist(ctx, memberKey.ChatID); err != nil {
		return false, fmt.Errorf("failed to populate chat members cache: %w", err)
	}

	isIn, err := r.redisClient.SIsMember(ctx, r.getChatUsersKey(memberKey.ChatID), memberKey.UserID).Result()
	if err != nil {
		return false, fmt.Errorf("an error occurred while checking if member is in the chat to the cache: %v", err)
	}

	return isIn, nil
}

func (r *chatMemberCacheRepositoryDecorator) Create(ctx context.Context, memberKey domain.ChatMemberIdentity) error {
	if err := r.ChatMemberRepository.Create(ctx, memberKey); err != nil {
		return err
	}

	if err := r.populateCacheIfNotExist(ctx, memberKey.ChatID); err != nil {
		return fmt.Errorf("failed to populate chat members cache: %w", err)
	}

	if err := r.redisClient.SAdd(ctx, r.getChatUsersKey(memberKey.ChatID), memberKey.UserID).Err(); err != nil {
		return fmt.Errorf("an error occurred while adding user_id to chat members cache: %v", err)
	}

	return nil
}

func (r *chatMemberCacheRepositoryDecorator) Update(ctx context.Context, dto domain.UpdateChatMemberDTO) error {
	if err := r.ChatMemberRepository.Update(ctx, dto); err != nil {
		return err
	}

	if err := r.populateCacheIfNotExist(ctx, dto.ChatID); err != nil {
		return fmt.Errorf("failed to populate chat members cache: %w", err)
	}

	chatUsersKey := r.getChatUsersKey(dto.ChatID)

	switch dto.StatusID {
	case domain.InChat:
		if err := r.redisClient.SAdd(ctx, chatUsersKey, dto.UserID).Err(); err != nil {
			return fmt.Errorf("an error occurred while adding user_id to chat members cache: %v", err)
		}
	case domain.Left, domain.Kicked:
		if err := r.redisClient.SRem(ctx, chatUsersKey, dto.UserID).Err(); err != nil {
			return fmt.Errorf("an error occurred while removing user_id from chat members cache: %v", err)
		}
	default:
		return fmt.Errorf("%w", domain.ErrChatMemberUnknownStatus)
	}

	return nil
}

func (r *chatMemberCacheRepositoryDecorator) populateCacheIfNotExist(ctx context.Context, chatID string) error {
	val, err := r.redisClient.Exists(ctx, r.getChatUsersKey(chatID)).Result()
	if err != nil {
		return fmt.Errorf("an error occurred while checking existence the chat members key: %v", err)
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
		return fmt.Errorf("an error occurred while adding users_ids to chat members cache: %v", err)
	}

	if err = r.redisClient.Expire(ctx, chatUsersKey, 15*time.Minute).Err(); err != nil {
		return fmt.Errorf("an error occurred while setting TTL: %v", err)
	}

	return nil
}

func (r *chatMemberCacheRepositoryDecorator) getChatUsersKey(chatID string) string {
	return fmt.Sprintf("chat:%s:user_ids", chatID)
}
