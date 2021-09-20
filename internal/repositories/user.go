package repositories

import (
	"context"

	"github.com/Mort4lis/scht-backend/internal/domain"
)

type UserRepository interface {
	List(ctx context.Context) ([]domain.User, error)
	Create(ctx context.Context, user domain.User) error
	GetByID(ctx context.Context, id string) domain.User
	Update(ctx context.Context, user domain.User) error
}
