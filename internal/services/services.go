package services

import (
	"context"

	"github.com/Mort4lis/scht-backend/internal/domain"
)

type UserService interface {
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	Create(ctx context.Context, dto domain.CreateUserDTO) (*domain.User, error)
}

type ServiceContainer struct {
	User UserService
}
