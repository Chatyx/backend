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

func (s *chatMemberService) ListMembersInChat(ctx context.Context, chatID, userID string) ([]domain.ChatMember, error) {
	members, err := s.repo.ListMembersInChat(ctx, chatID)
	if err != nil {
		return nil, err
	}

	isIn := false

	for _, member := range members {
		if member.UserID == userID {
			isIn = true
			break
		}
	}

	if !isIn {
		s.logger.WithFields(logging.Fields{
			"user_id": userID,
			"chat_id": chatID,
		}).Debug("member isn't in this chat")

		return nil, domain.ErrChatNotFound
	}

	return members, nil
}

func (s *chatMemberService) IsMemberInChat(ctx context.Context, userID, chatID string) (bool, error) {
	return s.repo.IsMemberInChat(ctx, userID, chatID)
}
