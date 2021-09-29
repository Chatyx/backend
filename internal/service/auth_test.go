package service

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/Mort4lis/scht-backend/pkg/logging"

	"github.com/golang/mock/gomock"

	"github.com/Mort4lis/scht-backend/internal/domain"
	mockrepository "github.com/Mort4lis/scht-backend/internal/repository/mocks"
	mockservice "github.com/Mort4lis/scht-backend/internal/service/mocks"
	mockauth "github.com/Mort4lis/scht-backend/pkg/auth/mocks"
	mockhasher "github.com/Mort4lis/scht-backend/pkg/hasher/mocks"
)

const (
	accessTokenTTL  = 15 * time.Minute    // 15 minutes
	refreshTokenTTL = 15 * 24 * time.Hour // 15 days
)

var (
	errUnexpected = errors.New("unexpected error")

	userCreatedAt = time.Date(2021, time.September, 27, 11, 10, 12, 411, time.Local)

	defaultSignInDTO = domain.SignInDTO{
		Username:    "john1967",
		Password:    "qwerty12345",
		Fingerprint: "5dc49b7a-6153-4eae-9c0f-297655c45f08",
	}
	defaultServiceUser = domain.User{
		ID:        "1",
		Username:  "john1967",
		Password:  "8743b52063cd84097a65d1633f5c74f5",
		Email:     "john1967@gmail.com",
		CreatedAt: &userCreatedAt,
	}
	defaultExpectedPair = domain.JWTPair{
		AccessToken:  "header.payload.sign",
		RefreshToken: "qGVFLRQw37TnSmG0LKFN",
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
