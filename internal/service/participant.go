package service

import (
	"context"
	"fmt"

	"github.com/Chatyx/backend/internal/entity"
	"github.com/Chatyx/backend/pkg/ctxutil"
)

type GroupParticipantFunc func(p *entity.GroupParticipant) error

type GroupParticipantRepository interface {
	List(ctx context.Context, groupID int) ([]entity.GroupParticipant, error)
	Get(ctx context.Context, groupID, userID int) (entity.GroupParticipant, error)
	GetThenUpdate(ctx context.Context, groupID, userID int, fn GroupParticipantFunc) error
	Create(ctx context.Context, p *entity.GroupParticipant) error
}

type StatusMatrix interface {
	IsCorrectTransit(from, to entity.GroupParticipantStatus) bool
}

type GroupParticipant struct {
	repo GroupParticipantRepository
}

func NewGroupParticipant(repo GroupParticipantRepository) *GroupParticipant {
	return &GroupParticipant{repo: repo}
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
	curParticipant, err := p.repo.Get(ctx, groupID, curUserID)
	if err != nil {
		return entity.GroupParticipant{}, fmt.Errorf("get current group participant: %w", err)
	}

	if !curParticipant.IsInGroup() {
		return entity.GroupParticipant{}, fmt.Errorf("%w: current participant isn't in the group", entity.ErrGroupNotFound)
	}

	participant, err := p.repo.Get(ctx, groupID, userID)
	if err != nil {
		return entity.GroupParticipant{}, fmt.Errorf("get group participant: %w", err)
	}
	return participant, nil
}

func (p *GroupParticipant) Invite(ctx context.Context, groupID, userID int) (entity.GroupParticipant, error) {
	curUserID := ctxutil.UserIDFromContext(ctx).ToInt()
	curParticipant, err := p.repo.Get(ctx, groupID, curUserID)
	if err != nil {
		return entity.GroupParticipant{}, fmt.Errorf("get current group participant: %w", err)
	}

	if !curParticipant.IsInGroup() {
		return entity.GroupParticipant{}, fmt.Errorf("%w: current participant isn't in the group", entity.ErrGroupNotFound)
	}
	if !curParticipant.IsAdmin {
		return entity.GroupParticipant{}, fmt.Errorf("%w: current participant isn't admin in the group", entity.ErrForbiddenPerformAction)
	}

	invitedParticipant := entity.GroupParticipant{
		GroupID: groupID,
		UserID:  userID,
		Status:  entity.JoinedStatus,
	}
	if err = p.repo.Create(ctx, &invitedParticipant); err != nil {
		return entity.GroupParticipant{}, fmt.Errorf("create participant: %w", err)
	}

	// TODO create a service message and publish it

	return invitedParticipant, nil
}

func (p *GroupParticipant) UpdateStatus(ctx context.Context, groupID, userID int, status entity.GroupParticipantStatus) error {
	statusMatrix := entity.MxActionOnOneself
	curUserID := ctxutil.UserIDFromContext(ctx).ToInt()
	actionOnSomeone := curUserID != userID

	if actionOnSomeone {
		curParticipant, err := p.repo.Get(ctx, groupID, curUserID)
		if err != nil {
			return fmt.Errorf("get current group participant: %w", err)
		}

		if !curParticipant.IsInGroup() {
			return fmt.Errorf("%w: current participant isn't in the group", entity.ErrGroupNotFound)
		}
		if !curParticipant.IsAdmin {
			return fmt.Errorf("%w: current participant isn't admin in the group", entity.ErrForbiddenPerformAction)
		}

		statusMatrix = entity.MxActionOnSomeone
	}

	err := p.repo.GetThenUpdate(ctx, groupID, userID, func(participant *entity.GroupParticipant) error {
		if !statusMatrix.IsCorrectTransit(participant.Status, status) {
			return fmt.Errorf("%w: transit from %s to %s", entity.ErrIncorrectGroupParticipantStatusTransit, participant.Status, status)
		}

		participant.Status = status
		return nil
	})
	if err != nil {
		return fmt.Errorf("get participant for the next update: %w", err)
	}

	// TODO create a service message and publish it

	return nil
}
