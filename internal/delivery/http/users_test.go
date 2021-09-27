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

	"github.com/google/uuid"

	"github.com/Mort4lis/scht-backend/internal/domain"
	mock_service "github.com/Mort4lis/scht-backend/internal/service/mocks"
	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/Mort4lis/scht-backend/pkg/validator"
	"github.com/golang/mock/gomock"
	"github.com/julienschmidt/httprouter"
)

var (
	userCreatedAt = time.Date(2021, time.September, 27, 11, 10, 12, 411, time.Local)
	userUpdatedAt = time.Date(2021, time.November, 14, 22, 0, 53, 512, time.Local)
)

func TestUserHandler_list(t *testing.T) {
	type mockBehaviour func(us *mock_service.MockUserService, ctx context.Context, users []domain.User)

	testTable := []struct {
		name                 string
		mockOutUsers         []domain.User
		mockBehaviour        mockBehaviour
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name: "Success",
			mockOutUsers: []domain.User{
				{
					ID:        "1",
					Username:  "john1967",
					Email:     "john1967@gmail.com",
					FirstName: "John",
					LastName:  "Lennon",
					BirthDate: "1940-10-09",
					CreatedAt: &userCreatedAt,
					UpdatedAt: &userUpdatedAt,
				},
				{
					ID:        "2",
					Username:  "mick49",
					Email:     "mick49@gmail.com",
					CreatedAt: &userCreatedAt,
				},
			},
			mockBehaviour: func(us *mock_service.MockUserService, ctx context.Context, users []domain.User) {
				us.EXPECT().List(ctx).Return(users, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"list":[{"id":"1","username":"john1967","email":"john1967@gmail.com","first_name":"John","last_name":"Lennon","birth_date":"1940-10-09","created_at":"2021-09-27T11:10:12.000000411+03:00","updated_at":"2021-11-14T22:00:53.000000512+03:00"},{"id":"2","username":"mick49","email":"mick49@gmail.com","created_at":"2021-09-27T11:10:12.000000411+03:00"}]}`,
		},
		{
			name:         "Success empty list",
			mockOutUsers: make([]domain.User, 0),
			mockBehaviour: func(us *mock_service.MockUserService, ctx context.Context, users []domain.User) {
				us.EXPECT().List(ctx).Return(users, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"list":[]}`,
		},
		{
			name:         "Unexpected error",
			mockOutUsers: nil,
			mockBehaviour: func(us *mock_service.MockUserService, ctx context.Context, users []domain.User) {
				us.EXPECT().List(ctx).Return(nil, errors.New("unexpected error"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"message":"internal server error"}`,
		},
	}

	validate, err := validator.New()
	if err != nil {
		t.Errorf("Unexpected error while creating validator: %v", err)
	}

	logging.InitLogger(logging.LogConfig{
		LoggerKind: "mock",
	})

	logger := logging.GetLogger()
	bashHandler := &baseHandler{
		logger:   logger,
		validate: validate,
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			us := mock_service.NewMockUserService(c)
			uh := &userHandler{
				baseHandler: bashHandler,
				userService: us,
				logger:      logger,
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/users", nil)

			if testCase.mockBehaviour != nil {
				testCase.mockBehaviour(us, req.Context(), testCase.mockOutUsers)
			}

			uh.list(rec, req, httprouter.Params{})

			resp := rec.Result()
			if resp.StatusCode != testCase.expectedStatusCode {
				t.Errorf("Wrong response status code. Expected %d, got %d", testCase.expectedStatusCode, resp.StatusCode)
			}

			respBodyPayload, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("Unexpected error while reading response body: %v", err)
				return
			}

			if string(respBodyPayload) != testCase.expectedResponseBody {
				t.Errorf("Wrong response body. Expected %s, got %s", testCase.expectedResponseBody, respBodyPayload)
			}
		})
	}
}

func TestUserHandler_detail(t *testing.T) {
	type mockBehaviour func(us *mock_service.MockUserService, ctx context.Context, id string, returnedUser domain.User)

	testTable := []struct {
		name                 string
		params               httprouter.Params
		mockOutUser          domain.User
		mockBehaviour        mockBehaviour
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:   "Success with required fields",
			params: []httprouter.Param{{Key: "id", Value: "1"}},
			mockOutUser: domain.User{
				ID:        "1",
				Username:  "john1967",
				Email:     "john1967@gmail.com",
				CreatedAt: &userCreatedAt,
			},
			mockBehaviour: func(us *mock_service.MockUserService, ctx context.Context, id string, returnedUser domain.User) {
				us.EXPECT().GetByID(ctx, id).Return(returnedUser, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"id":"1","username":"john1967","email":"john1967@gmail.com","created_at":"2021-09-27T11:10:12.000000411+03:00"}`,
		},
		{
			name:   "Success with full fields",
			params: []httprouter.Param{{Key: "id", Value: "1"}},
			mockOutUser: domain.User{
				ID:         "1",
				Username:   "john1967",
				Email:      "john1967@gmail.com",
				FirstName:  "John",
				LastName:   "Lennon",
				BirthDate:  "1940-10-09",
				Department: "IoT",
				CreatedAt:  &userCreatedAt,
				UpdatedAt:  &userUpdatedAt,
			},
			mockBehaviour: func(us *mock_service.MockUserService, ctx context.Context, id string, returnedUser domain.User) {
				us.EXPECT().GetByID(ctx, id).Return(returnedUser, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"id":"1","username":"john1967","email":"john1967@gmail.com","first_name":"John","last_name":"Lennon","birth_date":"1940-10-09","department":"IoT","created_at":"2021-09-27T11:10:12.000000411+03:00","updated_at":"2021-11-14T22:00:53.000000512+03:00"}`,
		},
		{
			name:   "Not found",
			params: []httprouter.Param{{Key: "id", Value: uuid.New().String()}},
			mockBehaviour: func(us *mock_service.MockUserService, ctx context.Context, id string, returnedUser domain.User) {
				us.EXPECT().GetByID(ctx, id).Return(domain.User{}, domain.ErrUserNotFound)
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `{"message":"user is not found"}`,
		},
		{
			name:   "Unexpected error",
			params: []httprouter.Param{{Key: "id", Value: "1"}},
			mockBehaviour: func(us *mock_service.MockUserService, ctx context.Context, id string, returnedUser domain.User) {
				us.EXPECT().GetByID(ctx, id).Return(domain.User{}, errors.New("unexpected error"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"message":"internal server error"}`,
		},
	}

	validate, err := validator.New()
	if err != nil {
		t.Errorf("Unexpected error while creating validator: %v", err)
	}

	logging.InitLogger(logging.LogConfig{
		LoggerKind: "mock",
	})

	logger := logging.GetLogger()
	bashHandler := &baseHandler{
		logger:   logger,
		validate: validate,
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			us := mock_service.NewMockUserService(c)
			uh := &userHandler{
				baseHandler: bashHandler,
				userService: us,
				logger:      logger,
			}

			id := testCase.params.ByName("id")
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/users/"+id, nil)

			if testCase.mockBehaviour != nil {
				testCase.mockBehaviour(us, req.Context(), id, testCase.mockOutUser)
			}

			uh.detail(rec, req, testCase.params)

			resp := rec.Result()
			if resp.StatusCode != testCase.expectedStatusCode {
				t.Errorf("Wrong response status code. Expected %d, got %d", testCase.expectedStatusCode, resp.StatusCode)
			}

			respBodyPayload, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("Unexpected error while reading response body: %v", err)
				return
			}

			if string(respBodyPayload) != testCase.expectedResponseBody {
				t.Errorf("Wrong response body. Expected %s, got %s", testCase.expectedResponseBody, respBodyPayload)
			}
		})
	}
}

func TestUserHandler_create(t *testing.T) {
	type mockBehavior func(us *mock_service.MockUserService, ctx context.Context, dto domain.CreateUserDTO, user domain.User)

	testTable := []struct {
		name                 string
		requestBody          string
		mockInUserDTO        domain.CreateUserDTO
		mockOutUser          domain.User
		mockBehavior         mockBehavior
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:        "Success with required fields",
			requestBody: `{"username":"test_user","password":"qwerty12345","email":"test_user@gmail.com"}`,
			mockInUserDTO: domain.CreateUserDTO{
				Username: "test_user",
				Password: "qwerty12345",
				Email:    "test_user@gmail.com",
			},
			mockOutUser: domain.User{
				ID:        "1",
				Username:  "test_user",
				Password:  uuid.New().String(),
				Email:     "test_user@gmail.com",
				CreatedAt: &userCreatedAt,
			},
			mockBehavior: func(us *mock_service.MockUserService, ctx context.Context, dto domain.CreateUserDTO, user domain.User) {
				us.EXPECT().Create(ctx, dto).Return(user, nil)
			},
			expectedStatusCode:   http.StatusCreated,
			expectedResponseBody: `{"id":"1","username":"test_user","email":"test_user@gmail.com","created_at":"2021-09-27T11:10:12.000000411+03:00"}`,
		},
		{
			name: "Success with full fields",
			requestBody: `
				{
					"username":"test_user",
					"password":"qwerty12345",
					"email":"test_user@gmail.com", 
					"first_name":"Test first name",
					"last_name":"Test last name",
					"birth_date":"1983-10-27",
					"department":"Test department"
				}`,
			mockInUserDTO: domain.CreateUserDTO{
				Username:   "test_user",
				Password:   "qwerty12345",
				Email:      "test_user@gmail.com",
				FirstName:  "Test first name",
				LastName:   "Test last name",
				BirthDate:  "1983-10-27",
				Department: "Test department",
			},
			mockOutUser: domain.User{
				ID:         "1",
				Username:   "test_user",
				Password:   uuid.New().String(),
				Email:      "test_user@gmail.com",
				FirstName:  "Test first name",
				LastName:   "Test last name",
				BirthDate:  "1983-10-27",
				Department: "Test department",
				CreatedAt:  &userCreatedAt,
			},
			mockBehavior: func(us *mock_service.MockUserService, ctx context.Context, dto domain.CreateUserDTO, user domain.User) {
				us.EXPECT().Create(ctx, dto).Return(user, nil)
			},
			expectedStatusCode:   http.StatusCreated,
			expectedResponseBody: `{"id":"1","username":"test_user","email":"test_user@gmail.com","first_name":"Test first name","last_name":"Test last name","birth_date":"1983-10-27","department":"Test department","created_at":"2021-09-27T11:10:12.000000411+03:00"}`,
		},
		{
			name:                 "Invalid JSON body",
			requestBody:          `{"username":"test_user","password":"qwerty12345","email":"test_user@gmail.com"`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"invalid json body"}`,
		},
		{
			name:                 "Empty body",
			requestBody:          `{}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"validation error","fields":{"email":"field validation for 'email' failed on the 'required' tag","password":"field validation for 'password' failed on the 'required' tag","username":"field validation for 'username' failed on the 'required' tag"}}`,
		},
		{
			name:                 "Short password",
			requestBody:          `{"username":"test_user","password":"test123","email":"test_user@gmail.com"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"validation error","fields":{"password":"field validation for 'password' failed on the 'min' tag"}}`,
		},
		{
			name:                 "Invalid email address",
			requestBody:          `{"username":"test_user","password":"qwerty12345","email":"12345"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"validation error","fields":{"email":"field validation for 'email' failed on the 'email' tag"}}`,
		},
		{
			name:                 "Invalid birth date",
			requestBody:          `{"username":"test_user","password":"qwerty12345","email":"test_user@gmail.com","birth_date":"1999-02-30"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"validation error","fields":{"birth_date":"field validation for 'birth_date' failed on the 'sql-date' tag"}}`,
		},
		{
			name:        "User with such username or email already exists",
			requestBody: `{"username":"test_user","password":"qwerty12345","email":"test_user@gmail.com"}`,
			mockInUserDTO: domain.CreateUserDTO{
				Username: "test_user",
				Password: "qwerty12345",
				Email:    "test_user@gmail.com",
			},
			mockBehavior: func(us *mock_service.MockUserService, ctx context.Context, dto domain.CreateUserDTO, user domain.User) {
				us.EXPECT().Create(ctx, dto).Return(domain.User{}, domain.ErrUserUniqueViolation)
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"user with such username or email already exists"}`,
		},
		{
			name:        "Unexpected error",
			requestBody: `{"username":"test_user","password":"qwerty12345","email":"test_user@gmail.com"}`,
			mockInUserDTO: domain.CreateUserDTO{
				Username: "test_user",
				Password: "qwerty12345",
				Email:    "test_user@gmail.com",
			},
			mockBehavior: func(us *mock_service.MockUserService, ctx context.Context, dto domain.CreateUserDTO, user domain.User) {
				us.EXPECT().Create(ctx, dto).Return(domain.User{}, errors.New("unexpected error"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"message":"internal server error"}`,
		},
	}

	validate, err := validator.New()
	if err != nil {
		t.Errorf("Unexpected error while creating validator: %v", err)
	}

	logging.InitLogger(logging.LogConfig{
		LoggerKind: "mock",
	})

	logger := logging.GetLogger()
	bashHandler := &baseHandler{
		logger:   logger,
		validate: validate,
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			us := mock_service.NewMockUserService(c)
			uh := &userHandler{
				baseHandler: bashHandler,
				userService: us,
				logger:      logger,
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, listUserURL, strings.NewReader(testCase.requestBody))

			if testCase.mockBehavior != nil {
				testCase.mockBehavior(us, req.Context(), testCase.mockInUserDTO, testCase.mockOutUser)
			}

			uh.create(rec, req, httprouter.Params{})

			resp := rec.Result()
			if resp.StatusCode != testCase.expectedStatusCode {
				t.Errorf("Wrong response status code. Expected %d, got %d", testCase.expectedStatusCode, resp.StatusCode)
			}

			respBodyPayload, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("Unexpected error while reading response body: %v", err)
				return
			}

			if string(respBodyPayload) != testCase.expectedResponseBody {
				t.Errorf("Wrong response body. Expected %s, got %s", testCase.expectedResponseBody, respBodyPayload)
			}
		})
	}
}

func TestUserHandler_update(t *testing.T) {
	type mockBehaviour func(us *mock_service.MockUserService, ctx context.Context, dto domain.UpdateUserDTO, updatedUser domain.User)

	testTable := []struct {
		name                 string
		params               httprouter.Params
		requestBody          string
		updateUserDTO        domain.UpdateUserDTO
		updatedUser          domain.User
		mockBehavior         mockBehaviour
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:        "Success",
			params:      []httprouter.Param{{Key: "id", Value: "1"}},
			requestBody: `{"username":"john1967","email":"john1967@gmail.com","first_name":"John","last_name":"Lennon","birth_date":"1967-10-09","department":"HR"}`,
			updateUserDTO: domain.UpdateUserDTO{
				ID:         "1",
				Username:   "john1967",
				Email:      "john1967@gmail.com",
				FirstName:  "John",
				LastName:   "Lennon",
				BirthDate:  "1967-10-09",
				Department: "HR",
			},
			updatedUser: domain.User{
				ID:         "1",
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
			mockBehavior: func(us *mock_service.MockUserService, ctx context.Context, dto domain.UpdateUserDTO, updatedUser domain.User) {
				us.EXPECT().Update(ctx, dto).Return(updatedUser, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"id":"1","username":"john1967","email":"john1967@gmail.com","first_name":"John","last_name":"Lennon","birth_date":"1967-10-09","department":"HR","created_at":"2021-09-27T11:10:12.000000411+03:00","updated_at":"2021-11-14T22:00:53.000000512+03:00"}`,
		},
		{
			name:          "Success with empty body",
			params:        []httprouter.Param{{Key: "id", Value: "1"}},
			requestBody:   `{}`,
			updateUserDTO: domain.UpdateUserDTO{ID: "1"},
			updatedUser: domain.User{
				ID:         "1",
				Username:   "john1967",
				Password:   uuid.New().String(),
				Email:      "john1967@gmail.com",
				FirstName:  "John",
				LastName:   "Lennon",
				BirthDate:  "1967-10-09",
				Department: "HR",
				CreatedAt:  &userCreatedAt,
			},
			mockBehavior: func(us *mock_service.MockUserService, ctx context.Context, dto domain.UpdateUserDTO, updatedUser domain.User) {
				us.EXPECT().Update(ctx, dto).Return(updatedUser, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"id":"1","username":"john1967","email":"john1967@gmail.com","first_name":"John","last_name":"Lennon","birth_date":"1967-10-09","department":"HR","created_at":"2021-09-27T11:10:12.000000411+03:00"}`,
		},
		{
			name:                 "Invalid JSON body",
			params:               []httprouter.Param{{Key: "id", Value: "1"}},
			requestBody:          `{"birth_date""1970-01-01"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"invalid json body"}`,
		},
		{
			name:                 "Short password",
			params:               []httprouter.Param{{Key: "id", Value: "1"}},
			requestBody:          `{"username":"john1967","password":"test123","email":"john1967@gmail.com"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"validation error","fields":{"password":"field validation for 'password' failed on the 'min' tag"}}`,
		},
		{
			name:                 "Invalid email address",
			params:               []httprouter.Param{{Key: "id", Value: "1"}},
			requestBody:          `{"username":"john1967","password":"qwerty12345","email":"12345"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"validation error","fields":{"email":"field validation for 'email' failed on the 'email' tag"}}`,
		},
		{
			name:                 "Invalid birth date",
			params:               []httprouter.Param{{Key: "id", Value: "1"}},
			requestBody:          `{"username":"john1967","password":"qwerty12345","email":"john1967@gmail.com","birth_date":"20.12.1994"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"validation error","fields":{"birth_date":"field validation for 'birth_date' failed on the 'sql-date' tag"}}`,
		},
		{
			name:          "User is not found",
			params:        []httprouter.Param{{Key: "id", Value: "2"}},
			requestBody:   `{"department":"IoT"}`,
			updateUserDTO: domain.UpdateUserDTO{ID: "2", Department: "IoT"},
			mockBehavior: func(us *mock_service.MockUserService, ctx context.Context, dto domain.UpdateUserDTO, updatedUser domain.User) {
				us.EXPECT().Update(ctx, dto).Return(domain.User{}, domain.ErrUserNotFound)
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `{"message":"user is not found"}`,
		},
		{
			name:        "User with such username or email already exists",
			params:      []httprouter.Param{{Key: "id", Value: "2"}},
			requestBody: `{"username":"john1967","password":"qwerty12345","email":"john1967@gmail.com"}`,
			updateUserDTO: domain.UpdateUserDTO{
				ID:       "2",
				Username: "john1967",
				Password: "qwerty12345",
				Email:    "john1967@gmail.com",
			},
			mockBehavior: func(us *mock_service.MockUserService, ctx context.Context, dto domain.UpdateUserDTO, updatedUser domain.User) {
				us.EXPECT().Update(ctx, dto).Return(domain.User{}, domain.ErrUserUniqueViolation)
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"user with such username or email already exists"}`,
		},
		{
			name:        "Unexpected error",
			params:      []httprouter.Param{{Key: "id", Value: "1"}},
			requestBody: `{"username":"john1967","password":"qwerty12345","email":"john1967@gmail.com","birth_date":"1998-01-01"}`,
			updateUserDTO: domain.UpdateUserDTO{
				ID:        "1",
				Username:  "john1967",
				Password:  "qwerty12345",
				Email:     "john1967@gmail.com",
				BirthDate: "1998-01-01",
			},
			mockBehavior: func(us *mock_service.MockUserService, ctx context.Context, dto domain.UpdateUserDTO, updatedUser domain.User) {
				us.EXPECT().Update(ctx, dto).Return(domain.User{}, errors.New("unexpected error"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"message":"internal server error"}`,
		},
	}

	validate, err := validator.New()
	if err != nil {
		t.Errorf("Unexpected error while creating validator: %v", err)
	}

	logging.InitLogger(logging.LogConfig{
		LoggerKind: "mock",
	})

	logger := logging.GetLogger()
	bashHandler := &baseHandler{
		logger:   logger,
		validate: validate,
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			us := mock_service.NewMockUserService(c)
			uh := &userHandler{
				baseHandler: bashHandler,
				userService: us,
				logger:      logger,
			}

			id := testCase.params.ByName("id")
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPatch, "/api/users/"+id, strings.NewReader(testCase.requestBody))

			if testCase.mockBehavior != nil {
				testCase.mockBehavior(us, req.Context(), testCase.updateUserDTO, testCase.updatedUser)
			}

			uh.update(rec, req, testCase.params)

			resp := rec.Result()
			if resp.StatusCode != testCase.expectedStatusCode {
				t.Errorf("Wrong response status code. Expected %d, got %d", testCase.expectedStatusCode, resp.StatusCode)
			}

			respBodyPayload, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("Unexpected error while reading response body: %v", err)
				return
			}

			if string(respBodyPayload) != testCase.expectedResponseBody {
				t.Errorf("Wrong response body. Expected %s, got %s", testCase.expectedResponseBody, respBodyPayload)
			}
		})
	}
}

func TestUserHandler_delete(t *testing.T) {
	type mockBehaviour func(us *mock_service.MockUserService, ctx context.Context, id string)

	testTable := []struct {
		name                 string
		params               httprouter.Params
		mockBehaviour        mockBehaviour
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:   "Success",
			params: []httprouter.Param{{Key: "id", Value: "1"}},
			mockBehaviour: func(us *mock_service.MockUserService, ctx context.Context, id string) {
				us.EXPECT().Delete(ctx, id).Return(nil)
			},
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name:   "Not found",
			params: []httprouter.Param{{Key: "id", Value: uuid.New().String()}},
			mockBehaviour: func(us *mock_service.MockUserService, ctx context.Context, id string) {
				us.EXPECT().Delete(ctx, id).Return(domain.ErrUserNotFound)
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `{"message":"user is not found"}`,
		},
		{
			name:   "Unexpected error",
			params: []httprouter.Param{{Key: "id", Value: "1"}},
			mockBehaviour: func(us *mock_service.MockUserService, ctx context.Context, id string) {
				us.EXPECT().Delete(ctx, id).Return(errors.New("unexpected error"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"message":"internal server error"}`,
		},
	}

	validate, err := validator.New()
	if err != nil {
		t.Errorf("Unexpected error while creating validator: %v", err)
	}

	logging.InitLogger(logging.LogConfig{
		LoggerKind: "mock",
	})

	logger := logging.GetLogger()
	bashHandler := &baseHandler{
		logger:   logger,
		validate: validate,
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			us := mock_service.NewMockUserService(c)
			uh := &userHandler{
				baseHandler: bashHandler,
				userService: us,
				logger:      logger,
			}

			id := testCase.params.ByName("id")
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/api/users/"+id, nil)

			if testCase.mockBehaviour != nil {
				testCase.mockBehaviour(us, req.Context(), id)
			}

			uh.delete(rec, req, testCase.params)

			resp := rec.Result()
			if resp.StatusCode != testCase.expectedStatusCode {
				t.Errorf("Wrong response status code. Expected %d, got %d", testCase.expectedStatusCode, resp.StatusCode)
			}

			respBodyPayload, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Errorf("Unexpected error while reading response body: %v", err)
				return
			}

			if string(respBodyPayload) != testCase.expectedResponseBody {
				t.Errorf("Wrong response body. Expected %s, got %s", testCase.expectedResponseBody, respBodyPayload)
			}
		})
	}
}
