package service

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/Mort4lis/scht-backend/internal/domain"
	mockrepository "github.com/Mort4lis/scht-backend/internal/repository/mocks"
	mockservice "github.com/Mort4lis/scht-backend/internal/service/mocks"
	mockauth "github.com/Mort4lis/scht-backend/pkg/auth/mocks"
	mockhasher "github.com/Mort4lis/scht-backend/pkg/hasher/mocks"
	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/dgrijalva/jwt-go/v4"
	"github.com/golang/mock/gomock"
)

const (
	accessTokenTTL  = 15 * time.Minute    // 15 minutes
	refreshTokenTTL = 15 * 24 * time.Hour // 15 days
)

var (
	errUnexpected = errors.New("unexpected error")

	currentTime = time.Date(2021, time.September, 27, 11, 10, 12, 411, time.Local)

	defaultSignInDTO = domain.SignInDTO{
		Username:    "john1967",
		Password:    "qwerty12345",
		Fingerprint: "5dc49b7a-6153-4eae-9c0f-297655c45f08",
	}
	defaultRefreshSessionDTO = domain.RefreshSessionDTO{
		RefreshToken: "qGVFLRQw37TnSmG0LKFN",
		Fingerprint:  "5dc49b7a-6153-4eae-9c0f-297655c45f08",
	}
	defaultServiceUser = domain.User{
		ID:        "1",
		Username:  "john1967",
		Password:  "8743b52063cd84097a65d1633f5c74f5",
		Email:     "john1967@gmail.com",
		CreatedAt: &currentTime,
	}
	defaultRefreshSession = domain.Session{
		UserID:       "1",
		RefreshToken: "qGVFLRQw37TnSmG0LKFN",
		Fingerprint:  "5dc49b7a-6153-4eae-9c0f-297655c45f08",
		ExpiresAt:    currentTime.Add(refreshTokenTTL),
		CreatedAt:    currentTime,
	}
	defaultExpectedPair = domain.JWTPair{
		AccessToken:  "header.payload.sign",
		RefreshToken: "7TnSmG0LKFNqGVFLRQw3",
	}
	defaultExpectedClaims = domain.Claims{
		StandardClaims: jwt.StandardClaims{
			ID:        "cdc82f62-f4eb-4d54-90dd-55185e60bb3f",
			Subject:   "1",
			Issuer:    "SCHT",
			ExpiresAt: jwt.At(currentTime.Add(accessTokenTTL)),
			IssuedAt:  jwt.At(currentTime),
		},
	}
)

func TestAuthService_SignIn(t *testing.T) {
	type userServiceMockBehaviour func(us *mockservice.MockUserService, username string, returnedUser domain.User)
	type sessionRepoMockBehaviour func(r *mockrepository.MockSessionRepository, refreshToken string)
	type hasherMockBehaviour func(h *mockhasher.MockPasswordHasher, hash, password string)
	type tokenManagerMockBehaviour func(tm *mockauth.MockTokenManager, pair domain.JWTPair)

	testTable := []struct {
		name                      string
		signInDTO                 domain.SignInDTO
		serviceUser               domain.User
		userServiceMockBehaviour  userServiceMockBehaviour
		sessionRepoMockBehaviour  sessionRepoMockBehaviour
		hasherMockBehaviour       hasherMockBehaviour
		tokenManagerMockBehaviour tokenManagerMockBehaviour
		expectedPair              domain.JWTPair
		expectedErr               error
	}{
		{
			name:        "Success",
			signInDTO:   defaultSignInDTO,
			serviceUser: defaultServiceUser,
			userServiceMockBehaviour: func(us *mockservice.MockUserService, username string, returnedUser domain.User) {
				us.EXPECT().GetByUsername(context.Background(), username).Return(returnedUser, nil)
			},
			sessionRepoMockBehaviour: func(r *mockrepository.MockSessionRepository, refreshToken string) {
				r.EXPECT().Set(context.Background(), refreshToken, gomock.Any()).Return(nil)
			},
			hasherMockBehaviour: func(h *mockhasher.MockPasswordHasher, hash, password string) {
				h.EXPECT().CompareHashAndPassword(hash, password).Return(true)
			},
			tokenManagerMockBehaviour: func(tm *mockauth.MockTokenManager, pair domain.JWTPair) {
				tm.EXPECT().NewAccessToken(gomock.Any()).Return(pair.AccessToken, nil)
				tm.EXPECT().NewRefreshToken().Return(pair.RefreshToken, nil)
			},
			expectedPair: defaultExpectedPair,
			expectedErr:  nil,
		},
		{
			name:      "Wrong username",
			signInDTO: defaultSignInDTO,
			userServiceMockBehaviour: func(us *mockservice.MockUserService, username string, returnedUser domain.User) {
				us.EXPECT().GetByUsername(context.Background(), username).Return(domain.User{}, domain.ErrUserNotFound)
			},
			expectedErr: domain.ErrWrongCredentials,
		},
		{
			name:        "Wrong password",
			signInDTO:   defaultSignInDTO,
			serviceUser: defaultServiceUser,
			userServiceMockBehaviour: func(us *mockservice.MockUserService, username string, returnedUser domain.User) {
				us.EXPECT().GetByUsername(context.Background(), username).Return(returnedUser, nil)
			},
			hasherMockBehaviour: func(h *mockhasher.MockPasswordHasher, hash, password string) {
				h.EXPECT().CompareHashAndPassword(hash, password).Return(false)
			},
			expectedErr: domain.ErrWrongCredentials,
		},
		{
			name:      "Unexpected error while getting user by username",
			signInDTO: defaultSignInDTO,
			userServiceMockBehaviour: func(us *mockservice.MockUserService, username string, returnedUser domain.User) {
				us.EXPECT().GetByUsername(context.Background(), username).Return(domain.User{}, errUnexpected)
			},
			expectedErr: errUnexpected,
		},
		{
			name:        "Unexpected error while creating access token",
			signInDTO:   defaultSignInDTO,
			serviceUser: defaultServiceUser,
			userServiceMockBehaviour: func(us *mockservice.MockUserService, username string, returnedUser domain.User) {
				us.EXPECT().GetByUsername(context.Background(), username).Return(returnedUser, nil)
			},
			hasherMockBehaviour: func(h *mockhasher.MockPasswordHasher, hash, password string) {
				h.EXPECT().CompareHashAndPassword(hash, password).Return(true)
			},
			tokenManagerMockBehaviour: func(tm *mockauth.MockTokenManager, pair domain.JWTPair) {
				tm.EXPECT().NewAccessToken(gomock.Any()).Return("", errUnexpected)
			},
			expectedErr: errUnexpected,
		},
		{
			name:        "Unexpected error while creating refresh token",
			signInDTO:   defaultSignInDTO,
			serviceUser: defaultServiceUser,
			userServiceMockBehaviour: func(us *mockservice.MockUserService, username string, returnedUser domain.User) {
				us.EXPECT().GetByUsername(context.Background(), username).Return(returnedUser, nil)
			},
			hasherMockBehaviour: func(h *mockhasher.MockPasswordHasher, hash, password string) {
				h.EXPECT().CompareHashAndPassword(hash, password).Return(true)
			},
			tokenManagerMockBehaviour: func(tm *mockauth.MockTokenManager, pair domain.JWTPair) {
				tm.EXPECT().NewAccessToken(gomock.Any()).Return(pair.AccessToken, nil)
				tm.EXPECT().NewRefreshToken().Return("", errUnexpected)
			},
			expectedErr: errUnexpected,
		},
		{
			name:        "Unexpected error while setting session to repository",
			signInDTO:   defaultSignInDTO,
			serviceUser: defaultServiceUser,
			userServiceMockBehaviour: func(us *mockservice.MockUserService, username string, returnedUser domain.User) {
				us.EXPECT().GetByUsername(context.Background(), username).Return(returnedUser, nil)
			},
			hasherMockBehaviour: func(h *mockhasher.MockPasswordHasher, hash, password string) {
				h.EXPECT().CompareHashAndPassword(hash, password).Return(true)
			},
			tokenManagerMockBehaviour: func(tm *mockauth.MockTokenManager, pair domain.JWTPair) {
				tm.EXPECT().NewAccessToken(gomock.Any()).Return(pair.AccessToken, nil)
				tm.EXPECT().NewRefreshToken().Return(pair.RefreshToken, nil)
			},
			sessionRepoMockBehaviour: func(r *mockrepository.MockSessionRepository, refreshToken string) {
				r.EXPECT().Set(context.Background(), refreshToken, gomock.Any()).Return(errUnexpected)
			},
			expectedErr: errUnexpected,
		},
	}

	logging.InitLogger(logging.LogConfig{
		LoggerKind: "mock",
	})

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			us := mockservice.NewMockUserService(c)
			sessionRepo := mockrepository.NewMockSessionRepository(c)
			hasher := mockhasher.NewMockPasswordHasher(c)
			tokenManager := mockauth.NewMockTokenManager(c)

			as := NewAuthService(AuthServiceConfig{
				UserService:     us,
				SessionRepo:     sessionRepo,
				Hasher:          hasher,
				TokenManager:    tokenManager,
				AccessTokenTTL:  accessTokenTTL,
				RefreshTokenTTL: refreshTokenTTL,
			})

			if testCase.userServiceMockBehaviour != nil {
				testCase.userServiceMockBehaviour(us, testCase.signInDTO.Username, testCase.serviceUser)
			}

			if testCase.hasherMockBehaviour != nil {
				testCase.hasherMockBehaviour(hasher, testCase.serviceUser.Password, testCase.signInDTO.Password)
			}

			if testCase.tokenManagerMockBehaviour != nil {
				testCase.tokenManagerMockBehaviour(tokenManager, testCase.expectedPair)
			}

			if testCase.sessionRepoMockBehaviour != nil {
				testCase.sessionRepoMockBehaviour(sessionRepo, testCase.expectedPair.RefreshToken)
			}

			pair, err := as.SignIn(context.Background(), testCase.signInDTO)

			if testCase.expectedErr != nil && !errors.Is(testCase.expectedErr, err) {
				t.Errorf("Wrong returned error. Expected error %v, got %v", testCase.expectedErr, err)
			}

			if testCase.expectedErr == nil && err != nil {
				t.Errorf("Unexpected error %v", err)
			}

			if !reflect.DeepEqual(testCase.expectedPair, pair) {
				t.Errorf("Wrong token pair result. Expected %#v, got %#v", testCase.expectedPair, pair)
			}
		})
	}
}

func TestAuthService_Refresh(t *testing.T) {
	type sessionRepoMockBehaviour func(r *mockrepository.MockSessionRepository, oldRefreshToken string, newRefreshToken string,
		returnedSession domain.Session)
	type userServiceMockBehaviour func(us *mockservice.MockUserService, id string, returnedUser domain.User)
	type tokenManagerMockBehaviour func(tm *mockauth.MockTokenManager, pair domain.JWTPair)

	testTable := []struct {
		name                      string
		refreshSessionDTO         domain.RefreshSessionDTO
		oldRefreshSession         domain.Session
		serviceUser               domain.User
		sessionRepoMockBehaviour  sessionRepoMockBehaviour
		userServiceMockBehaviour  userServiceMockBehaviour
		tokenManagerMockBehaviour tokenManagerMockBehaviour
		expectedPair              domain.JWTPair
		expectedErr               error
	}{
		{
			name:              "Success",
			refreshSessionDTO: defaultRefreshSessionDTO,
			oldRefreshSession: defaultRefreshSession,
			serviceUser:       defaultServiceUser,
			sessionRepoMockBehaviour: func(r *mockrepository.MockSessionRepository, oldRefreshToken string, newRefreshToken string,
				returnedSession domain.Session) {
				r.EXPECT().Get(context.Background(), oldRefreshToken).Return(returnedSession, nil)
				r.EXPECT().Set(context.Background(), newRefreshToken, gomock.Any()).Return(nil)
				r.EXPECT().Delete(context.Background(), oldRefreshToken).Return(nil)
			},
			userServiceMockBehaviour: func(us *mockservice.MockUserService, id string, returnedUser domain.User) {
				us.EXPECT().GetByID(context.Background(), id).Return(returnedUser, nil)
			},
			tokenManagerMockBehaviour: func(tm *mockauth.MockTokenManager, pair domain.JWTPair) {
				tm.EXPECT().NewAccessToken(gomock.Any()).Return(pair.AccessToken, nil)
				tm.EXPECT().NewRefreshToken().Return(pair.RefreshToken, nil)
			},
			expectedPair: defaultExpectedPair,
			expectedErr:  nil,
		},
		{
			name:              "Refresh session is not found",
			refreshSessionDTO: defaultRefreshSessionDTO,
			sessionRepoMockBehaviour: func(r *mockrepository.MockSessionRepository, oldRefreshToken string, newRefreshToken string,
				returnedSession domain.Session) {
				r.EXPECT().Get(context.Background(), oldRefreshToken).Return(domain.Session{}, domain.ErrSessionNotFound)
			},
			expectedErr: domain.ErrInvalidRefreshToken,
		},
		{
			name: "Wrong fingerprint passed",
			refreshSessionDTO: domain.RefreshSessionDTO{
				RefreshToken: "qGVFLRQw37TnSmG0LKFN",
				Fingerprint:  "a8347bd7-a308-498e-a6f3-a50517c18057",
			},
			oldRefreshSession: defaultRefreshSession,
			sessionRepoMockBehaviour: func(r *mockrepository.MockSessionRepository, oldRefreshToken string, newRefreshToken string,
				returnedSession domain.Session) {
				r.EXPECT().Get(context.Background(), oldRefreshToken).Return(returnedSession, nil)
				r.EXPECT().Delete(context.Background(), oldRefreshToken).Return(nil)
			},
			expectedErr: domain.ErrInvalidRefreshToken,
		},
		{
			name:              "Session user is not found",
			refreshSessionDTO: defaultRefreshSessionDTO,
			oldRefreshSession: defaultRefreshSession,
			serviceUser:       defaultServiceUser,
			sessionRepoMockBehaviour: func(r *mockrepository.MockSessionRepository, oldRefreshToken string, newRefreshToken string,
				returnedSession domain.Session) {
				r.EXPECT().Get(context.Background(), oldRefreshToken).Return(returnedSession, nil)
				r.EXPECT().Delete(context.Background(), oldRefreshToken).Return(nil)
			},
			userServiceMockBehaviour: func(us *mockservice.MockUserService, id string, returnedUser domain.User) {
				us.EXPECT().GetByID(context.Background(), id).Return(domain.User{}, domain.ErrUserNotFound)
			},
			expectedErr: domain.ErrInvalidRefreshToken,
		},
		{
			name:              "Unexpected get old session",
			refreshSessionDTO: defaultRefreshSessionDTO,
			sessionRepoMockBehaviour: func(r *mockrepository.MockSessionRepository, oldRefreshToken string, newRefreshToken string,
				returnedSession domain.Session) {
				r.EXPECT().Get(context.Background(), oldRefreshToken).Return(domain.Session{}, errUnexpected)
			},
			expectedErr: errUnexpected,
		},
		{
			name:              "Unexpected get session user",
			refreshSessionDTO: defaultRefreshSessionDTO,
			oldRefreshSession: defaultRefreshSession,
			sessionRepoMockBehaviour: func(r *mockrepository.MockSessionRepository, oldRefreshToken string, newRefreshToken string,
				returnedSession domain.Session) {
				r.EXPECT().Get(context.Background(), oldRefreshToken).Return(returnedSession, nil)
				r.EXPECT().Delete(context.Background(), oldRefreshToken).Return(nil)
			},
			userServiceMockBehaviour: func(us *mockservice.MockUserService, id string, returnedUser domain.User) {
				us.EXPECT().GetByID(context.Background(), id).Return(domain.User{}, errUnexpected)
			},
			expectedErr: errUnexpected,
		},
		{
			name:              "Unexpected error while creating access token",
			refreshSessionDTO: defaultRefreshSessionDTO,
			oldRefreshSession: defaultRefreshSession,
			serviceUser:       defaultServiceUser,
			sessionRepoMockBehaviour: func(r *mockrepository.MockSessionRepository, oldRefreshToken string, newRefreshToken string,
				returnedSession domain.Session) {
				r.EXPECT().Get(context.Background(), oldRefreshToken).Return(returnedSession, nil)
				r.EXPECT().Delete(context.Background(), oldRefreshToken).Return(nil)
			},
			userServiceMockBehaviour: func(us *mockservice.MockUserService, id string, returnedUser domain.User) {
				us.EXPECT().GetByID(context.Background(), id).Return(returnedUser, nil)
			},
			tokenManagerMockBehaviour: func(tm *mockauth.MockTokenManager, pair domain.JWTPair) {
				tm.EXPECT().NewAccessToken(gomock.Any()).Return("", errUnexpected)
			},
			expectedErr: errUnexpected,
		},
		{
			name:              "Unexpected error while creating refresh token",
			refreshSessionDTO: defaultRefreshSessionDTO,
			oldRefreshSession: defaultRefreshSession,
			serviceUser:       defaultServiceUser,
			sessionRepoMockBehaviour: func(r *mockrepository.MockSessionRepository, oldRefreshToken string, newRefreshToken string,
				returnedSession domain.Session) {
				r.EXPECT().Get(context.Background(), oldRefreshToken).Return(returnedSession, nil)
				r.EXPECT().Delete(context.Background(), oldRefreshToken).Return(nil)
			},
			userServiceMockBehaviour: func(us *mockservice.MockUserService, id string, returnedUser domain.User) {
				us.EXPECT().GetByID(context.Background(), id).Return(returnedUser, nil)
			},
			tokenManagerMockBehaviour: func(tm *mockauth.MockTokenManager, pair domain.JWTPair) {
				tm.EXPECT().NewAccessToken(gomock.Any()).Return(pair.AccessToken, nil)
				tm.EXPECT().NewRefreshToken().Return("", errUnexpected)
			},
			expectedErr: errUnexpected,
		},
		{
			name:              "Unexpected error while setting new refresh session",
			refreshSessionDTO: defaultRefreshSessionDTO,
			oldRefreshSession: defaultRefreshSession,
			serviceUser:       defaultServiceUser,
			sessionRepoMockBehaviour: func(r *mockrepository.MockSessionRepository, oldRefreshToken string, newRefreshToken string,
				returnedSession domain.Session) {
				r.EXPECT().Get(context.Background(), oldRefreshToken).Return(returnedSession, nil)
				r.EXPECT().Delete(context.Background(), oldRefreshToken).Return(nil)
				r.EXPECT().Set(context.Background(), newRefreshToken, gomock.Any()).Return(errUnexpected)
			},
			userServiceMockBehaviour: func(us *mockservice.MockUserService, id string, returnedUser domain.User) {
				us.EXPECT().GetByID(context.Background(), id).Return(returnedUser, nil)
			},
			tokenManagerMockBehaviour: func(tm *mockauth.MockTokenManager, pair domain.JWTPair) {
				tm.EXPECT().NewAccessToken(gomock.Any()).Return(pair.AccessToken, nil)
				tm.EXPECT().NewRefreshToken().Return(pair.RefreshToken, nil)
			},
			expectedErr: errUnexpected,
		},
	}

	logging.InitLogger(logging.LogConfig{
		LoggerKind: "mock",
	})

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			us := mockservice.NewMockUserService(c)
			sessionRepo := mockrepository.NewMockSessionRepository(c)
			hasher := mockhasher.NewMockPasswordHasher(c)
			tokenManager := mockauth.NewMockTokenManager(c)

			as := NewAuthService(AuthServiceConfig{
				UserService:     us,
				SessionRepo:     sessionRepo,
				Hasher:          hasher,
				TokenManager:    tokenManager,
				AccessTokenTTL:  accessTokenTTL,
				RefreshTokenTTL: refreshTokenTTL,
			})

			if testCase.sessionRepoMockBehaviour != nil {
				testCase.sessionRepoMockBehaviour(
					sessionRepo,
					testCase.refreshSessionDTO.RefreshToken,
					testCase.expectedPair.RefreshToken,
					testCase.oldRefreshSession,
				)
			}

			if testCase.userServiceMockBehaviour != nil {
				testCase.userServiceMockBehaviour(us, testCase.oldRefreshSession.UserID, testCase.serviceUser)
			}

			if testCase.tokenManagerMockBehaviour != nil {
				testCase.tokenManagerMockBehaviour(tokenManager, testCase.expectedPair)
			}

			pair, err := as.Refresh(context.Background(), testCase.refreshSessionDTO)

			if testCase.expectedErr != nil && !errors.Is(testCase.expectedErr, err) {
				t.Errorf("Wrong returned error. Expected error %v, got %v", testCase.expectedErr, err)
			}

			if testCase.expectedErr == nil && err != nil {
				t.Errorf("Unexpected error %v", err)
			}

			if !reflect.DeepEqual(testCase.expectedPair, pair) {
				t.Errorf("Wrong token pair result. Expected %#v, got %#v", testCase.expectedPair, pair)
			}
		})
	}
}

func TestAuthService_Authorize(t *testing.T) {
	type tokenManagerMockBehaviour func(tm *mockauth.MockTokenManager, accessToken string, returnedClaims domain.Claims)

	testTable := []struct {
		name                      string
		accessToken               string
		tokenManagerMockBehaviour tokenManagerMockBehaviour
		expectedClaims            domain.Claims
		expectedErr               error
	}{
		{
			name:        "Success",
			accessToken: "header.payload.sign",
			tokenManagerMockBehaviour: func(tm *mockauth.MockTokenManager, accessToken string, returnedClaims domain.Claims) {
				tm.EXPECT().Parse(accessToken, &domain.Claims{}).SetArg(1, returnedClaims).Return(nil)
			},
			expectedClaims: defaultExpectedClaims,
			expectedErr:    nil,
		},
		{
			name:        "Access token parse error",
			accessToken: "header.payload.sign",
			tokenManagerMockBehaviour: func(tm *mockauth.MockTokenManager, accessToken string, returnedClaims domain.Claims) {
				tm.EXPECT().Parse(accessToken, &domain.Claims{}).Return(errors.New("access token parse error"))
			},
			expectedErr: domain.ErrInvalidAccessToken,
		},
	}

	logging.InitLogger(logging.LogConfig{
		LoggerKind: "mock",
	})

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			us := mockservice.NewMockUserService(c)
			sessionRepo := mockrepository.NewMockSessionRepository(c)
			hasher := mockhasher.NewMockPasswordHasher(c)
			tokenManager := mockauth.NewMockTokenManager(c)

			as := NewAuthService(AuthServiceConfig{
				UserService:     us,
				SessionRepo:     sessionRepo,
				Hasher:          hasher,
				TokenManager:    tokenManager,
				AccessTokenTTL:  accessTokenTTL,
				RefreshTokenTTL: refreshTokenTTL,
			})

			if testCase.tokenManagerMockBehaviour != nil {
				testCase.tokenManagerMockBehaviour(tokenManager, testCase.accessToken, testCase.expectedClaims)
			}

			claims, err := as.Authorize(testCase.accessToken)

			if testCase.expectedErr != nil && !errors.Is(testCase.expectedErr, err) {
				t.Errorf("Wrong returned error. Expected error %v, got %v", testCase.expectedErr, err)
			}

			if testCase.expectedErr == nil && err != nil {
				t.Errorf("Unexpected error %v", err)
			}

			if !reflect.DeepEqual(testCase.expectedClaims, claims) {
				t.Errorf("Wrong claims result. Expected %#v, got %#v", testCase.expectedClaims, claims)
			}
		})
	}
}
