package service

import (
	"context"

	"github.com/Mort4lis/scht-backend/pkg/hasher"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/internal/repository"
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
	return s.userRepo.List(ctx)
}

func (s *userService) Create(ctx context.Context, dto domain.CreateUserDTO) (domain.User, error) {
	hash, err := s.hasher.Hash(dto.Password)
	if err != nil {
		s.logger.WithError(err).Error("An error occurred while hashing password")
		return domain.User{}, err
	}

	dto.Password = hash

	return s.userRepo.Create(ctx, dto)
}

func (s *userService) GetByID(ctx context.Context, id string) (domain.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *userService) GetByUsername(ctx context.Context, username string) (domain.User, error) {
	return s.userRepo.GetByUsername(ctx, username)
}

func (s *userService) Update(ctx context.Context, dto domain.UpdateUserDTO) (domain.User, error) {
	return s.userRepo.Update(ctx, dto)
}

func (s *userService) UpdatePassword(ctx context.Context, dto domain.UpdateUserPasswordDTO) error {
	user, err := s.GetByID(ctx, dto.UserID)
	if err != nil {
		return err
	}

	if !s.hasher.CompareHashAndPassword(user.Password, dto.Current) {
		s.logger.Debugf("can't update password for user_id = %s due passed current password is wrong", dto.UserID)
		return domain.ErrWrongCurrentPassword
	}

	hash, err := s.hasher.Hash(dto.New)
	if err != nil {
		s.logger.WithError(err).Error("An error occurred while hashing new password")
		return err
	}

	if err = s.sessionRepo.DeleteAllByUserID(ctx, user.ID); err != nil {
		return err
	}

	return s.userRepo.UpdatePassword(ctx, dto.UserID, hash)
}

func (s *userService) Delete(ctx context.Context, id string) error {
	return s.userRepo.Delete(ctx, id)
}
