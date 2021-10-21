package repository

import (
	"context"
	"fmt"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/go-redis/redis/v8"
)

type userCacheRepositoryDecorator struct {
	repo        UserRepository
	redisClient *redis.Client
	logger      logging.Logger
}

func NewUserCacheRepositoryDecorator(repo UserRepository, redisClient *redis.Client) UserRepository {
	return &userCacheRepositoryDecorator{
		repo:        repo,
		redisClient: redisClient,
		logger:      logging.GetLogger(),
	}
}

func (r *userCacheRepositoryDecorator) List(ctx context.Context) ([]domain.User, error) {
	return r.repo.List(ctx)
}

func (r *userCacheRepositoryDecorator) Create(ctx context.Context, dto domain.CreateUserDTO) (domain.User, error) {
	return r.repo.Create(ctx, dto)
}

func (r *userCacheRepositoryDecorator) GetByID(ctx context.Context, id string) (domain.User, error) {
	return r.repo.GetByID(ctx, id)
}

func (r *userCacheRepositoryDecorator) GetByUsername(ctx context.Context, username string) (domain.User, error) {
	return r.repo.GetByUsername(ctx, username)
}

func (r *userCacheRepositoryDecorator) Update(ctx context.Context, dto domain.UpdateUserDTO) (domain.User, error) {
	return r.repo.Update(ctx, dto)
}

func (r *userCacheRepositoryDecorator) UpdatePassword(ctx context.Context, id, password string) error {
	return r.repo.UpdatePassword(ctx, id, password)
}

func (r *userCacheRepositoryDecorator) Delete(ctx context.Context, id string) error {
	if err := r.repo.Delete(ctx, id); err != nil {
		return err
	}

	userChatsKey := fmt.Sprintf("user:%s:chat_ids", id)

	if err := r.redisClient.Del(ctx, userChatsKey).Err(); err != nil {
		r.logger.WithFields(logging.Fields{
			"error": err,
			"key":   userChatsKey,
		}).Error("An error occurred while removing redis key")

		return err
	}

	return nil
}
