package service

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/Mort4lis/scht-backend/internal/domain"
	mockrepository "github.com/Mort4lis/scht-backend/internal/repository/mocks"
	mockhasher "github.com/Mort4lis/scht-backend/pkg/hasher/mocks"
	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/golang/mock/gomock"
)

var (
	defaultBeginCreateUserDTO = domain.CreateUserDTO{
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

	testTable := []struct {
		name                  string
		beginUserDTO          domain.CreateUserDTO
		modifiedUserDTO       domain.CreateUserDTO
		hasherMockBehaviour   hasherMockBehaviour
		userRepoMockBehaviour userRepoMockBehaviour
		expectedUser          domain.User
		expectedErr           error
	}{
		{
			name:            "Success",
			beginUserDTO:    defaultBeginCreateUserDTO,
			modifiedUserDTO: defaultModifiedCreateUserDTO,
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
			name:            "Fail to create user with the same username or email",
			beginUserDTO:    defaultBeginCreateUserDTO,
			modifiedUserDTO: defaultModifiedCreateUserDTO,
			hasherMockBehaviour: func(h *mockhasher.MockPasswordHasher, password, hash string) {
				h.EXPECT().Hash(password).Return(hash, nil)
			},
			userRepoMockBehaviour: func(ur *mockrepository.MockUserRepository, dto domain.CreateUserDTO, createdUser domain.User) {
				ur.EXPECT().Create(context.Background(), dto).Return(domain.User{}, domain.ErrUserUniqueViolation)
			},
			expectedErr: domain.ErrUserUniqueViolation,
		},
		{
			name:         "Unexpected error while hashing password",
			beginUserDTO: defaultBeginCreateUserDTO,
			hasherMockBehaviour: func(h *mockhasher.MockPasswordHasher, password, hash string) {
				h.EXPECT().Hash(password).Return("", errUnexpected)
			},
			expectedErr: errUnexpected,
		},
		{
			name:            "Unexpected error while creating user",
			beginUserDTO:    defaultBeginCreateUserDTO,
			modifiedUserDTO: defaultModifiedCreateUserDTO,
			hasherMockBehaviour: func(h *mockhasher.MockPasswordHasher, password, hash string) {
				h.EXPECT().Hash(password).Return(hash, nil)
			},
			userRepoMockBehaviour: func(ur *mockrepository.MockUserRepository, dto domain.CreateUserDTO, createdUser domain.User) {
				ur.EXPECT().Create(context.Background(), dto).Return(domain.User{}, errUnexpected)
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

			hasher := mockhasher.NewMockPasswordHasher(c)
			userRepo := mockrepository.NewMockUserRepository(c)

			us := NewUserService(userRepo, hasher)

			if testCase.hasherMockBehaviour != nil {
				testCase.hasherMockBehaviour(hasher, testCase.beginUserDTO.Password, testCase.modifiedUserDTO.Password)
			}

			if testCase.userRepoMockBehaviour != nil {
				testCase.userRepoMockBehaviour(userRepo, testCase.modifiedUserDTO, testCase.expectedUser)
			}

			user, err := us.Create(context.Background(), testCase.beginUserDTO)

			if testCase.expectedErr != nil && !errors.Is(testCase.expectedErr, err) {
				t.Errorf("Wrong returned error. Expected error %v, got %v", testCase.expectedErr, err)
			}

			if testCase.expectedErr == nil && err != nil {
				t.Errorf("Unexpected error %v", err)
			}

			if !reflect.DeepEqual(testCase.expectedUser, user) {
				t.Errorf("Wrong created user. Expected %#v, got %#v", testCase.expectedUser, user)
			}
		})
	}
}
