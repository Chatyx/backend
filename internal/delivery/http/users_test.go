//go:build unit
// +build unit

package http

import (
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Mort4lis/scht-backend/internal/domain"
	mockservice "github.com/Mort4lis/scht-backend/internal/service/mocks"
	inVld "github.com/Mort4lis/scht-backend/internal/validator"
	"github.com/Mort4lis/scht-backend/pkg/logging"
	pkgVld "github.com/Mort4lis/scht-backend/pkg/validator"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	userCreatedAt = time.Date(2021, time.September, 27, 11, 10, 12, 411, time.Local)
	userUpdatedAt = time.Date(2021, time.November, 14, 22, 0, 53, 512, time.Local)
)

func TestUserHandler_list(t *testing.T) {
	type mockBehaviour func(us *mockservice.MockUserService, returnedUsers []domain.User)

	logging.InitLogger(
		logging.LogConfig{
			LoggerKind: "mock",
		},
	)

	validate, err := inVld.New()
	require.NoError(t, err, "Unexpected error while creating validator")

	pkgVld.SetValidate(validate)

	testTable := []struct {
		name                 string
		returnedUsers        []domain.User
		mockBehaviour        mockBehaviour
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name: "Success",
			returnedUsers: []domain.User{
				{
					ID:        "80323cde-599d-4c25-9f5b-a67357193b1f",
					Username:  "john1967",
					Email:     "john1967@gmail.com",
					FirstName: "John",
					LastName:  "Lennon",
					BirthDate: "1940-10-09",
					CreatedAt: &userCreatedAt,
					UpdatedAt: &userUpdatedAt,
				},
				{
					ID:        "46a5c17b-5354-4e24-8ad0-87e2913ab1bd",
					Username:  "mick49",
					Email:     "mick49@gmail.com",
					CreatedAt: &userCreatedAt,
				},
			},
			mockBehaviour: func(us *mockservice.MockUserService, returnedUsers []domain.User) {
				us.EXPECT().List(gomock.Any()).Return(returnedUsers, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"list":[{"id":"80323cde-599d-4c25-9f5b-a67357193b1f","username":"john1967","email":"john1967@gmail.com","first_name":"John","last_name":"Lennon","birth_date":"1940-10-09","created_at":"2021-09-27T11:10:12.000000411+03:00","updated_at":"2021-11-14T22:00:53.000000512+03:00"},{"id":"46a5c17b-5354-4e24-8ad0-87e2913ab1bd","username":"mick49","email":"mick49@gmail.com","created_at":"2021-09-27T11:10:12.000000411+03:00"}]}`,
		},
		{
			name:          "Success empty list",
			returnedUsers: make([]domain.User, 0),
			mockBehaviour: func(us *mockservice.MockUserService, returnedUsers []domain.User) {
				us.EXPECT().List(gomock.Any()).Return(returnedUsers, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"list":[]}`,
		},
		{
			name:          "Unexpected error",
			returnedUsers: nil,
			mockBehaviour: func(us *mockservice.MockUserService, returnedUsers []domain.User) {
				us.EXPECT().List(gomock.Any()).Return(nil, errors.New("unexpected error"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"message":"internal server error"}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			us := mockservice.NewMockUserService(c)
			uh := newUserHandler(us)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/users", nil)

			if testCase.mockBehaviour != nil {
				testCase.mockBehaviour(us, testCase.returnedUsers)
			}

			uh.list(rec, req)

			resp := rec.Result()

			respBodyPayload, err := ioutil.ReadAll(resp.Body)
			assert.NoError(t, err, "Unexpected error while reading response body")

			assert.Equal(t, testCase.expectedStatusCode, resp.StatusCode)
			assert.Equal(t, testCase.expectedResponseBody, string(respBodyPayload))
		})
	}
}

func TestUserHandler_detail(t *testing.T) {
	type mockBehaviour func(us *mockservice.MockUserService, id string, returnedUser domain.User)

	logging.InitLogger(
		logging.LogConfig{
			LoggerKind: "mock",
		},
	)

	validate, err := inVld.New()
	require.NoError(t, err, "Unexpected error while creating validator")

	pkgVld.SetValidate(validate)

	testTable := []struct {
		name                 string
		userID               string
		returnedUser         domain.User
		mockBehaviour        mockBehaviour
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:   "Success with required fields",
			userID: "80323cde-599d-4c25-9f5b-a67357193b1f",
			returnedUser: domain.User{
				ID:        "80323cde-599d-4c25-9f5b-a67357193b1f",
				Username:  "john1967",
				Email:     "john1967@gmail.com",
				CreatedAt: &userCreatedAt,
			},
			mockBehaviour: func(us *mockservice.MockUserService, id string, returnedUser domain.User) {
				us.EXPECT().GetByID(gomock.Any(), id).Return(returnedUser, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"id":"80323cde-599d-4c25-9f5b-a67357193b1f","username":"john1967","email":"john1967@gmail.com","created_at":"2021-09-27T11:10:12.000000411+03:00"}`,
		},
		{
			name:   "Success with full fields",
			userID: "80323cde-599d-4c25-9f5b-a67357193b1f",
			returnedUser: domain.User{
				ID:         "80323cde-599d-4c25-9f5b-a67357193b1f",
				Username:   "john1967",
				Email:      "john1967@gmail.com",
				FirstName:  "John",
				LastName:   "Lennon",
				BirthDate:  "1940-10-09",
				Department: "IoT",
				CreatedAt:  &userCreatedAt,
				UpdatedAt:  &userUpdatedAt,
			},
			mockBehaviour: func(us *mockservice.MockUserService, id string, returnedUser domain.User) {
				us.EXPECT().GetByID(gomock.Any(), id).Return(returnedUser, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"id":"80323cde-599d-4c25-9f5b-a67357193b1f","username":"john1967","email":"john1967@gmail.com","first_name":"John","last_name":"Lennon","birth_date":"1940-10-09","department":"IoT","created_at":"2021-09-27T11:10:12.000000411+03:00","updated_at":"2021-11-14T22:00:53.000000512+03:00"}`,
		},
		{
			name:   "Not found",
			userID: uuid.New().String(),
			mockBehaviour: func(us *mockservice.MockUserService, id string, returnedUser domain.User) {
				us.EXPECT().GetByID(gomock.Any(), id).Return(domain.User{}, domain.ErrUserNotFound)
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `{"message":"user is not found"}`,
		},
		{
			name:   "Unexpected error",
			userID: "80323cde-599d-4c25-9f5b-a67357193b1f",
			mockBehaviour: func(us *mockservice.MockUserService, id string, returnedUser domain.User) {
				us.EXPECT().GetByID(gomock.Any(), id).Return(domain.User{}, errors.New("unexpected error"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"message":"internal server error"}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			us := mockservice.NewMockUserService(c)
			uh := newUserHandler(us)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/users/"+testCase.userID, nil)
			ctx := context.WithValue(
				req.Context(),
				httprouter.ParamsKey,
				httprouter.Params{{Key: "user_id", Value: testCase.userID}},
			)

			req = req.WithContext(ctx)

			if testCase.mockBehaviour != nil {
				testCase.mockBehaviour(us, testCase.userID, testCase.returnedUser)
			}

			uh.detail(rec, req)

			resp := rec.Result()

			respBodyPayload, err := ioutil.ReadAll(resp.Body)
			assert.NoError(t, err, "Unexpected error while reading response body")

			assert.Equal(t, testCase.expectedStatusCode, resp.StatusCode)
			assert.Equal(t, testCase.expectedResponseBody, string(respBodyPayload))
		})
	}
}

func TestUserHandler_create(t *testing.T) {
	type mockBehavior func(us *mockservice.MockUserService, dto domain.CreateUserDTO, createdUser domain.User)

	logging.InitLogger(
		logging.LogConfig{
			LoggerKind: "mock",
		},
	)

	validate, err := inVld.New()
	require.NoError(t, err, "Unexpected error while creating validator")

	pkgVld.SetValidate(validate)

	testTable := []struct {
		name                 string
		requestBody          string
		createUserDTO        domain.CreateUserDTO
		createdUser          domain.User
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:        "Success with required fields",
			requestBody: `{"username":"john1967","password":"qwerty12345","email":"john1967@gmail.com"}`,
			createUserDTO: domain.CreateUserDTO{
				Username: "john1967",
				Password: "qwerty12345",
				Email:    "john1967@gmail.com",
			},
			createdUser: domain.User{
				ID:        "80323cde-599d-4c25-9f5b-a67357193b1f",
				Username:  "john1967",
				Password:  uuid.New().String(),
				Email:     "john1967@gmail.com",
				CreatedAt: &userCreatedAt,
			},
			mockBehavior: func(us *mockservice.MockUserService, dto domain.CreateUserDTO, createdUser domain.User) {
				us.EXPECT().Create(gomock.Any(), dto).Return(createdUser, nil)
			},
			expectedStatusCode:   http.StatusCreated,
			expectedResponseBody: `{"id":"80323cde-599d-4c25-9f5b-a67357193b1f","username":"john1967","email":"john1967@gmail.com","created_at":"2021-09-27T11:10:12.000000411+03:00"}`,
		},
		{
			name:        "Success with full fields",
			requestBody: `{"username":"john1967","password":"qwerty12345","email":"john1967@gmail.com","first_name":"John","last_name":"Lennon","birth_date":"1983-10-27","department":"IoT"}`,
			createUserDTO: domain.CreateUserDTO{
				Username:   "john1967",
				Password:   "qwerty12345",
				Email:      "john1967@gmail.com",
				FirstName:  "John",
				LastName:   "Lennon",
				BirthDate:  "1983-10-27",
				Department: "IoT",
			},
			createdUser: domain.User{
				ID:         "80323cde-599d-4c25-9f5b-a67357193b1f",
				Username:   "john1967",
				Password:   uuid.New().String(),
				Email:      "john1967@gmail.com",
				FirstName:  "John",
				LastName:   "Lennon",
				BirthDate:  "1983-10-27",
				Department: "IoT",
				CreatedAt:  &userCreatedAt,
			},
			mockBehavior: func(us *mockservice.MockUserService, dto domain.CreateUserDTO, createdUser domain.User) {
				us.EXPECT().Create(gomock.Any(), dto).Return(createdUser, nil)
			},
			expectedStatusCode:   http.StatusCreated,
			expectedResponseBody: `{"id":"80323cde-599d-4c25-9f5b-a67357193b1f","username":"john1967","email":"john1967@gmail.com","first_name":"John","last_name":"Lennon","birth_date":"1983-10-27","department":"IoT","created_at":"2021-09-27T11:10:12.000000411+03:00"}`,
		},
		{
			name:                 "Invalid JSON body",
			requestBody:          `{"username":"john1967","password":"qwerty12345","email":"john1967@gmail.com"`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"invalid body to decode"}`,
		},
		{
			name:                 "Empty body",
			requestBody:          `{}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"validation error","fields":{"email":"field validation for 'email' failed on the 'required' tag","password":"field validation for 'password' failed on the 'required' tag","username":"field validation for 'username' failed on the 'required' tag"}}`,
		},
		{
			name:                 "Short password",
			requestBody:          `{"username":"john1967","password":"test123","email":"john1967@gmail.com"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"validation error","fields":{"password":"field validation for 'password' failed on the 'min' tag"}}`,
		},
		{
			name:                 "Invalid email address",
			requestBody:          `{"username":"john1967","password":"qwerty12345","email":"12345"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"validation error","fields":{"email":"field validation for 'email' failed on the 'email' tag"}}`,
		},
		{
			name:                 "Invalid birth date",
			requestBody:          `{"username":"john1967","password":"qwerty12345","email":"john1967@gmail.com","birth_date":"1999-02-30"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"validation error","fields":{"birth_date":"field validation for 'birth_date' failed on the 'sql-date' tag"}}`,
		},
		{
			name:        "User with such username or email already exists",
			requestBody: `{"username":"john1967","password":"qwerty12345","email":"john1967@gmail.com"}`,
			createUserDTO: domain.CreateUserDTO{
				Username: "john1967",
				Password: "qwerty12345",
				Email:    "john1967@gmail.com",
			},
			mockBehavior: func(us *mockservice.MockUserService, dto domain.CreateUserDTO, createdUser domain.User) {
				us.EXPECT().Create(gomock.Any(), dto).Return(domain.User{}, domain.ErrUserUniqueViolation)
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"user with such username or email already exists"}`,
		},
		{
			name:        "Unexpected error",
			requestBody: `{"username":"john1967","password":"qwerty12345","email":"john1967@gmail.com"}`,
			createUserDTO: domain.CreateUserDTO{
				Username: "john1967",
				Password: "qwerty12345",
				Email:    "john1967@gmail.com",
			},
			mockBehavior: func(us *mockservice.MockUserService, dto domain.CreateUserDTO, createdUser domain.User) {
				us.EXPECT().Create(gomock.Any(), dto).Return(domain.User{}, errors.New("unexpected error"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"message":"internal server error"}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			us := mockservice.NewMockUserService(c)
			uh := newUserHandler(us)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(testCase.requestBody))

			if testCase.mockBehavior != nil {
				testCase.mockBehavior(us, testCase.createUserDTO, testCase.createdUser)
			}

			uh.create(rec, req)

			resp := rec.Result()

			respBodyPayload, err := ioutil.ReadAll(resp.Body)
			assert.NoError(t, err, "Unexpected error while reading response body")

			assert.Equal(t, testCase.expectedStatusCode, resp.StatusCode)
			assert.Equal(t, testCase.expectedResponseBody, string(respBodyPayload))
		})
	}
}

func TestUserHandler_update(t *testing.T) {
	type mockBehaviour func(us *mockservice.MockUserService, dto domain.UpdateUserDTO, updatedUser domain.User)

	logging.InitLogger(
		logging.LogConfig{
			LoggerKind: "mock",
		},
	)

	validate, err := inVld.New()
	require.NoError(t, err, "Unexpected error while creating validator")

	pkgVld.SetValidate(validate)

	testTable := []struct {
		name                 string
		authUser             domain.AuthUser
		requestBody          string
		updateUserDTO        domain.UpdateUserDTO
		updatedUser          domain.User
		mockBehavior         mockBehaviour
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:        "Success",
			authUser:    domain.AuthUser{UserID: "80323cde-599d-4c25-9f5b-a67357193b1f"},
			requestBody: `{"username":"john1967","email":"john1967@gmail.com","first_name":"John","last_name":"Lennon","birth_date":"1967-10-09","department":"HR"}`,
			updateUserDTO: domain.UpdateUserDTO{
				ID:         "80323cde-599d-4c25-9f5b-a67357193b1f",
				Username:   "john1967",
				Email:      "john1967@gmail.com",
				FirstName:  "John",
				LastName:   "Lennon",
				BirthDate:  "1967-10-09",
				Department: "HR",
			},
			updatedUser: domain.User{
				ID:         "80323cde-599d-4c25-9f5b-a67357193b1f",
				Username:   "john1967",
				Password:   uuid.New().String(),
				Email:      "john1967@gmail.com",
				FirstName:  "John",
				LastName:   "Lennon",
				BirthDate:  "1967-10-09",
				Department: "HR",
				CreatedAt:  &userCreatedAt,
				UpdatedAt:  &userUpdatedAt,
			},
			mockBehavior: func(us *mockservice.MockUserService, dto domain.UpdateUserDTO, updatedUser domain.User) {
				us.EXPECT().Update(gomock.Any(), dto).Return(updatedUser, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"id":"80323cde-599d-4c25-9f5b-a67357193b1f","username":"john1967","email":"john1967@gmail.com","first_name":"John","last_name":"Lennon","birth_date":"1967-10-09","department":"HR","created_at":"2021-09-27T11:10:12.000000411+03:00","updated_at":"2021-11-14T22:00:53.000000512+03:00"}`,
		},
		{
			name:                 "Invalid JSON body",
			authUser:             domain.AuthUser{UserID: "80323cde-599d-4c25-9f5b-a67357193b1f"},
			requestBody:          `{"birth_date""1970-01-01"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"invalid body to decode"}`,
		},
		{
			name:                 "Invalid email address",
			authUser:             domain.AuthUser{UserID: "80323cde-599d-4c25-9f5b-a67357193b1f"},
			requestBody:          `{"username":"john1967","email":"12345"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"validation error","fields":{"email":"field validation for 'email' failed on the 'email' tag"}}`,
		},
		{
			name:                 "Invalid birth date",
			authUser:             domain.AuthUser{UserID: "80323cde-599d-4c25-9f5b-a67357193b1f"},
			requestBody:          `{"username":"john1967","email":"john1967@gmail.com","birth_date":"20.12.1994"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"validation error","fields":{"birth_date":"field validation for 'birth_date' failed on the 'sql-date' tag"}}`,
		},
		{
			name:        "User is not found",
			authUser:    domain.AuthUser{UserID: "46a5c17b-5354-4e24-8ad0-87e2913ab1bd"},
			requestBody: `{"username":"john1967","email":"john1967@gmail.com"}`,
			updateUserDTO: domain.UpdateUserDTO{
				ID:       "46a5c17b-5354-4e24-8ad0-87e2913ab1bd",
				Username: "john1967",
				Email:    "john1967@gmail.com",
			},
			mockBehavior: func(us *mockservice.MockUserService, dto domain.UpdateUserDTO, updatedUser domain.User) {
				us.EXPECT().Update(gomock.Any(), dto).Return(domain.User{}, domain.ErrUserNotFound)
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `{"message":"user is not found"}`,
		},
		{
			name:        "User with such username or email already exists",
			authUser:    domain.AuthUser{UserID: "46a5c17b-5354-4e24-8ad0-87e2913ab1bd"},
			requestBody: `{"username":"john1967","email":"john1967@gmail.com"}`,
			updateUserDTO: domain.UpdateUserDTO{
				ID:       "46a5c17b-5354-4e24-8ad0-87e2913ab1bd",
				Username: "john1967",
				Email:    "john1967@gmail.com",
			},
			mockBehavior: func(us *mockservice.MockUserService, dto domain.UpdateUserDTO, updatedUser domain.User) {
				us.EXPECT().Update(gomock.Any(), dto).Return(domain.User{}, domain.ErrUserUniqueViolation)
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"user with such username or email already exists"}`,
		},
		{
			name:        "Unexpected error",
			authUser:    domain.AuthUser{UserID: "80323cde-599d-4c25-9f5b-a67357193b1f"},
			requestBody: `{"username":"john1967","email":"john1967@gmail.com","birth_date":"1998-01-01"}`,
			updateUserDTO: domain.UpdateUserDTO{
				ID:        "80323cde-599d-4c25-9f5b-a67357193b1f",
				Username:  "john1967",
				Email:     "john1967@gmail.com",
				BirthDate: "1998-01-01",
			},
			mockBehavior: func(us *mockservice.MockUserService, dto domain.UpdateUserDTO, updatedUser domain.User) {
				us.EXPECT().Update(gomock.Any(), dto).Return(domain.User{}, errors.New("unexpected error"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"message":"internal server error"}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			us := mockservice.NewMockUserService(c)
			uh := newUserHandler(us)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, "/api/user", strings.NewReader(testCase.requestBody))
			req = req.WithContext(domain.NewContextFromAuthUser(context.Background(), testCase.authUser))

			if testCase.mockBehavior != nil {
				testCase.mockBehavior(us, testCase.updateUserDTO, testCase.updatedUser)
			}

			uh.update(rec, req)

			resp := rec.Result()

			respBodyPayload, err := ioutil.ReadAll(resp.Body)
			assert.NoError(t, err, "Unexpected error while reading response body")

			assert.Equal(t, testCase.expectedStatusCode, resp.StatusCode)
			assert.Equal(t, testCase.expectedResponseBody, string(respBodyPayload))
		})
	}
}

func TestUserHandler_updatePassword(t *testing.T) {
	type mockBehaviour func(us *mockservice.MockUserService, dto domain.UpdateUserPasswordDTO)

	logging.InitLogger(
		logging.LogConfig{
			LoggerKind: "mock",
		},
	)

	validate, err := inVld.New()
	require.NoError(t, err, "Unexpected error while creating validator")

	pkgVld.SetValidate(validate)

	testTable := []struct {
		name                  string
		authUser              domain.AuthUser
		requestBody           string
		updateUserPasswordDTO domain.UpdateUserPasswordDTO
		mockBehavior          mockBehaviour
		expectedStatusCode    int
		expectedResponseBody  string
	}{
		{
			name:        "Success",
			authUser:    domain.AuthUser{UserID: "80323cde-599d-4c25-9f5b-a67357193b1f"},
			requestBody: `{"current_password":"qwerty12345","new_password":"admin55555"}`,
			updateUserPasswordDTO: domain.UpdateUserPasswordDTO{
				UserID:  "80323cde-599d-4c25-9f5b-a67357193b1f",
				New:     "admin55555",
				Current: "qwerty12345",
			},
			mockBehavior: func(us *mockservice.MockUserService, dto domain.UpdateUserPasswordDTO) {
				us.EXPECT().UpdatePassword(gomock.Any(), dto).Return(nil)
			},
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name:                 "Invalid JSON body",
			authUser:             domain.AuthUser{UserID: "80323cde-599d-4c25-9f5b-a67357193b1f"},
			requestBody:          `{"current_password":"qwerty12345","new_password":admin55555"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"invalid body to decode"}`,
		},
		{
			name:                 "Empty body",
			authUser:             domain.AuthUser{UserID: "80323cde-599d-4c25-9f5b-a67357193b1f"},
			requestBody:          `{}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"validation error","fields":{"current_password":"field validation for 'current_password' failed on the 'required' tag","new_password":"field validation for 'new_password' failed on the 'required' tag"}}`,
		},
		{
			name:        "Wrong user's current password",
			authUser:    domain.AuthUser{UserID: "80323cde-599d-4c25-9f5b-a67357193b1f"},
			requestBody: `{"current_password":"qwerty12345","new_password":"admin55555"}`,
			updateUserPasswordDTO: domain.UpdateUserPasswordDTO{
				UserID:  "80323cde-599d-4c25-9f5b-a67357193b1f",
				New:     "admin55555",
				Current: "qwerty12345",
			},
			mockBehavior: func(us *mockservice.MockUserService, dto domain.UpdateUserPasswordDTO) {
				us.EXPECT().UpdatePassword(gomock.Any(), dto).Return(domain.ErrWrongCurrentPassword)
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"wrong current password"}`,
		},
		{
			name:        "User is not found",
			authUser:    domain.AuthUser{UserID: "80323cde-599d-4c25-9f5b-a67357193b1f"},
			requestBody: `{"current_password":"qwerty12345","new_password":"admin55555"}`,
			updateUserPasswordDTO: domain.UpdateUserPasswordDTO{
				UserID:  "80323cde-599d-4c25-9f5b-a67357193b1f",
				New:     "admin55555",
				Current: "qwerty12345",
			},
			mockBehavior: func(us *mockservice.MockUserService, dto domain.UpdateUserPasswordDTO) {
				us.EXPECT().UpdatePassword(gomock.Any(), dto).Return(domain.ErrUserNotFound)
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `{"message":"user is not found"}`,
		},
		{
			name:        "Unexpected error",
			authUser:    domain.AuthUser{UserID: "80323cde-599d-4c25-9f5b-a67357193b1f"},
			requestBody: `{"current_password":"qwerty12345","new_password":"admin55555"}`,
			updateUserPasswordDTO: domain.UpdateUserPasswordDTO{
				UserID:  "80323cde-599d-4c25-9f5b-a67357193b1f",
				New:     "admin55555",
				Current: "qwerty12345",
			},
			mockBehavior: func(us *mockservice.MockUserService, dto domain.UpdateUserPasswordDTO) {
				us.EXPECT().UpdatePassword(gomock.Any(), dto).Return(errors.New("unexpected error"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"message":"internal server error"}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			us := mockservice.NewMockUserService(c)
			uh := newUserHandler(us)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, "/api/user/password", strings.NewReader(testCase.requestBody))
			req = req.WithContext(domain.NewContextFromAuthUser(context.Background(), testCase.authUser))

			if testCase.mockBehavior != nil {
				testCase.mockBehavior(us, testCase.updateUserPasswordDTO)
			}

			uh.updatePassword(rec, req)

			resp := rec.Result()

			respBodyPayload, err := ioutil.ReadAll(resp.Body)
			assert.NoError(t, err, "Unexpected error while reading response body")

			assert.Equal(t, testCase.expectedStatusCode, resp.StatusCode)
			assert.Equal(t, testCase.expectedResponseBody, string(respBodyPayload))
		})
	}
}

func TestUserHandler_delete(t *testing.T) {
	type mockBehaviour func(us *mockservice.MockUserService, id string)

	logging.InitLogger(
		logging.LogConfig{
			LoggerKind: "mock",
		},
	)

	validate, err := inVld.New()
	require.NoError(t, err, "Unexpected error while creating validator")

	pkgVld.SetValidate(validate)

	testTable := []struct {
		name                 string
		authUser             domain.AuthUser
		mockBehaviour        mockBehaviour
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:     "Success",
			authUser: domain.AuthUser{UserID: "80323cde-599d-4c25-9f5b-a67357193b1f"},
			mockBehaviour: func(us *mockservice.MockUserService, id string) {
				us.EXPECT().Delete(gomock.Any(), id).Return(nil)
			},
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name:     "Not found",
			authUser: domain.AuthUser{UserID: "46a5c17b-5354-4e24-8ad0-87e2913ab1bd"},
			mockBehaviour: func(us *mockservice.MockUserService, id string) {
				us.EXPECT().Delete(gomock.Any(), id).Return(domain.ErrUserNotFound)
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `{"message":"user is not found"}`,
		},
		{
			name:     "Unexpected error",
			authUser: domain.AuthUser{UserID: "80323cde-599d-4c25-9f5b-a67357193b1f"},
			mockBehaviour: func(us *mockservice.MockUserService, id string) {
				us.EXPECT().Delete(gomock.Any(), id).Return(errors.New("unexpected error"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"message":"internal server error"}`,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			us := mockservice.NewMockUserService(c)
			uh := newUserHandler(us)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/api/user", nil)
			req = req.WithContext(domain.NewContextFromAuthUser(context.Background(), testCase.authUser))

			if testCase.mockBehaviour != nil {
				testCase.mockBehaviour(us, testCase.authUser.UserID)
			}

			uh.delete(rec, req)

			resp := rec.Result()

			respBodyPayload, err := ioutil.ReadAll(resp.Body)
			assert.NoError(t, err, "Unexpected error while reading response body")

			assert.Equal(t, testCase.expectedStatusCode, resp.StatusCode)
			assert.Equal(t, testCase.expectedResponseBody, string(respBodyPayload))
		})
	}
}
