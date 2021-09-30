package service

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/Mort4lis/scht-backend/internal/domain"
	mockrepository "github.com/Mort4lis/scht-backend/internal/repository/mocks"
	mockhasher "github.com/Mort4lis/scht-backend/pkg/hasher/mocks"
	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/golang/mock/gomock"
)

var (
	userCreatedAt = time.Date(2021, time.September, 27, 11, 10, 12, 411, time.Local)
	userUpdatedAt = time.Date(2021, time.November, 14, 22, 0, 53, 512, time.Local)
)

var (
	defaultShortUser = domain.User{
		ID:        "1",
		Username:  "john1967",
		Password:  "8743b52063cd84097a65d1633f5c74f5",
		Email:     "john1967@gmail.com",
		CreatedAt: &userCreatedAt,
	}
	defaultFullUser = domain.User{
		ID:         "1",
		Username:   "john1967",
		Password:   "8743b52063cd84097a65d1633f5c74f5",
		Email:      "john1967@gmail.com",
		FirstName:  "John",
		LastName:   "Lennon",
		BirthDate:  "1949-10-25",
		Department: "IoT",
		IsDeleted:  false,
		CreatedAt:  &userCreatedAt,
		UpdatedAt:  &userUpdatedAt,
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
	defaultUpdateUserDTO = domain.UpdateUserDTO{
		ID:         "1",
		Username:   "john1967",
		Email:      "john1967@gmail.com",
		FirstName:  "John",
		LastName:   "Lennon",
		BirthDate:  "1949-10-25",
		Department: "IoT",
	}
)

func TestUserService_List(t *testing.T) {
	type userRepoMockBehaviour func(ur *mockrepository.MockUserRepository, listUsers []domain.User)

	testTable := []struct {
		name                  string
		userRepoMockBehaviour userRepoMockBehaviour
		expectedUsers         []domain.User
		expectedErr           error
	}{
		{
			name: "Success",
			userRepoMockBehaviour: func(ur *mockrepository.MockUserRepository, listUsers []domain.User) {
				ur.EXPECT().List(context.Background()).Return(listUsers, nil)
			},
			expectedUsers: []domain.User{defaultFullUser, defaultShortUser},
			expectedErr:   nil,
		},
		{
			name: "Unexpected error",
			userRepoMockBehaviour: func(ur *mockrepository.MockUserRepository, listUsers []domain.User) {
				ur.EXPECT().List(context.Background()).Return(nil, errUnexpected)
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

			if testCase.userRepoMockBehaviour != nil {
				testCase.userRepoMockBehaviour(userRepo, testCase.expectedUsers)
			}

			users, err := us.List(context.Background())

			if testCase.expectedErr != nil && !errors.Is(testCase.expectedErr, err) {
				t.Errorf("Wrong returned error. Expected error %v, got %v", testCase.expectedErr, err)
			}

			if testCase.expectedErr == nil && err != nil {
				t.Errorf("Unexpected error %v", err)
			}

			if !reflect.DeepEqual(testCase.expectedUsers, users) {
				t.Errorf("Wrong list user. Expected %#v, got %#v", testCase.expectedUsers, users)
			}
		})
	}
}

func TestUserService_GetByID(t *testing.T) {
	type userRepoMockBehaviour func(ur *mockrepository.MockUserRepository, id string, returnedUser domain.User)

	testTable := []struct {
		name                  string
		id                    string
		userRepoMockBehaviour userRepoMockBehaviour
		expectedUser          domain.User
		expectedErr           error
	}{
		{
			name: "Success",
			id:   "1",
			userRepoMockBehaviour: func(ur *mockrepository.MockUserRepository, id string, returnedUser domain.User) {
				ur.EXPECT().GetByID(context.Background(), id).Return(returnedUser, nil)
			},
			expectedUser: defaultFullUser,
			expectedErr:  nil,
		},
		{
			name: "User is not found",
			id:   "1",
			userRepoMockBehaviour: func(ur *mockrepository.MockUserRepository, id string, returnedUser domain.User) {
				ur.EXPECT().GetByID(context.Background(), id).Return(domain.User{}, domain.ErrUserNotFound)
			},
			expectedErr: domain.ErrUserNotFound,
		},
		{
			name: "Unexpected error",
			id:   "1",
			userRepoMockBehaviour: func(ur *mockrepository.MockUserRepository, id string, returnedUser domain.User) {
				ur.EXPECT().GetByID(context.Background(), id).Return(domain.User{}, errUnexpected)
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

			if testCase.userRepoMockBehaviour != nil {
				testCase.userRepoMockBehaviour(userRepo, testCase.id, testCase.expectedUser)
			}

			user, err := us.GetByID(context.Background(), testCase.id)

			if testCase.expectedErr != nil && !errors.Is(testCase.expectedErr, err) {
				t.Errorf("Wrong returned error. Expected error %v, got %v", testCase.expectedErr, err)
			}

			if testCase.expectedErr == nil && err != nil {
				t.Errorf("Unexpected error %v", err)
			}

			if !reflect.DeepEqual(testCase.expectedUser, user) {
				t.Errorf("Wrong found user. Expected %#v, got %#v", testCase.expectedUser, user)
			}
		})
	}
}

func TestUserService_Username(t *testing.T) {
	type userRepoMockBehaviour func(ur *mockrepository.MockUserRepository, username string, returnedUser domain.User)

	testTable := []struct {
		name                  string
		username              string
		userRepoMockBehaviour userRepoMockBehaviour
		expectedUser          domain.User
		expectedErr           error
	}{
		{
			name:     "Success",
			username: "john1967",
			userRepoMockBehaviour: func(ur *mockrepository.MockUserRepository, username string, returnedUser domain.User) {
				ur.EXPECT().GetByUsername(context.Background(), username).Return(returnedUser, nil)
			},
			expectedUser: defaultFullUser,
			expectedErr:  nil,
		},
		{
			name:     "User is not found",
			username: "john1967",
			userRepoMockBehaviour: func(ur *mockrepository.MockUserRepository, username string, returnedUser domain.User) {
				ur.EXPECT().GetByUsername(context.Background(), username).Return(domain.User{}, domain.ErrUserNotFound)
			},
			expectedErr: domain.ErrUserNotFound,
		},
		{
			name:     "Unexpected error",
			username: "john1967",
			userRepoMockBehaviour: func(ur *mockrepository.MockUserRepository, username string, returnedUser domain.User) {
				ur.EXPECT().GetByUsername(context.Background(), username).Return(domain.User{}, errUnexpected)
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

			if testCase.userRepoMockBehaviour != nil {
				testCase.userRepoMockBehaviour(userRepo, testCase.username, testCase.expectedUser)
			}

			user, err := us.GetByUsername(context.Background(), testCase.username)

			if testCase.expectedErr != nil && !errors.Is(testCase.expectedErr, err) {
				t.Errorf("Wrong returned error. Expected error %v, got %v", testCase.expectedErr, err)
			}

			if testCase.expectedErr == nil && err != nil {
				t.Errorf("Unexpected error %v", err)
			}

			if !reflect.DeepEqual(testCase.expectedUser, user) {
				t.Errorf("Wrong found user. Expected %#v, got %#v", testCase.expectedUser, user)
			}
		})
	}
}

func TestUserService_Create(t *testing.T) {
	type hasherMockBehaviour func(h *mockhasher.MockPasswordHasher, password, hash string)

	type userRepoMockBehaviour func(ur *mockrepository.MockUserRepository, dto domain.CreateUserDTO, createdUser domain.User)

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
			expectedUser: defaultShortUser,
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
				testCase.hasherMockBehaviour(hasher, testCase.createUserDTO.Password, testCase.modifiedCreateUserDTO.Password)
			}

			if testCase.userRepoMockBehaviour != nil {
				testCase.userRepoMockBehaviour(userRepo, testCase.modifiedCreateUserDTO, testCase.expectedUser)
			}

			user, err := us.Create(context.Background(), testCase.createUserDTO)

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

func TestUserService_Update(t *testing.T) {
	type hasherMockBehaviour func(h *mockhasher.MockPasswordHasher, password, hash string)

	type userRepoMockBehaviour func(ur *mockrepository.MockUserRepository, dto domain.UpdateUserDTO, updatedUser domain.User)

	testTable := []struct {
		name                  string
		updateUserDTO         domain.UpdateUserDTO
		modifiedUpdateUserDTO domain.UpdateUserDTO
		hasherMockBehaviour   hasherMockBehaviour
		userRepoMockBehaviour userRepoMockBehaviour
		expectedUser          domain.User
		expectedErr           error
	}{
		{
			name:                  "Success",
			updateUserDTO:         defaultUpdateUserDTO,
			modifiedUpdateUserDTO: defaultUpdateUserDTO,
			userRepoMockBehaviour: func(ur *mockrepository.MockUserRepository, dto domain.UpdateUserDTO, updatedUser domain.User) {
				ur.EXPECT().Update(context.Background(), dto).Return(updatedUser, nil)
			},
			expectedUser: defaultFullUser,
			expectedErr:  nil,
		},
		{
			name:                  "No need to update",
			updateUserDTO:         domain.UpdateUserDTO{ID: "1"},
			modifiedUpdateUserDTO: domain.UpdateUserDTO{ID: "1"},
			userRepoMockBehaviour: func(ur *mockrepository.MockUserRepository, dto domain.UpdateUserDTO, updatedUser domain.User) {
				ur.EXPECT().Update(context.Background(), dto).Return(domain.User{}, domain.ErrUserNoNeedUpdate)
				ur.EXPECT().GetByID(context.Background(), dto.ID).Return(updatedUser, nil)
			},
			expectedUser: defaultFullUser,
			expectedErr:  nil,
		},
		{
			name:                  "Update only password",
			updateUserDTO:         domain.UpdateUserDTO{ID: "1", Password: "qwerty12345"},
			modifiedUpdateUserDTO: domain.UpdateUserDTO{ID: "1", Password: "8743b52063cd84097a65d1633f5c74f5"},
			hasherMockBehaviour: func(h *mockhasher.MockPasswordHasher, password, hash string) {
				h.EXPECT().Hash(password).Return(hash, nil)
			},
			userRepoMockBehaviour: func(ur *mockrepository.MockUserRepository, dto domain.UpdateUserDTO, updatedUser domain.User) {
				ur.EXPECT().Update(context.Background(), dto).Return(updatedUser, nil)
			},
			expectedUser: defaultFullUser,
			expectedErr:  nil,
		},
		{
			name:                  "User is not found to update",
			updateUserDTO:         defaultUpdateUserDTO,
			modifiedUpdateUserDTO: defaultUpdateUserDTO,
			userRepoMockBehaviour: func(ur *mockrepository.MockUserRepository, dto domain.UpdateUserDTO, updatedUser domain.User) {
				ur.EXPECT().Update(context.Background(), dto).Return(domain.User{}, domain.ErrUserNotFound)
			},
			expectedErr: domain.ErrUserNotFound,
		},
		{
			name:                  "Update user with such username or email",
			updateUserDTO:         defaultUpdateUserDTO,
			modifiedUpdateUserDTO: defaultUpdateUserDTO,
			userRepoMockBehaviour: func(ur *mockrepository.MockUserRepository, dto domain.UpdateUserDTO, updatedUser domain.User) {
				ur.EXPECT().Update(context.Background(), dto).Return(domain.User{}, domain.ErrUserUniqueViolation)
			},
			expectedErr: domain.ErrUserUniqueViolation,
		},
		{
			name:          "Unexpected error while hashing password",
			updateUserDTO: domain.UpdateUserDTO{ID: "1", Password: "qwerty12345"},
			hasherMockBehaviour: func(h *mockhasher.MockPasswordHasher, password, hash string) {
				h.EXPECT().Hash(password).Return("", errUnexpected)
			},
			expectedErr: errUnexpected,
		},
		{
			name:                  "Unexpected error while updating user",
			updateUserDTO:         defaultUpdateUserDTO,
			modifiedUpdateUserDTO: defaultUpdateUserDTO,
			userRepoMockBehaviour: func(ur *mockrepository.MockUserRepository, dto domain.UpdateUserDTO, updatedUser domain.User) {
				ur.EXPECT().Update(context.Background(), dto).Return(domain.User{}, errUnexpected)
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
				testCase.hasherMockBehaviour(hasher, testCase.updateUserDTO.Password, testCase.modifiedUpdateUserDTO.Password)
			}

			if testCase.userRepoMockBehaviour != nil {
				testCase.userRepoMockBehaviour(userRepo, testCase.modifiedUpdateUserDTO, testCase.expectedUser)
			}

			user, err := us.Update(context.Background(), testCase.updateUserDTO)

			if testCase.expectedErr != nil && !errors.Is(testCase.expectedErr, err) {
				t.Errorf("Wrong returned error. Expected error %v, got %v", testCase.expectedErr, err)
			}

			if testCase.expectedErr == nil && err != nil {
				t.Errorf("Unexpected error %v", err)
			}

			if !reflect.DeepEqual(testCase.expectedUser, user) {
				t.Errorf("Wrong updated user. Expected %#v, got %#v", testCase.expectedUser, user)
			}
		})
	}
}

func TestUserService_Delete(t *testing.T) {
	type userRepoMockBehaviour func(ur *mockrepository.MockUserRepository, id string)

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

			if testCase.userRepoMockBehaviour != nil {
				testCase.userRepoMockBehaviour(userRepo, testCase.id)
			}

			err := us.Delete(context.Background(), testCase.id)

			if testCase.expectedErr != nil && !errors.Is(testCase.expectedErr, err) {
				t.Errorf("Wrong returned error. Expected error %v, got %v", testCase.expectedErr, err)
			}

			if testCase.expectedErr == nil && err != nil {
				t.Errorf("Unexpected error %v", err)
			}
		})
	}
}
