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

func (s *userService) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	return s.repo.GetByUsername(ctx, username)
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
