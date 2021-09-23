package services

import (
	"context"

	"github.com/Mort4lis/scht-backend/internal/domain"
)

type UserService interface {
	List(ctx context.Context) ([]*domain.User, error)
	Create(ctx context.Context, dto domain.CreateUserDTO) (*domain.User, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	Update(ctx context.Context, dto domain.UpdateUserDTO) (*domain.User, error)
	Delete(ctx context.Context, id string) error
}

type AuthService interface {
	SignIn(ctx context.Context, dto domain.SignInDTO) (domain.JWTPair, error)
	Refresh(ctx context.Context, refreshToken string) (domain.JWTPair, error)
	Authorize(accessToken string) (domain.Claims, error)
}

type ServiceContainer struct {
	User UserService
	Auth AuthService
}
