package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Chatyx/backend/internal/dto"
	"github.com/Chatyx/backend/internal/entity"
	"github.com/Chatyx/backend/pkg/ctxutil"
	"github.com/Chatyx/backend/pkg/log"
)

type DialogRepository interface {
	List(ctx context.Context) ([]entity.Dialog, error)
	Create(ctx context.Context, dialog *entity.Dialog) error
	GetByID(ctx context.Context, id int) (entity.Dialog, error)
	Update(ctx context.Context, dialog *entity.Dialog) error
}

type DialogParticipantEventProducer interface {
	Produce(ctx context.Context, event entity.ParticipantEvent) error
}

type Dialog struct {
	repo DialogRepository
	prod DialogParticipantEventProducer
}

func NewDialog(repo DialogRepository, prod DialogParticipantEventProducer) *Dialog {
	return &Dialog{
		repo: repo,
		prod: prod,
	}
}

func (g *Dialog) List(ctx context.Context) ([]entity.Dialog, error) {
	dialogs, err := g.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list of dialogs: %w", err)
	}

	return dialogs, nil
}

func (g *Dialog) Create(ctx context.Context, obj dto.DialogCreate) (entity.Dialog, error) {
	dialog := entity.Dialog{
		Partner: entity.DialogPartner{
			UserID: obj.PartnerUserID,
		},
		CreatedAt: time.Now(),
	}

	if err := g.repo.Create(ctx, &dialog); err != nil {
		return entity.Dialog{}, fmt.Errorf("create dialog: %w", err)
	}

	chatID := entity.ChatID{ID: dialog.ID, Type: entity.DialogChatType}
	events := []entity.ParticipantEvent{
		{
			Type:   entity.AddedParticipant,
			ChatID: chatID,
			UserID: ctxutil.UserIDFromContext(ctx).ToInt(),
		},
		{
			Type:   entity.AddedParticipant,
			ChatID: chatID,
			UserID: obj.PartnerUserID,
		},
	}
	for _, event := range events {
		if err := g.prod.Produce(ctx, event); err != nil {
			return entity.Dialog{}, fmt.Errorf("produce dialog participant event: %w", err)
		}
	}

	return dialog, nil
}

func (g *Dialog) GetByID(ctx context.Context, id int) (entity.Dialog, error) {
	dialog, err := g.repo.GetByID(ctx, id)
	if err != nil {
		return entity.Dialog{}, fmt.Errorf("get dialog by id: %w", err)
	}

	return dialog, nil
}

func (g *Dialog) Update(ctx context.Context, obj dto.DialogUpdate) error {
	if obj.PartnerIsBlocked == nil {
		log.FromContext(ctx).Debug("no need to update dialog")
		return nil
	}

	dialog := entity.Dialog{
		ID: obj.ID,
		Partner: entity.DialogPartner{
			IsBlocked: *obj.PartnerIsBlocked,
		},
	}

	if err := g.repo.Update(ctx, &dialog); err != nil {
		return fmt.Errorf("update dialog: %w", err)
	}

	eventType := entity.RemovedParticipant
	if !*obj.PartnerIsBlocked {
		eventType = entity.AddedParticipant
	}

	event := entity.ParticipantEvent{
		Type: eventType,
		ChatID: entity.ChatID{
			ID:   dialog.ID,
			Type: entity.DialogChatType,
		},
		UserID: dialog.Partner.UserID,
	}
	if err := g.prod.Produce(ctx, event); err != nil {
		return fmt.Errorf("produce dialog participant event: %w", err)
	}

	return nil
}
