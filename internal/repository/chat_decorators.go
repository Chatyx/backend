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
	dbPool      PgxPool
	redisClient *redis.Client
	logger      logging.Logger
}

func NewChatCacheRepositoryDecorator(repo ChatRepository, redisClient *redis.Client, dbPool PgxPool) ChatRepository {
	return &chatCacheRepositoryDecorator{
		repo:        repo,
		dbPool:      dbPool,
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
	userIDs, err := r.getUserIDsWhoBelongToChat(ctx, chatID)
	if err != nil {
		return err
	}

	if err = r.repo.Delete(ctx, chatID, creatorID); err != nil {
		return err
	}

	pipeline := r.redisClient.Pipeline()

	for _, userID := range userIDs {
		userChatsKey := fmt.Sprintf("user:%s:chat_ids", userID)
		if err = pipeline.SRem(ctx, userChatsKey, chatID).Err(); err != nil {
			r.logger.WithError(err).Error("An error occurred while removing chat_id from redis")
			return err
		}
	}

	if _, err = pipeline.Exec(ctx); err != nil {
		r.logger.WithError(err).Error("An error occurred while removing chat_ids from redis")
		return err
	}

	return nil
}

func (r *chatCacheRepositoryDecorator) getUserIDsWhoBelongToChat(ctx context.Context, chatID string) ([]string, error) {
	query := `SELECT users_chats.user_id FROM users_chats 
	INNER JOIN users 
		ON users_chats.user_id = users.id
	WHERE users_chats.chat_id = $1 AND users.is_deleted IS FALSE`

	rows, err := r.dbPool.Query(ctx, query, chatID)
	if err != nil {
		r.logger.WithError(err).Error("An error occurred while getting user_ids who belong to chat")
		return nil, err
	}

	defer rows.Close()

	userIDs := make([]string, 0)

	for rows.Next() {
		var userID string

		if err = rows.Scan(&userID); err != nil {
			r.logger.WithError(err).Error("Unable to scan user id")
			return nil, err
		}

		userIDs = append(userIDs, userID)
	}

	if err = rows.Err(); err != nil {
		r.logger.WithError(err).Error("An error occurred while getting user_ids who belong to chat")
		return nil, err
	}

	return userIDs, nil
}
