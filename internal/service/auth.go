package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/internal/repository"
	"github.com/Mort4lis/scht-backend/pkg/auth"
	pkgErrors "github.com/Mort4lis/scht-backend/pkg/errors"
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
	}
}

func (s *authService) SignIn(ctx context.Context, dto domain.SignInDTO) (domain.JWTPair, error) {
	user, err := s.userService.GetByUsername(ctx, dto.Username)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return domain.JWTPair{}, fmt.Errorf("failed to sign-in user due %v: %w", err, domain.ErrWrongCredentials)
		}

		return domain.JWTPair{}, fmt.Errorf("failed to get user by username: %w", err)
	}

	ctxFields := pkgErrors.ContextFields{"session.user_id": user.ID}
	logger := logging.GetLoggerFromContext(ctx).WithFields(logging.Fields(ctxFields))
	ctx = logging.NewContextFromLogger(ctx, logger)

	if !s.hasher.CompareHashAndPassword(user.Password, dto.Password) {
		return domain.JWTPair{}, pkgErrors.WrapInContextError(domain.ErrWrongCredentials, "failed to sign-in user due password mismatch", ctxFields)
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
		return domain.JWTPair{}, pkgErrors.WrapInContextError(err, "an error occurred while creating a new access token", ctxFields)
	}

	refreshToken, err := s.tokenManager.NewRefreshToken()
	if err != nil {
		return domain.JWTPair{}, pkgErrors.WrapInContextError(err, "an error occurred while creating a new refresh token", ctxFields)
	}

	session := domain.Session{
		UserID:       user.ID,
		RefreshToken: refreshToken,
		Fingerprint:  dto.Fingerprint,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(s.refreshTokenTTL),
	}
	if err = s.sessionRepo.Set(ctx, session, s.refreshTokenTTL); err != nil {
		return domain.JWTPair{}, pkgErrors.WrapInContextError(err, "failed to set session", ctxFields)
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
			return domain.JWTPair{}, fmt.Errorf("failed to refresh token due %v: %w", err, domain.ErrInvalidRefreshToken)
		}

		return domain.JWTPair{}, fmt.Errorf("failed to get session: %w", err)
	}

	defer func() {
		_ = s.sessionRepo.Delete(ctx, dto.RefreshToken, session.UserID)
	}()

	ctxFields := pkgErrors.ContextFields{"session.user_id": session.UserID}
	logger := logging.GetLoggerFromContext(ctx).WithFields(logging.Fields(ctxFields))
	ctx = logging.NewContextFromLogger(ctx, logger)

	if dto.Fingerprint != session.Fingerprint {
		ctxFields["fingerprint"] = dto.Fingerprint
		ctxFields["session.fingerprint"] = session.Fingerprint
		ctxFields["session.refresh_token"] = session.RefreshToken

		logger.WithFields(logging.Fields(ctxFields)).Warning("Refresh token is compromised!")

		return domain.JWTPair{}, pkgErrors.WrapInContextError(domain.ErrInvalidRefreshToken, "refresh token is compromised due fingerprints don't match", ctxFields)
	}

	user, err := s.userService.GetByID(ctx, session.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return domain.JWTPair{}, pkgErrors.WrapInContextError(domain.ErrInvalidRefreshToken, "failed to refresh token due user is not found", ctxFields)
		}

		return domain.JWTPair{}, pkgErrors.WrapInContextError(err, "failed to get user", ctxFields)
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
		return domain.JWTPair{}, pkgErrors.WrapInContextError(err, "an error occurred while creating a new access token", ctxFields)
	}

	newRefreshToken, err := s.tokenManager.NewRefreshToken()
	if err != nil {
		return domain.JWTPair{}, pkgErrors.WrapInContextError(err, "an error occurred while creating a new refresh token", ctxFields)
	}

	newSession := domain.Session{
		UserID:       user.ID,
		RefreshToken: newRefreshToken,
		Fingerprint:  dto.Fingerprint,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(s.refreshTokenTTL),
	}
	if err = s.sessionRepo.Set(ctx, newSession, s.refreshTokenTTL); err != nil {
		return domain.JWTPair{}, pkgErrors.WrapInContextError(err, "failed to set session", ctxFields)
	}

	return domain.JWTPair{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *authService) Authorize(accessToken string) (domain.Claims, error) {
	var claims domain.Claims

	if err := s.tokenManager.Parse(accessToken, &claims); err != nil {
		return claims, fmt.Errorf("failed to parse access token due %v: %w", err, domain.ErrInvalidAccessToken)
	}

	return claims, nil
}
