package service

import (
	"context"
	"fmt"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/internal/repository"
)

type chatService struct {
	repo repository.ChatRepository
}

func NewChatService(repo repository.ChatRepository) ChatService {
	return &chatService{repo: repo}
}

func (s *chatService) List(ctx context.Context, userID string) ([]domain.Chat, error) {
	chats, err := s.repo.List(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get list of chats: %w", err)
	}

	return chats, nil
}

func (s *chatService) Create(ctx context.Context, dto domain.CreateChatDTO) (domain.Chat, error) {
	chat, err := s.repo.Create(ctx, dto)
	if err != nil {
		return domain.Chat{}, fmt.Errorf("failed to create chat: %w", err)
	}

	return chat, nil
}

func (s *chatService) Get(ctx context.Context, memberKey domain.ChatMemberIdentity) (domain.Chat, error) {
	chat, err := s.repo.Get(ctx, memberKey)
	if err != nil {
		return domain.Chat{}, fmt.Errorf("failed to get chat: %w", err)
	}

	return chat, nil
}

func (s *chatService) Update(ctx context.Context, dto domain.UpdateChatDTO) (domain.Chat, error) {
	chat, err := s.repo.Update(ctx, dto)
	if err != nil {
		return domain.Chat{}, fmt.Errorf("failed to update chat: %w", err)
	}

	return chat, nil
}

func (s *chatService) Delete(ctx context.Context, memberKey domain.ChatMemberIdentity) error {
	if err := s.repo.Delete(ctx, memberKey); err != nil {
		return fmt.Errorf("failed to delete chat: %w", err)
	}

	return nil
}
