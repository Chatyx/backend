// +build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/Mort4lis/scht-backend/internal/domain"
	mockrepository "github.com/Mort4lis/scht-backend/internal/repository/mocks"
	mockhasher "github.com/Mort4lis/scht-backend/pkg/hasher/mocks"
	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

var userCreatedAt = time.Date(2021, time.September, 27, 11, 10, 12, 411, time.Local)

var (
	defaultUser = domain.User{
		ID:        "1",
		Username:  "john1967",
		Password:  "8743b52063cd84097a65d1633f5c74f5",
		Email:     "john1967@gmail.com",
		CreatedAt: &userCreatedAt,
	}

	defaultCreateUserDTO = domain.CreateUserDTO{
		Username: "john1967",
		Password: "qwerty12345",
		Email:    "john1967@gmail.com",
	}
	defaultModifiedCreateUserDTO = domain.CreateUserDTO{
		Username: "john1967",
		Password: "8743b52063cd84097a65d1633f5c74f5",
		Email:    "john1967@gmail.com",
	}
)

func TestUserService_Create(t *testing.T) {
	type hasherMockBehaviour func(h *mockhasher.MockPasswordHasher, password, hash string)

	type userRepoMockBehaviour func(ur *mockrepository.MockUserRepository, dto domain.CreateUserDTO, createdUser domain.User)

	logging.InitLogger(
		logging.LogConfig{
			LoggerKind: "mock",
		},
	)

	testTable := []struct {
		name                  string
		createUserDTO         domain.CreateUserDTO
		modifiedCreateUserDTO domain.CreateUserDTO
		hasherMockBehaviour   hasherMockBehaviour
		userRepoMockBehaviour userRepoMockBehaviour
		expectedUser          domain.User
		expectedErr           error
	}{
		{
			name:                  "Success",
			createUserDTO:         defaultCreateUserDTO,
			modifiedCreateUserDTO: defaultModifiedCreateUserDTO,
			hasherMockBehaviour: func(h *mockhasher.MockPasswordHasher, password, hash string) {
				h.EXPECT().Hash(password).Return(hash, nil)
			},
			userRepoMockBehaviour: func(ur *mockrepository.MockUserRepository, dto domain.CreateUserDTO, createdUser domain.User) {
				ur.EXPECT().Create(context.Background(), dto).Return(createdUser, nil)
			},
			expectedUser: defaultUser,
			expectedErr:  nil,
		},
		{
			name:                  "Fail to create user with the same username or email",
			createUserDTO:         defaultCreateUserDTO,
			modifiedCreateUserDTO: defaultModifiedCreateUserDTO,
			hasherMockBehaviour: func(h *mockhasher.MockPasswordHasher, password, hash string) {
				h.EXPECT().Hash(password).Return(hash, nil)
			},
			userRepoMockBehaviour: func(ur *mockrepository.MockUserRepository, dto domain.CreateUserDTO, createdUser domain.User) {
				ur.EXPECT().Create(context.Background(), dto).Return(domain.User{}, domain.ErrUserUniqueViolation)
			},
			expectedErr: domain.ErrUserUniqueViolation,
		},
		{
			name:          "Unexpected error while hashing password",
			createUserDTO: defaultCreateUserDTO,
			hasherMockBehaviour: func(h *mockhasher.MockPasswordHasher, password, hash string) {
				h.EXPECT().Hash(password).Return("", errUnexpected)
			},
			expectedErr: errUnexpected,
		},
		{
			name:                  "Unexpected error while creating user",
			createUserDTO:         defaultCreateUserDTO,
			modifiedCreateUserDTO: defaultModifiedCreateUserDTO,
			hasherMockBehaviour: func(h *mockhasher.MockPasswordHasher, password, hash string) {
				h.EXPECT().Hash(password).Return(hash, nil)
			},
			userRepoMockBehaviour: func(ur *mockrepository.MockUserRepository, dto domain.CreateUserDTO, createdUser domain.User) {
				ur.EXPECT().Create(context.Background(), dto).Return(domain.User{}, errUnexpected)
			},
			expectedErr: errUnexpected,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			hasher := mockhasher.NewMockPasswordHasher(c)
			userRepo := mockrepository.NewMockUserRepository(c)
			sessionRepo := mockrepository.NewMockSessionRepository(c)

			us := NewUserService(userRepo, sessionRepo, hasher)

			if testCase.hasherMockBehaviour != nil {
				testCase.hasherMockBehaviour(hasher, testCase.createUserDTO.Password, testCase.modifiedCreateUserDTO.Password)
			}

			if testCase.userRepoMockBehaviour != nil {
				testCase.userRepoMockBehaviour(userRepo, testCase.modifiedCreateUserDTO, testCase.expectedUser)
			}

			user, err := us.Create(context.Background(), testCase.createUserDTO)

			if testCase.expectedErr != nil {
				assert.EqualError(t, err, testCase.expectedErr.Error())
			}

			if testCase.expectedErr == nil {
				assert.NoError(t, err)
			}

			assert.Equal(t, testCase.expectedUser, user)
		})
	}
}

func TestUserService_Delete(t *testing.T) {
	type userRepoMockBehaviour func(ur *mockrepository.MockUserRepository, id string)

	logging.InitLogger(
		logging.LogConfig{
			LoggerKind: "mock",
		},
	)

	testTable := []struct {
		name                  string
		id                    string
		userRepoMockBehaviour userRepoMockBehaviour
		expectedErr           error
	}{
		{
			name: "Success",
			id:   "1",
			userRepoMockBehaviour: func(ur *mockrepository.MockUserRepository, id string) {
				ur.EXPECT().Delete(context.Background(), id).Return(nil)
			},
			expectedErr: nil,
		},
		{
			name: "User is not found",
			id:   "1",
			userRepoMockBehaviour: func(ur *mockrepository.MockUserRepository, id string) {
				ur.EXPECT().Delete(context.Background(), id).Return(domain.ErrUserNotFound)
			},
			expectedErr: domain.ErrUserNotFound,
		},
		{
			name: "Unexpected error",
			id:   "1",
			userRepoMockBehaviour: func(ur *mockrepository.MockUserRepository, id string) {
				ur.EXPECT().Delete(context.Background(), id).Return(errUnexpected)
			},
			expectedErr: errUnexpected,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			hasher := mockhasher.NewMockPasswordHasher(c)
			userRepo := mockrepository.NewMockUserRepository(c)
			sessionRepo := mockrepository.NewMockSessionRepository(c)

			us := NewUserService(userRepo, sessionRepo, hasher)

			if testCase.userRepoMockBehaviour != nil {
				testCase.userRepoMockBehaviour(userRepo, testCase.id)
			}

			err := us.Delete(context.Background(), testCase.id)

			if testCase.expectedErr != nil {
				assert.EqualError(t, err, testCase.expectedErr.Error())
			}

			if testCase.expectedErr == nil {
				assert.NoError(t, err)
			}
		})
	}
}
