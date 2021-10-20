package repository

import (
	"context"
	"fmt"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/go-redis/redis/v8"
)

type chatCacheRepositoryDecorator struct {
	repo        ChatRepository
	redisClient *redis.Client
	logger      logging.Logger
}

func NewChatCacheRepositoryDecorator(repo ChatRepository, redisClient *redis.Client) ChatRepository {
	return &chatCacheRepositoryDecorator{
		repo:        repo,
		redisClient: redisClient,
		logger:      logging.GetLogger(),
	}
}

func (r *chatCacheRepositoryDecorator) List(ctx context.Context, memberID string) ([]domain.Chat, error) {
	chats, err := r.repo.List(ctx, memberID)
	if err != nil {
		return nil, err
	}

	userChatsKey := fmt.Sprintf("user:%s:chat_ids", memberID)

	val, err := r.redisClient.Exists(ctx, userChatsKey).Result()
	if err != nil {
		r.logger.WithError(err).Errorf("An error occurred while checking existence the key %s", userChatsKey)
		return nil, err
	}

	if val == 0 {
		chatIDs := make([]interface{}, 0, len(chats))
		for _, chat := range chats {
			chatIDs = append(chatIDs, chat.ID)
		}

		if err = r.redisClient.SAdd(ctx, userChatsKey, chatIDs...).Err(); err != nil {
			r.logger.WithError(err).Errorf("An error occurred while adding chat ids into the redis")
			return nil, err
		}
	}

	return chats, nil
}

func (r *chatCacheRepositoryDecorator) Create(ctx context.Context, dto domain.CreateChatDTO) (domain.Chat, error) {
	return r.repo.Create(ctx, dto)
}

func (r *chatCacheRepositoryDecorator) GetByID(ctx context.Context, chatID, memberID string) (domain.Chat, error) {
	return r.repo.GetByID(ctx, chatID, memberID)
}

func (r *chatCacheRepositoryDecorator) IsBelongToChat(ctx context.Context, chatID, memberID string) (bool, error) {
	userChatsKey := fmt.Sprintf("user:%s:chat_ids", memberID)

	val, err := r.redisClient.Exists(ctx, userChatsKey).Result()
	if err != nil {
		r.logger.WithError(err).Errorf("An error occurred while checking existence the key %s", userChatsKey)
		return false, err
	}

	if val == 0 {
		if _, err = r.List(ctx, memberID); err != nil {
			return false, err
		}
	}

	exist, err := r.redisClient.SIsMember(ctx, userChatsKey, chatID).Result()
	if err != nil {
		r.logger.WithError(err).Errorf("An error occurred while checking if")
		return false, err
	}

	return exist, nil
}

func (r *chatCacheRepositoryDecorator) Update(ctx context.Context, dto domain.UpdateChatDTO) (domain.Chat, error) {
	return r.repo.Update(ctx, dto)
}

func (r *chatCacheRepositoryDecorator) Delete(ctx context.Context, chatID, creatorID string) error {
	return r.repo.Delete(ctx, chatID, creatorID)
}
