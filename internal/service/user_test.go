//go:build unit
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
	defaultUpdateUserPasswordDTO = domain.UpdateUserPasswordDTO{
		UserID:  "1",
		New:     "admin5555",
		Current: "qwerty12345",
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
				assert.ErrorIs(t, err, testCase.expectedErr)
			}

			if testCase.expectedErr == nil {
				assert.NoError(t, err)
			}

			assert.Equal(t, testCase.expectedUser, user)
		})
	}
}

func TestUserService_UpdatePassword(t *testing.T) {
	type hasherMockBehaviour func(h *mockhasher.MockPasswordHasher, currentPassword, newPassword, currentHash, newHash string)

	type userRepoMockBehaviour func(ur *mockrepository.MockUserRepository, userID, newHash string, returnedUser domain.User)

	type sessionRepoMockBehaviour func(sr *mockrepository.MockSessionRepository, userID string)

	logging.InitLogger(
		logging.LogConfig{
			LoggerKind: "mock",
		},
	)

	testTable := []struct {
		name                     string
		updateUserPasswordDTO    domain.UpdateUserPasswordDTO
		returnedUser             domain.User
		newPasswordHash          string
		userRepoMockBehaviour    userRepoMockBehaviour
		sessionRepoMockBehaviour sessionRepoMockBehaviour
		hasherMockBehaviour      hasherMockBehaviour
		expectedErr              error
	}{
		{
			name:                  "Success",
			updateUserPasswordDTO: defaultUpdateUserPasswordDTO,
			returnedUser:          defaultUser,
			newPasswordHash:       "84097a65d1633f5c74f58743b",
			userRepoMockBehaviour: func(ur *mockrepository.MockUserRepository, userID, newHash string, returnedUser domain.User) {
				ur.EXPECT().GetByID(context.Background(), userID).Return(returnedUser, nil)
				ur.EXPECT().UpdatePassword(context.Background(), userID, newHash).Return(nil)
			},
			sessionRepoMockBehaviour: func(sr *mockrepository.MockSessionRepository, userID string) {
				sr.EXPECT().DeleteAllByUserID(context.Background(), userID).Return(nil)
			},
			hasherMockBehaviour: func(h *mockhasher.MockPasswordHasher, currentPassword, newPassword, currentHash, newHash string) {
				h.EXPECT().CompareHashAndPassword(currentHash, currentPassword).Return(true)
				h.EXPECT().Hash(newPassword).Return(newHash, nil)
			},
			expectedErr: nil,
		},
		{
			name:                  "User is not found",
			updateUserPasswordDTO: defaultUpdateUserPasswordDTO,
			userRepoMockBehaviour: func(ur *mockrepository.MockUserRepository, userID, newHash string, returnedUser domain.User) {
				ur.EXPECT().GetByID(context.Background(), userID).Return(domain.User{}, domain.ErrUserNotFound)
			},
			expectedErr: domain.ErrUserNotFound,
		},
		{
			name:                  "Current password is wrong",
			updateUserPasswordDTO: defaultUpdateUserPasswordDTO,
			returnedUser:          defaultUser,
			userRepoMockBehaviour: func(ur *mockrepository.MockUserRepository, userID, newHash string, returnedUser domain.User) {
				ur.EXPECT().GetByID(context.Background(), userID).Return(returnedUser, nil)
			},
			hasherMockBehaviour: func(h *mockhasher.MockPasswordHasher, currentPassword, newPassword, currentHash, newHash string) {
				h.EXPECT().CompareHashAndPassword(currentHash, currentPassword).Return(false)
			},
			expectedErr: domain.ErrWrongCurrentPassword,
		},
		{
			name:                  "Unexpected error while hashing new password",
			updateUserPasswordDTO: defaultUpdateUserPasswordDTO,
			returnedUser:          defaultUser,
			userRepoMockBehaviour: func(ur *mockrepository.MockUserRepository, userID, newHash string, returnedUser domain.User) {
				ur.EXPECT().GetByID(context.Background(), userID).Return(returnedUser, nil)
			},
			hasherMockBehaviour: func(h *mockhasher.MockPasswordHasher, currentPassword, newPassword, currentHash, newHash string) {
				h.EXPECT().CompareHashAndPassword(currentHash, currentPassword).Return(true)
				h.EXPECT().Hash(newPassword).Return("", errUnexpected)
			},
			expectedErr: errUnexpected,
		},
		{
			name:                  "Unexpected error while reset all user sessions",
			updateUserPasswordDTO: defaultUpdateUserPasswordDTO,
			returnedUser:          defaultUser,
			newPasswordHash:       "84097a65d1633f5c74f58743b",
			userRepoMockBehaviour: func(ur *mockrepository.MockUserRepository, userID, newHash string, returnedUser domain.User) {
				ur.EXPECT().GetByID(context.Background(), userID).Return(returnedUser, nil)
			},
			sessionRepoMockBehaviour: func(sr *mockrepository.MockSessionRepository, userID string) {
				sr.EXPECT().DeleteAllByUserID(context.Background(), userID).Return(errUnexpected)
			},
			hasherMockBehaviour: func(h *mockhasher.MockPasswordHasher, currentPassword, newPassword, currentHash, newHash string) {
				h.EXPECT().CompareHashAndPassword(currentHash, currentPassword).Return(true)
				h.EXPECT().Hash(newPassword).Return(newHash, nil)
			},
			expectedErr: errUnexpected,
		},
		{
			name:                  "Unexpected error while update password",
			updateUserPasswordDTO: defaultUpdateUserPasswordDTO,
			returnedUser:          defaultUser,
			newPasswordHash:       "84097a65d1633f5c74f58743b",
			userRepoMockBehaviour: func(ur *mockrepository.MockUserRepository, userID, newHash string, returnedUser domain.User) {
				ur.EXPECT().GetByID(context.Background(), userID).Return(returnedUser, nil)
				ur.EXPECT().UpdatePassword(context.Background(), userID, newHash).Return(errUnexpected)
			},
			sessionRepoMockBehaviour: func(sr *mockrepository.MockSessionRepository, userID string) {
				sr.EXPECT().DeleteAllByUserID(context.Background(), userID).Return(nil)
			},
			hasherMockBehaviour: func(h *mockhasher.MockPasswordHasher, currentPassword, newPassword, currentHash, newHash string) {
				h.EXPECT().CompareHashAndPassword(currentHash, currentPassword).Return(true)
				h.EXPECT().Hash(newPassword).Return(newHash, nil)
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
				testCase.userRepoMockBehaviour(
					userRepo,
					testCase.updateUserPasswordDTO.UserID, testCase.newPasswordHash, testCase.returnedUser,
				)
			}

			if testCase.hasherMockBehaviour != nil {
				testCase.hasherMockBehaviour(
					hasher,
					testCase.updateUserPasswordDTO.Current, testCase.updateUserPasswordDTO.New,
					testCase.returnedUser.Password, testCase.newPasswordHash,
				)
			}

			if testCase.sessionRepoMockBehaviour != nil {
				testCase.sessionRepoMockBehaviour(sessionRepo, testCase.updateUserPasswordDTO.UserID)
			}

			err := us.UpdatePassword(context.Background(), testCase.updateUserPasswordDTO)

			if testCase.expectedErr != nil {
				assert.ErrorIs(t, err, testCase.expectedErr)
			}

			if testCase.expectedErr == nil {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUserService_Delete(t *testing.T) {
	type userRepoMockBehaviour func(ur *mockrepository.MockUserRepository, id string)

	type sessionRepoMockBehaviour func(sr *mockrepository.MockSessionRepository, id string)

	logging.InitLogger(
		logging.LogConfig{
			LoggerKind: "mock",
		},
	)

	testTable := []struct {
		name                     string
		id                       string
		userRepoMockBehaviour    userRepoMockBehaviour
		sessionRepoMockBehaviour sessionRepoMockBehaviour
		expectedErr              error
	}{
		{
			name: "Success",
			id:   "1",
			userRepoMockBehaviour: func(ur *mockrepository.MockUserRepository, id string) {
				ur.EXPECT().Delete(context.Background(), id).Return(nil)
			},
			sessionRepoMockBehaviour: func(sr *mockrepository.MockSessionRepository, id string) {
				sr.EXPECT().DeleteAllByUserID(context.Background(), id).Return(nil)
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
			name: "Unexpected error while deleting user's session",
			id:   "1",
			userRepoMockBehaviour: func(ur *mockrepository.MockUserRepository, id string) {
				ur.EXPECT().Delete(context.Background(), id).Return(nil)
			},
			sessionRepoMockBehaviour: func(sr *mockrepository.MockSessionRepository, id string) {
				sr.EXPECT().DeleteAllByUserID(context.Background(), id).Return(errUnexpected)
			},
			expectedErr: errUnexpected,
		},
		{
			name: "Unexpected error while deleting user from repo",
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

			if testCase.sessionRepoMockBehaviour != nil {
				testCase.sessionRepoMockBehaviour(sessionRepo, testCase.id)
			}

			err := us.Delete(context.Background(), testCase.id)

			if testCase.expectedErr != nil {
				assert.ErrorIs(t, err, testCase.expectedErr)
			}

			if testCase.expectedErr == nil {
				assert.NoError(t, err)
			}
		})
	}
}
