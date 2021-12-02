package service

import (
	"context"
	"fmt"

	"github.com/Mort4lis/scht-backend/internal/utils"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/internal/repository"
	"github.com/Mort4lis/scht-backend/pkg/logging"
)

type chatMemberService struct {
	userService         UserService
	chatMemberRepo      repository.ChatMemberRepository
	messageRepo         repository.MessageRepository
	messagePubSub       repository.MessagePubSub
	memberStatusMatrix  utils.StatusMatrix
	creatorStatusMatrix utils.StatusMatrix
	logger              logging.Logger
}

type ChatMemberConfig struct {
	UserService    UserService
	ChatMemberRepo repository.ChatMemberRepository
	MessageRepo    repository.MessageRepository
	MessagePubSub  repository.MessagePubSub
}

func NewChatMemberService(cfg ChatMemberConfig) ChatMemberService {
	return &chatMemberService{
		userService:    cfg.UserService,
		chatMemberRepo: cfg.ChatMemberRepo,
		messageRepo:    cfg.MessageRepo,
		messagePubSub:  cfg.MessagePubSub,
		memberStatusMatrix: utils.StatusMatrix{
			domain.InChat: utils.NewStatusSet(domain.Left),
			domain.Left:   utils.NewStatusSet(domain.InChat),
		},
		creatorStatusMatrix: utils.StatusMatrix{
			domain.InChat: utils.NewStatusSet(domain.Kicked),
			domain.Kicked: utils.NewStatusSet(domain.InChat),
		},
		logger: logging.GetLogger(),
	}
}

func (s *chatMemberService) ListByChatID(ctx context.Context, chatID, userID string) ([]domain.ChatMember, error) {
	members, err := s.chatMemberRepo.ListByChatID(ctx, chatID)
	if err != nil {
		return nil, err
	}

	isIn := false

	for _, member := range members {
		if member.UserID == userID && member.IsInChat() {
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

func (s *chatMemberService) ListByUserID(ctx context.Context, userID string) ([]domain.ChatMember, error) {
	return s.chatMemberRepo.ListByUserID(ctx, userID)
}

func (s *chatMemberService) IsInChat(ctx context.Context, userID, chatID string) (bool, error) {
	return s.chatMemberRepo.IsInChat(ctx, userID, chatID)
}

func (s *chatMemberService) Get(ctx context.Context, chatID, userID string) (domain.ChatMember, error) {
	return s.chatMemberRepo.Get(ctx, userID, chatID)
}

func (s *chatMemberService) GetAnother(ctx context.Context, authUserID, chatID, userID string) (domain.ChatMember, error) {
	ok, err := s.IsInChat(ctx, authUserID, chatID)
	if err != nil {
		return domain.ChatMember{}, err
	}

	if !ok {
		s.logger.WithFields(logging.Fields{
			"user_id": authUserID,
			"chat_id": chatID,
		}).Debug("member isn't in this chat")

		return domain.ChatMember{}, domain.ErrChatNotFound
	}

	return s.Get(ctx, chatID, userID)
}

func (s *chatMemberService) JoinToChat(ctx context.Context, chatID, creatorID, userID string) error {
	ok, err := s.chatMemberRepo.IsChatCreator(ctx, creatorID, chatID)
	if err != nil {
		return err
	}

	if !ok {
		s.logger.WithFields(logging.Fields{
			"user_id": creatorID,
			"chat_id": chatID,
		}).Debug("can't join member to chat due the user isn't a creator")

		return domain.ErrChatNotFound
	}

	user, err := s.userService.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	if err = s.chatMemberRepo.Create(ctx, userID, chatID); err != nil {
		return err
	}

	dto := domain.CreateMessageDTO{
		ActionID: domain.MessageJoinAction,
		Text:     fmt.Sprintf("%s successfully joined to the chat", user.Username),
		ChatID:   chatID,
		SenderID: userID,
	}

	message, err := s.messageRepo.Create(ctx, dto)
	if err != nil {
		return err
	}

	return s.messagePubSub.Publish(ctx, message)
}

func (s *chatMemberService) UpdateStatus(ctx context.Context, dto domain.UpdateChatMemberDTO) error {
	return s.updateStatus(ctx, s.memberStatusMatrix, dto)
}

func (s *chatMemberService) UpdateStatusByCreator(ctx context.Context, creatorID string, dto domain.UpdateChatMemberDTO) error {
	ok, err := s.chatMemberRepo.IsChatCreator(ctx, creatorID, dto.ChatID)
	if err != nil {
		return err
	}

	if !ok {
		s.logger.WithFields(logging.Fields{
			"user_id": creatorID,
			"chat_id": dto.ChatID,
		}).Debug("can't update member status due the user isn't a creator")

		return domain.ErrChatNotFound
	}

	return s.updateStatus(ctx, s.creatorStatusMatrix, dto)
}

func (s *chatMemberService) updateStatus(ctx context.Context, mx utils.StatusMatrix, dto domain.UpdateChatMemberDTO) error {
	member, err := s.chatMemberRepo.Get(ctx, dto.UserID, dto.ChatID)
	if err != nil {
		return err
	}

	logger := s.logger.WithFields(logging.Fields{
		"user_id": dto.UserID,
		"chat_id": dto.ChatID,
	})

	if member.IsCreator {
		logger.Debug("can't update chat member status if he's a chat creator")
		return domain.ErrChatCreatorInvalidUpdateStatus
	}

	if !mx.IsCorrectTransit(member.StatusID, dto.StatusID) {
		logger.WithFields(logging.Fields{
			"from_status_id": member.StatusID,
			"to_status_id":   dto.StatusID,
		}).Debug("wrong chat member transition")

		return domain.ErrChatMemberWrongStatusTransit
	}

	if err = s.chatMemberRepo.Update(ctx, dto); err != nil {
		return err
	}

	msgDTO := domain.CreateMessageDTO{
		ChatID:   dto.ChatID,
		SenderID: dto.UserID,
	}

	switch dto.StatusID {
	case domain.InChat:
		msgDTO.ActionID = domain.MessageJoinAction
		msgDTO.Text = fmt.Sprintf("%s successfully joined to the chat", member.Username)
	case domain.Left:
		msgDTO.ActionID = domain.MessageLeaveAction
		msgDTO.Text = fmt.Sprintf("%s has left from the chat", member.Username)
	case domain.Kicked:
		msgDTO.ActionID = domain.MessageKickAction
		msgDTO.Text = fmt.Sprintf("%s has been kicked from the chat", member.Username)
	default:
		logger.WithFields(logging.Fields{
			"status_id": dto.StatusID,
		}).Error("unknown chat member status")

		return domain.ErrChatMemberUnknownStatus
	}

	message, err := s.messageRepo.Create(ctx, msgDTO)
	if err != nil {
		return err
	}

	return s.messagePubSub.Publish(ctx, message)
}
