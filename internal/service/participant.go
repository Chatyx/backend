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
	if err := p.checkPermission(ctx, groupID, curUserID, false); err != nil {
		return entity.GroupParticipant{}, fmt.Errorf("check permission: %w", err)
	}

	participant, err := p.repo.Get(ctx, groupID, userID)
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

	// TODO create a service message and publish it

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

func (p *GroupParticipant) checkPermission(ctx context.Context, groupID, userID int, checkAdmin bool) error {
	curParticipant, err := p.repo.Get(ctx, groupID, userID)
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