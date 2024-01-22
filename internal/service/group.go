package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Chatyx/backend/internal/dto"
	"github.com/Chatyx/backend/internal/entity"
)

type GroupRepository interface {
	List(ctx context.Context) ([]entity.Group, error)
	Create(ctx context.Context, group *entity.Group) error
	GetByID(ctx context.Context, id int) (entity.Group, error)
	Update(ctx context.Context, group *entity.Group) error
	Delete(ctx context.Context, id int) error
}

type Group struct {
	repo GroupRepository
}

func NewGroup(repo GroupRepository) *Group {
	return &Group{repo: repo}
}

func (g *Group) List(ctx context.Context) ([]entity.Group, error) {
	groups, err := g.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list of groups: %w", err)
	}

	return groups, nil
}

func (g *Group) Create(ctx context.Context, obj dto.GroupCreate) (entity.Group, error) {
	group := entity.Group{
		Name:        obj.Name,
		Description: obj.Description,
		CreatedAt:   time.Now(),
	}

	if err := g.repo.Create(ctx, &group); err != nil {
		return entity.Group{}, fmt.Errorf("create group: %w", err)
	}
	return group, nil
}

func (g *Group) GetByID(ctx context.Context, id int) (entity.Group, error) {
	group, err := g.repo.GetByID(ctx, id)
	if err != nil {
		return entity.Group{}, fmt.Errorf("get group by id: %w", err)
	}

	return group, nil
}

func (g *Group) Update(ctx context.Context, obj dto.GroupUpdate) (entity.Group, error) {
	group := entity.Group{
		ID:          obj.ID,
		Name:        obj.Name,
		Description: obj.Description,
	}

	if err := g.repo.Update(ctx, &group); err != nil {
		return entity.Group{}, fmt.Errorf("update group: %w", err)
	}
	return group, nil
}

func (g *Group) Delete(ctx context.Context, id int) error {
	if err := g.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete group: %w", err)
	}

	return nil
}
