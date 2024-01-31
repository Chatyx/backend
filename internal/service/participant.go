package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Chatyx/backend/internal/entity"
	"github.com/Chatyx/backend/pkg/ctxutil"
)

type GroupParticipantFunc func(p *entity.GroupParticipant) error

//go:generate mockery --inpackage --testonly --case underscore --name GroupParticipantRepository
type GroupParticipantRepository interface {
	List(ctx context.Context, groupID int) ([]entity.GroupParticipant, error)
	Get(ctx context.Context, groupID, userID int, withLock bool) (entity.GroupParticipant, error)
	Create(ctx context.Context, p *entity.GroupParticipant) error
	Update(ctx context.Context, p *entity.GroupParticipant) error
}

//go:generate mockery --inpackage --testonly --case underscore --name GroupParticipantEventProducer
type GroupParticipantEventProducer interface {
	Produce(ctx context.Context, event entity.ParticipantEvent) error
}

//go:generate mockery --inpackage --testonly --case underscore --name TransactionManager
type TransactionManager interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

type StatusMatrix interface {
	IsCorrectTransit(from, to entity.GroupParticipantStatus) bool
}

type GroupParticipantConfig struct {
	TxManager     TransactionManager
	Repository    GroupParticipantRepository
	EventProducer GroupParticipantEventProducer
}

type GroupParticipant struct {
	txm  TransactionManager
	repo GroupParticipantRepository
	prod GroupParticipantEventProducer
}

func NewGroupParticipant(conf GroupParticipantConfig) *GroupParticipant {
	return &GroupParticipant{
		txm:  conf.TxManager,
		repo: conf.Repository,
		prod: conf.EventProducer,
	}
}

func (p *GroupParticipant) List(ctx context.Context, groupID int) ([]entity.GroupParticipant, error) {
	participants, err := p.repo.List(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("list of group participants: %w", err)
	}

	var isCurUserInGroup bool
	curUserID := ctxutil.UserIDFromContext(ctx).ToInt()

	for _, participant := range participants {
		if participant.UserID == curUserID {
			isCurUserInGroup = participant.IsInGroup()
			break
		}
	}

	if !isCurUserInGroup {
		return nil, fmt.Errorf("%w: current participant isn't in the group", entity.ErrGroupNotFound)
	}
	return participants, nil
}

func (p *GroupParticipant) Get(ctx context.Context, groupID, userID int) (entity.GroupParticipant, error) {
	curUserID := ctxutil.UserIDFromContext(ctx).ToInt()
	if err := p.checkPermission(ctx, groupID, curUserID, false); err != nil {
		return entity.GroupParticipant{}, fmt.Errorf("check permission: %w", err)
	}

	participant, err := p.repo.Get(ctx, groupID, userID, false)
	if err != nil {
		return entity.GroupParticipant{}, fmt.Errorf("get group participant: %w", err)
	}
	return participant, nil
}

func (p *GroupParticipant) Invite(ctx context.Context, groupID, userID int) (entity.GroupParticipant, error) {
	curUserID := ctxutil.UserIDFromContext(ctx).ToInt()
	if err := p.checkPermission(ctx, groupID, curUserID, true); err != nil {
		return entity.GroupParticipant{}, fmt.Errorf("check permission: %w", err)
	}

	invitedParticipant := entity.GroupParticipant{
		GroupID: groupID,
		UserID:  userID,
		Status:  entity.JoinedStatus,
	}
	if err := p.repo.Create(ctx, &invitedParticipant); err != nil {
		return entity.GroupParticipant{}, fmt.Errorf("create participant: %w", err)
	}

	err := p.prod.Produce(ctx, entity.ParticipantEvent{
		Type: entity.AddedParticipant,
		ChatID: entity.ChatID{
			ID:   groupID,
			Type: entity.GroupChatType,
		},
		UserID: userID,
	})
	if err != nil {
		return entity.GroupParticipant{}, fmt.Errorf("produce group participant event: %w", err)
	}

	return invitedParticipant, nil
}

func (p *GroupParticipant) UpdateStatus(ctx context.Context, groupID, userID int, status entity.GroupParticipantStatus) error {
	statusMatrix := entity.MxActionOnOneself
	curUserID := ctxutil.UserIDFromContext(ctx).ToInt()
	actionOnSomeone := curUserID != userID

	if actionOnSomeone {
		if err := p.checkPermission(ctx, groupID, curUserID, true); err != nil {
			return fmt.Errorf("check permission: %w", err)
		}

		statusMatrix = entity.MxActionOnSomeone
	}

	err := p.txm.Do(ctx, func(ctx context.Context) error {
		participant, err := p.repo.Get(ctx, groupID, userID, true)
		if err != nil {
			return fmt.Errorf("get group participant: %w", err)
		}

		if !statusMatrix.IsCorrectTransit(participant.Status, status) {
			return fmt.Errorf("%w: transit from %s to %s", entity.ErrIncorrectGroupParticipantStatusTransit, participant.Status, status)
		}

		participant.Status = status
		if err = p.repo.Update(ctx, &participant); err != nil {
			return fmt.Errorf("update group participant: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("call transaction manager: %w", err)
	}

	var eventType entity.ParticipantEventType
	switch status {
	case entity.JoinedStatus:
		eventType = entity.AddedParticipant
	case entity.KickedStatus, entity.LeftStatus:
		eventType = entity.RemovedParticipant
	}

	err = p.prod.Produce(ctx, entity.ParticipantEvent{
		Type: eventType,
		ChatID: entity.ChatID{
			ID:   groupID,
			Type: entity.GroupChatType,
		},
		UserID: userID,
	})
	if err != nil {
		return fmt.Errorf("produce group participant event: %w", err)
	}

	return nil
}

func (p *GroupParticipant) checkPermission(ctx context.Context, groupID, userID int, checkAdmin bool) error {
	curParticipant, err := p.repo.Get(ctx, groupID, userID, false)
	if err != nil {
		if errors.Is(err, entity.ErrGroupParticipantNotFound) {
			return fmt.Errorf("%w: current participant isn't in the group", entity.ErrGroupNotFound)
		}
		return fmt.Errorf("get current group participant: %w", err)
	}

	if !curParticipant.IsInGroup() {
		return fmt.Errorf("%w: current participant isn't in the group", entity.ErrGroupNotFound)
	}
	if checkAdmin && !curParticipant.IsAdmin {
		return fmt.Errorf("%w: current participant isn't admin in the group", entity.ErrForbiddenPerformAction)
	}

	return nil
}
