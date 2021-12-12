package service

import (
	"context"
	"fmt"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/internal/repository"
	"github.com/Mort4lis/scht-backend/pkg/hasher"
	"github.com/Mort4lis/scht-backend/pkg/logging"
)

type userService struct {
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
	hasher      hasher.PasswordHasher
	logger      logging.Logger
}

func NewUserService(userRepo repository.UserRepository, sessionRepo repository.SessionRepository, hasher hasher.PasswordHasher) UserService {
	return &userService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		hasher:      hasher,
		logger:      logging.GetLogger(),
	}
}

func (s *userService) List(ctx context.Context) ([]domain.User, error) {
	users, err := s.userRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get list of users: %w", err)
	}

	return users, nil
}

func (s *userService) Create(ctx context.Context, dto domain.CreateUserDTO) (domain.User, error) {
	hash, err := s.hasher.Hash(dto.Password)
	if err != nil {
		return domain.User{}, fmt.Errorf("failed to hashing password: %w", err)
	}

	dto.Password = hash

	user, err := s.userRepo.Create(ctx, dto)
	if err != nil {
		return domain.User{}, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *userService) GetByID(ctx context.Context, userID string) (domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return domain.User{}, fmt.Errorf("failed to get user by id: %w", err)
	}

	return user, nil
}

func (s *userService) GetByUsername(ctx context.Context, username string) (domain.User, error) {
	user, err := s.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return domain.User{}, fmt.Errorf("failed to get user by username: %w", err)
	}

	return user, nil
}

func (s *userService) Update(ctx context.Context, dto domain.UpdateUserDTO) (domain.User, error) {
	user, err := s.userRepo.Update(ctx, dto)
	if err != nil {
		return domain.User{}, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

func (s *userService) UpdatePassword(ctx context.Context, dto domain.UpdateUserPasswordDTO) error {
	user, err := s.GetByID(ctx, dto.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user by uuid: %w", err)
	}

	if !s.hasher.CompareHashAndPassword(user.Password, dto.Current) {
		return fmt.Errorf("can't update password: %w", domain.ErrWrongCurrentPassword)
	}

	hash, err := s.hasher.Hash(dto.New)
	if err != nil {
		return fmt.Errorf("failed to hashing password: %w", err)
	}

	if err = s.sessionRepo.DeleteAllByUserID(ctx, user.ID); err != nil {
		return fmt.Errorf("failed to delete all user sessions: %w", err)
	}

	if err = s.userRepo.UpdatePassword(ctx, dto.UserID, hash); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

func (s *userService) Delete(ctx context.Context, userID string) error {
	if err := s.userRepo.Delete(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if err := s.sessionRepo.DeleteAllByUserID(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete all user sessions: %w", err)
	}

	return nil
}
