package services

import (
	"context"

	"github.com/Mort4lis/scht-backend/pkg/hasher"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/internal/repositories"
	"github.com/Mort4lis/scht-backend/pkg/logging"
)

type userService struct {
	repo   repositories.UserRepository
	hasher hasher.PasswordHasher

	logger logging.Logger
}

func NewUserService(repo repositories.UserRepository, hasher hasher.PasswordHasher) UserService {
	return &userService{
		repo:   repo,
		hasher: hasher,
		logger: logging.GetLogger(),
	}
}

func (s *userService) List(ctx context.Context) ([]*domain.User, error) {
	return s.repo.List(ctx)
}

func (s *userService) Create(ctx context.Context, dto domain.CreateUserDTO) (*domain.User, error) {
	hash, err := s.hasher.Hash(dto.Password)
	if err != nil {
		s.logger.WithError(err).Error("Error occurred while hashing password")

		return nil, err
	}

	dto.Password = hash

	return s.repo.Create(ctx, dto)
}

func (s *userService) GetByID(ctx context.Context, id string) (*domain.User, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *userService) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	return s.repo.GetByUsername(ctx, username)
}

func (s *userService) Update(ctx context.Context, dto domain.UpdateUserDTO) (*domain.User, error) {
	return s.repo.Update(ctx, dto)
}

func (s *userService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
