package services

import (
	"context"
	"errors"
	"time"

	"github.com/Mort4lis/scht-backend/pkg/auth"
	"github.com/dgrijalva/jwt-go/v4"
	"github.com/google/uuid"

	"github.com/Mort4lis/scht-backend/pkg/hasher"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/pkg/logging"
)

type authService struct {
	userService  UserService
	hasher       hasher.PasswordHasher
	tokenManager *auth.TokenManager

	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration

	logger logging.Logger
}

type AuthServiceConfig struct {
	UserService  UserService
	Hasher       hasher.PasswordHasher
	TokenManager *auth.TokenManager

	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func NewAuthService(cfg AuthServiceConfig) AuthService {
	return &authService{
		userService:     cfg.UserService,
		hasher:          cfg.Hasher,
		tokenManager:    cfg.TokenManager,
		accessTokenTTL:  cfg.AccessTokenTTL,
		refreshTokenTTL: cfg.RefreshTokenTTL,
		logger:          logging.GetLogger(),
	}
}

func (s *authService) SignIn(ctx context.Context, dto domain.SignInDTO) (domain.JWTPair, error) {
	user, err := s.userService.GetByUsername(ctx, dto.Username)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			s.logger.Debugf("failed to login user %q: user doesn't exists", dto.Username)
			return domain.JWTPair{}, domain.ErrInvalidCredentials
		}

		return domain.JWTPair{}, err
	}

	if !s.hasher.CompareHashAndPassword(user.Password, dto.Password) {
		s.logger.Debugf("failed to login user %q: password mismatch", dto.Username)
		return domain.JWTPair{}, domain.ErrInvalidCredentials
	}

	claims := &domain.Claims{
		StandardClaims: jwt.StandardClaims{
			ID:        uuid.New().String(),
			Subject:   user.ID,
			Issuer:    "SCHT",
			IssuedAt:  jwt.At(time.Now()),
			ExpiresAt: jwt.At(time.Now().Add(s.accessTokenTTL)),
		},
	}

	accessToken, err := s.tokenManager.NewAccessToken(claims)
	if err != nil {
		s.logger.WithError(err).Error("Error occurred while creating a new access token")
		return domain.JWTPair{}, err
	}

	refreshToken, err := s.tokenManager.NewRefreshToken()
	if err != nil {
		s.logger.WithError(err).Error("Error occurred while creating a new refresh token")
	}

	// TODO: save refresh token to storage

	return domain.JWTPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *authService) Refresh(ctx context.Context, refreshToken string) (domain.JWTPair, error) {
	panic("implement me")
}

func (s *authService) Authorize(ctx context.Context, accessToken string) (*domain.User, error) {
	var claims domain.Claims

	if err := s.tokenManager.Parse(accessToken, &claims); err != nil {
		s.logger.WithError(err).Debug("invalid access token")
		return nil, domain.ErrInvalidToken
	}

	user, err := s.userService.GetByID(ctx, claims.Subject)
	if err != nil {
		s.logger.WithError(err).Errorf(
			"Error occurred while getting authorized user with id = %s",
			claims.Subject,
		)

		return nil, err
	}

	return user, nil
}
