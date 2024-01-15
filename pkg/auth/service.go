package auth

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/Chatyx/backend/pkg/auth/model"
	"github.com/Chatyx/backend/pkg/auth/storage"
	"github.com/Chatyx/backend/pkg/token"
)

const (
	defaultRefreshTokenSize = 32
)

const (
	defaultAccessTokenTTL = 15 * time.Minute
	defaultRefreshTokeTTL = 60 * 24 * time.Hour
)

type CheckPasswordFunc func(user, password string) (userID string, ok bool, err error)

type EnrichClaimsFunc func(claims token.Claims)

type ServiceConfig struct {
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
	Issuer          string
	SignedKey       any
	CheckPassword   CheckPasswordFunc
	EnrichClaims    EnrichClaimsFunc
	Logger          Logger
}

type ServiceOption func(conf *ServiceConfig)

func WithAccessTokenTTL(d time.Duration) ServiceOption {
	return func(conf *ServiceConfig) {
		conf.AccessTokenTTL = d
	}
}

func WithRefreshTokenTTL(d time.Duration) ServiceOption {
	return func(conf *ServiceConfig) {
		conf.RefreshTokenTTL = d
	}
}

func WithIssuer(iss string) ServiceOption {
	return func(conf *ServiceConfig) {
		conf.Issuer = iss
	}
}

func WithSignedKey(key any) ServiceOption {
	return func(conf *ServiceConfig) {
		conf.SignedKey = key
	}
}

func WithCheckPassword(fn CheckPasswordFunc) ServiceOption {
	return func(conf *ServiceConfig) {
		conf.CheckPassword = fn
	}
}

func WithEnrichClaims(fn EnrichClaimsFunc) ServiceOption {
	return func(conf *ServiceConfig) {
		conf.EnrichClaims = fn
	}
}

func WithLogger(logger Logger) ServiceOption {
	return func(conf *ServiceConfig) {
		conf.Logger = logger
	}
}

type Meta struct {
	IP net.IP
}

type MetaOption func(meta *Meta)

func WithIP(ip net.IP) MetaOption {
	return func(meta *Meta) {
		meta.IP = ip
	}
}

type SessionStorage interface {
	Set(ctx context.Context, sess model.Session) error
	GetWithDelete(ctx context.Context, refreshToken string) (model.Session, error)
}

type Service struct {
	issuer              string
	checkPassword       CheckPasswordFunc
	enrichClaims        EnrichClaimsFunc
	storage             SessionStorage
	accessTokenManager  token.JWT
	refreshTokenManager token.Hex
	refreshTokenTTL     time.Duration
	logger              Logger
}

func NewService(storage SessionStorage, opts ...ServiceOption) *Service {
	conf := &ServiceConfig{
		AccessTokenTTL:  defaultAccessTokenTTL,
		RefreshTokenTTL: defaultRefreshTokeTTL,
		Issuer:          "pkg/auth",
		CheckPassword: func(user, password string) (userID string, ok bool, err error) {
			if user == "root" && password == "root" {
				return "1", true, nil
			}
			return "", false, nil
		},
		Logger: noOpLogger{},
	}
	for _, opt := range opts {
		opt(conf)
	}

	s := &Service{
		refreshTokenTTL:     conf.RefreshTokenTTL,
		issuer:              conf.Issuer,
		checkPassword:       conf.CheckPassword,
		enrichClaims:        conf.EnrichClaims,
		logger:              conf.Logger,
		storage:             storage,
		accessTokenManager:  token.NewJWT(conf.Issuer, conf.SignedKey, conf.AccessTokenTTL),
		refreshTokenManager: token.Hex{},
	}

	return s
}

func (s *Service) Login(ctx context.Context, cred model.Credentials, opts ...MetaOption) (model.TokenPair, error) {
	var pair model.TokenPair

	userID, ok, err := s.checkPassword(cred.Username, cred.Password)
	if err != nil {
		return pair, fmt.Errorf("check password: %w", err)
	}
	if !ok {
		s.logger.Warnf("User `%s` failed to login", cred.Username)
		return pair, ErrUserNotFound
	}

	var claims token.Claims
	if s.enrichClaims != nil {
		s.enrichClaims(claims)
	}

	accessToken, err := s.accessTokenManager.Token(userID, claims)
	if err != nil {
		return pair, fmt.Errorf("create access token: %w", err)
	}

	refreshToken, err := s.refreshTokenManager.Token(defaultRefreshTokenSize)
	if err != nil {
		return pair, fmt.Errorf("create refresh token: %w", err)
	}

	meta := &Meta{}
	for _, opt := range opts {
		opt(meta)
	}

	sess := model.Session{
		UserID:       userID,
		RefreshToken: refreshToken,
		Fingerprint:  cred.Fingerprint,
		IP:           meta.IP,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Now().Add(s.refreshTokenTTL),
	}
	if err = s.storage.Set(ctx, sess); err != nil {
		return pair, fmt.Errorf("set session to storage: %w", err)
	}

	s.logger.Infof("User `%s` login successfully", cred.Username)

	return model.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	if _, err := s.storage.GetWithDelete(ctx, refreshToken); err != nil {
		if errors.Is(err, storage.ErrSessionNotFound) {
			return fmt.Errorf("%w: %v", ErrInvalidRefreshToken, storage.ErrSessionNotFound)
		}
		return fmt.Errorf("delete session: %w", err)
	}

	return nil
}

func (s *Service) RefreshSession(ctx context.Context, rs model.RefreshSession, opts ...MetaOption) (model.TokenPair, error) {
	var pair model.TokenPair

	sess, err := s.storage.GetWithDelete(ctx, rs.RefreshToken)
	if err != nil {
		if errors.Is(err, storage.ErrSessionNotFound) {
			return pair, fmt.Errorf("%w: %v", ErrInvalidRefreshToken, storage.ErrSessionNotFound)
		}
		return pair, fmt.Errorf("get session with delete: %w", err)
	}
	if sess.ExpiresAt.Before(time.Now()) {
		return pair, fmt.Errorf("%w: session is expired", ErrInvalidRefreshToken)
	}

	if sess.Fingerprint != rs.Fingerprint {
		s.logger.Warnf("Fingerprints don't match for user with id `%s` while refreshing session", sess.UserID)
		return pair, fmt.Errorf("%w: fingerprints don't match", ErrInvalidRefreshToken)
	}

	var claims token.Claims
	if s.enrichClaims != nil {
		s.enrichClaims(claims)
	}

	accessToken, err := s.accessTokenManager.Token(sess.UserID, claims)
	if err != nil {
		return pair, fmt.Errorf("create access token: %w", err)
	}

	refreshToken, err := s.refreshTokenManager.Token(defaultRefreshTokenSize)
	if err != nil {
		return pair, fmt.Errorf("create refresh token: %w", err)
	}

	meta := &Meta{}
	for _, opt := range opts {
		opt(meta)
	}

	sess.IP = meta.IP
	sess.RefreshToken = refreshToken
	sess.CreatedAt = time.Now()
	sess.ExpiresAt = time.Now().Add(s.refreshTokenTTL)

	if err = s.storage.Set(ctx, sess); err != nil {
		return pair, fmt.Errorf("set session to storage: %w", err)
	}

	return model.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
