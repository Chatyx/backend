package service

import (
	"context"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/internal/repository"
	"github.com/Mort4lis/scht-backend/pkg/logging"
)

type userChatService struct {
	repo   repository.UserChatRepository
	logger logging.Logger
}

func NewUserChatService(repo repository.UserChatRepository) UserChatService {
	return &userChatService{
		repo:   repo,
		logger: logging.GetLogger(),
	}
}

func (s *userChatService) ListUsersWhoBelongToChat(ctx context.Context, chatID, memberID string) ([]domain.User, error) {
	users, err := s.repo.ListUsersWhoBelongToChat(ctx, chatID)
	if err != nil {
		return nil, err
	}

	isBelong := false

	for _, user := range users {
		if user.ID == memberID {
			isBelong = true
			break
		}
	}

	if !isBelong {
		s.logger.WithFields(logging.Fields{
			"user_id": memberID,
			"chat_id": chatID,
		}).Debug("user doesn't belong to this chat")

		return nil, domain.ErrChatNotFound
	}

	return users, nil
}

func (s *userChatService) IsUserBelongToChat(ctx context.Context, userID, chatID string) (bool, error) {
	return s.repo.IsUserBelongToChat(ctx, userID, chatID)
}
