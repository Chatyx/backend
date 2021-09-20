package postgres

import (
	"context"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/jackc/pgx/v4/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool

	logger logging.Logger
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		pool:   pool,
		logger: logging.GetLogger(),
	}
}

func (r *UserRepository) List(ctx context.Context) ([]domain.User, error) {
	panic("implement me")
}

func (r *UserRepository) Create(ctx context.Context, user domain.User) error {
	panic("implement me")
}

func (r *UserRepository) GetByID(ctx context.Context, id string) domain.User {
	panic("implement me")
}

func (r *UserRepository) Update(ctx context.Context, user domain.User) error {
	panic("implement me")
}
