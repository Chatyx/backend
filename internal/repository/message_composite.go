package repository

import (
	"context"

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

func (r *messageCompositeRepository) List(ctx context.Context, chatID string, dto domain.MessageListDTO) (domain.MessageList, error) {
	messageList, err := r.cacheRepo.List(ctx, chatID, dto)
	if err != nil {
		return domain.MessageList{}, err
	}

	if len(messageList.Messages) != 0 {
		return messageList, err
	}

	logger := logging.GetLoggerFromContext(ctx)
	logger.Debug("not found messages into the cache for chat, going to the database...")

	return r.dbRepo.List(ctx, chatID, dto)
}
