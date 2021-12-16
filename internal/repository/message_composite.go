package repository

import (
	"context"
	"time"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/pkg/logging"
)

type messageCompositeRepository struct {
	cacheRepo MessageRepository
	dbRepo    MessageRepository
}

func NewMessageCompositeRepository(cacheRepo, dbRepo MessageRepository) MessageRepository {
	return &messageCompositeRepository{
		cacheRepo: cacheRepo,
		dbRepo:    dbRepo,
	}
}

func (r *messageCompositeRepository) Create(ctx context.Context, dto domain.CreateMessageDTO) (domain.Message, error) {
	return r.cacheRepo.Create(ctx, dto)
}

func (r *messageCompositeRepository) List(ctx context.Context, chatID string, timestamp time.Time) ([]domain.Message, error) {
	messages, err := r.cacheRepo.List(ctx, chatID, timestamp)
	if err != nil {
		return nil, err
	}

	if len(messages) != 0 {
		return messages, err
	}

	logger := logging.GetLoggerFromContext(ctx)
	logger.Debug("not found messages into the cache for chat, going to the database...")

	return r.dbRepo.List(ctx, chatID, timestamp)
}
