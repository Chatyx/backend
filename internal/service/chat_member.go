package service

import (
	"context"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/internal/repository"
	"github.com/Mort4lis/scht-backend/pkg/logging"
)

type chatMemberService struct {
	repo   repository.ChatMemberRepository
	logger logging.Logger
}

func NewChatMemberService(repo repository.ChatMemberRepository) ChatMemberService {
	return &chatMemberService{
		repo:   repo,
		logger: logging.GetLogger(),
	}
}

func (s *chatMemberService) ListMembersWhoBelongToChat(ctx context.Context, chatID, userID string) ([]domain.ChatMember, error) {
	members, err := s.repo.ListMembersWhoBelongToChat(ctx, chatID)
	if err != nil {
		return nil, err
	}

	isBelong := false

	for _, member := range members {
		if member.UserID == userID {
			isBelong = true
			break
		}
	}

	if !isBelong {
		s.logger.WithFields(logging.Fields{
			"user_id": userID,
			"chat_id": chatID,
		}).Debug("user doesn't belong to this chat")

		return nil, domain.ErrChatNotFound
	}

	return members, nil
}

func (s *chatMemberService) IsMemberBelongToChat(ctx context.Context, userID, chatID string) (bool, error) {
	return s.repo.IsMemberBelongToChat(ctx, userID, chatID)
}
