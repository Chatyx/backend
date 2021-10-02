package repository

import (
	"context"
	"time"

	"github.com/Mort4lis/scht-backend/internal/domain"
)

//go:generate mockgen -source=repository.go -destination=mocks/mock.go

type UserRepository interface {
	List(ctx context.Context) ([]domain.User, error)
	Create(ctx context.Context, dto domain.CreateUserDTO) (domain.User, error)
	GetByID(ctx context.Context, id string) (domain.User, error)
	GetByUsername(ctx context.Context, username string) (domain.User, error)
	Update(ctx context.Context, dto domain.UpdateUserDTO) (domain.User, error)
	Delete(ctx context.Context, id string) error
}

type SessionRepository interface {
	Get(ctx context.Context, key string) (domain.Session, error)
	Set(ctx context.Context, key string, session domain.Session, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}
