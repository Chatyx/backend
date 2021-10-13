package service

import (
	"context"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/internal/repository"
	"github.com/Mort4lis/scht-backend/pkg/logging"
)

type chatService struct {
	repo   repository.ChatRepository
	logger logging.Logger
}

func NewChatService(repo repository.ChatRepository) ChatService {
	return &chatService{
		repo:   repo,
		logger: logging.GetLogger(),
	}
}

func (s *chatService) List(ctx context.Context, memberID string) ([]domain.Chat, error) {
	return s.repo.List(ctx, memberID)
}

func (s *chatService) Create(ctx context.Context, dto domain.CreateChatDTO) (domain.Chat, error) {
	return s.repo.Create(ctx, dto)
}

func (s *chatService) GetByID(ctx context.Context, chatID, memberID string) (domain.Chat, error) {
	return s.repo.GetByID(ctx, chatID, memberID)
}

func (s *chatService) Update(ctx context.Context, dto domain.UpdateChatDTO) (domain.Chat, error) {
	return s.repo.Update(ctx, dto)
}

func (s *chatService) Delete(ctx context.Context, chatID, creatorID string) error {
	return s.repo.Delete(ctx, chatID, creatorID)
}