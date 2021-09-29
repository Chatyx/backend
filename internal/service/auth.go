package service

import (
	"context"
	"errors"
	"time"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/internal/repository"
	"github.com/Mort4lis/scht-backend/pkg/auth"
	"github.com/Mort4lis/scht-backend/pkg/hasher"
	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/dgrijalva/jwt-go/v4"
	"github.com/google/uuid"
)

type authService struct {
	userService UserService
	sessionRepo repository.SessionRepository

	hasher       hasher.PasswordHasher
	tokenManager auth.TokenManager

	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration

	logger logging.Logger
}

type AuthServiceConfig struct {
	UserService UserService
	SessionRepo repository.SessionRepository

	Hasher       hasher.PasswordHasher
	TokenManager auth.TokenManager

	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func NewAuthService(cfg AuthServiceConfig) AuthService {
	return &authService{
		userService:     cfg.UserService,
		sessionRepo:     cfg.SessionRepo,
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
			return domain.JWTPair{}, domain.ErrWrongCredentials
		}

		return domain.JWTPair{}, err
	}

	if !s.hasher.CompareHashAndPassword(user.Password, dto.Password) {
		s.logger.Debugf("failed to login user %q: password mismatch", dto.Username)
		return domain.JWTPair{}, domain.ErrWrongCredentials
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
		return domain.JWTPair{}, err
	}

	session := domain.Session{
		UserID:       user.ID,
		RefreshToken: refreshToken,
		Fingerprint:  dto.Fingerprint,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(s.refreshTokenTTL),
	}
	if err = s.sessionRepo.Set(ctx, refreshToken, session); err != nil {
		return domain.JWTPair{}, err
	}

	return domain.JWTPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *authService) Refresh(ctx context.Context, dto domain.RefreshSessionDTO) (domain.JWTPair, error) {
	session, err := s.sessionRepo.Get(ctx, dto.RefreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrSessionNotFound) {
			return domain.JWTPair{}, domain.ErrInvalidRefreshToken
		}

		return domain.JWTPair{}, err
	}

	defer func() {
		_ = s.sessionRepo.Delete(ctx, dto.RefreshToken)
	}()

	if dto.Fingerprint != session.Fingerprint {
		s.logger.Warningf("Refresh token %s is compromised (fingerprints don't match)", dto.RefreshToken)
		return domain.JWTPair{}, domain.ErrInvalidRefreshToken
	}

	user, err := s.userService.GetByID(ctx, session.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return domain.JWTPair{}, domain.ErrInvalidRefreshToken
		}

		return domain.JWTPair{}, err
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

	newAccessToken, err := s.tokenManager.NewAccessToken(claims)
	if err != nil {
		s.logger.WithError(err).Error("Error occurred while creating a new access token")
		return domain.JWTPair{}, err
	}

	newRefreshToken, err := s.tokenManager.NewRefreshToken()
	if err != nil {
		s.logger.WithError(err).Error("Error occurred while creating a new refresh token")
		return domain.JWTPair{}, err
	}

	newSession := domain.Session{
		UserID:       user.ID,
		RefreshToken: newRefreshToken,
		Fingerprint:  dto.Fingerprint,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(s.refreshTokenTTL),
	}
	if err = s.sessionRepo.Set(ctx, newRefreshToken, newSession); err != nil {
		return domain.JWTPair{}, err
	}

	return domain.JWTPair{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *authService) Authorize(accessToken string) (domain.Claims, error) {
	var claims domain.Claims

	if err := s.tokenManager.Parse(accessToken, &claims); err != nil {
		s.logger.WithError(err).Debug("invalid access token")
		return claims, domain.ErrInvalidAccessToken
	}

	return claims, nil
}
