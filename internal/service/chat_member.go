package service

import (
	"context"
	"fmt"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/internal/repository"
	"github.com/Mort4lis/scht-backend/internal/utils"
	pkgErrors "github.com/Mort4lis/scht-backend/pkg/errors"
)

type chatMemberService struct {
	userService         UserService
	chatMemberRepo      repository.ChatMemberRepository
	messageRepo         repository.MessageRepository
	messagePubSub       repository.MessagePubSub
	memberStatusMatrix  utils.StatusMatrix
	creatorStatusMatrix utils.StatusMatrix
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
	}
}

func (s *chatMemberService) List(ctx context.Context, memberKey domain.ChatMemberIdentity) ([]domain.ChatMember, error) {
	members, err := s.chatMemberRepo.ListByChatID(ctx, memberKey.ChatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get list of chat members: %w", err)
	}

	isIn := false

	for _, member := range members {
		if member.UserID == memberKey.UserID && member.IsInChat() {
			isIn = true
			break
		}
	}

	if !isIn {
		return nil, fmt.Errorf("member isn't in this chat: %w", domain.ErrChatNotFound)
	}

	return members, nil
}

func (s *chatMemberService) GetByKey(ctx context.Context, memberKey domain.ChatMemberIdentity, user domain.AuthUser) (domain.ChatMember, error) {
	ok, err := s.chatMemberRepo.IsInChat(ctx, domain.ChatMemberIdentity{
		UserID: user.UserID,
		ChatID: memberKey.ChatID,
	})
	if err != nil {
		return domain.ChatMember{}, fmt.Errorf("failed to check if authenticated user is in this chat: %w", err)
	}

	if !ok {
		return domain.ChatMember{}, fmt.Errorf("member isn't in this chat: %w", domain.ErrChatNotFound)
	}

	return s.chatMemberRepo.GetByKey(ctx, memberKey)
}

func (s *chatMemberService) JoinToChat(ctx context.Context, memberKey domain.ChatMemberIdentity, user domain.AuthUser) error {
	ok, err := s.chatMemberRepo.IsChatCreator(ctx, domain.ChatMemberIdentity{
		UserID: user.UserID,
		ChatID: memberKey.ChatID,
	})
	if err != nil {
		return fmt.Errorf("failed to check if authenticated user is a creator of this chat: %w", err)
	}

	if !ok {
		return fmt.Errorf("can't join member to chat due the authenticated user isn't a creator: %w", domain.ErrChatNotFound)
	}

	joinUser, err := s.userService.GetByID(ctx, memberKey.UserID)
	if err != nil {
		return fmt.Errorf("failed to get joined member: %w", err)
	}

	if err = s.chatMemberRepo.Create(ctx, memberKey); err != nil {
		return fmt.Errorf("failed to create member in this chat: %w", err)
	}

	dto := domain.CreateMessageDTO{
		ActionID: domain.MessageJoinAction,
		Text:     fmt.Sprintf("%s successfully joined to the chat", joinUser.Username),
		ChatID:   memberKey.ChatID,
		SenderID: memberKey.UserID,
	}

	message, err := s.messageRepo.Create(ctx, dto)
	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	if err = s.messagePubSub.Publish(ctx, message); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

func (s *chatMemberService) UpdateStatus(ctx context.Context, dto domain.UpdateChatMemberDTO, user domain.AuthUser) error {
	isAuthUserCreator, err := s.chatMemberRepo.IsChatCreator(ctx, domain.ChatMemberIdentity{
		UserID: user.UserID,
		ChatID: dto.ChatID,
	})
	if err != nil {
		return fmt.Errorf("failed to check if authenticated user is a creator of this chat: %w", err)
	}

	member, err := s.chatMemberRepo.GetByKey(ctx, dto.ChatMemberIdentity)
	if err != nil {
		return fmt.Errorf("failed to get member: %w", err)
	}

	var mx utils.StatusMatrix

	switch {
	case member.UserID == user.UserID:
		mx = s.memberStatusMatrix
	case isAuthUserCreator:
		mx = s.creatorStatusMatrix
	default:
		return fmt.Errorf("can't update chat member status due authenticated user isn't a creator of this chat: %w", domain.ErrChatNotFound)
	}

	if !mx.IsCorrectTransit(member.StatusID, dto.StatusID) {
		ctxFields := pkgErrors.ContextFields{
			"source.status_id":      member.StatusID,
			"destination.status_id": dto.StatusID,
		}

		return pkgErrors.WrapInContextError(domain.ErrChatMemberWrongStatusTransit, "wrong chat member transition", ctxFields)
	}

	return s.updateStatus(ctx, member, dto)
}

func (s *chatMemberService) updateStatus(ctx context.Context, member domain.ChatMember, dto domain.UpdateChatMemberDTO) error {
	if err := s.chatMemberRepo.Update(ctx, dto); err != nil {
		return fmt.Errorf("failed to update chat member: %w", err)
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
		return fmt.Errorf("unknown chat member status: %w", domain.ErrChatMemberUnknownStatus)
	}

	message, err := s.messageRepo.Create(ctx, msgDTO)
	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	if err = s.messagePubSub.Publish(ctx, message); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}
