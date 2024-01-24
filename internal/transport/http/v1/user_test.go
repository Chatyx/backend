package v1

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Chatyx/backend/internal/dto"
	"github.com/Chatyx/backend/internal/entity"
	"github.com/Chatyx/backend/pkg/ctxutil"
	"github.com/Chatyx/backend/pkg/validator"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	defaultBirthday  = time.Date(1940, time.October, 9, 0, 0, 0, 0, time.UTC)
	defaultCreatedAt = time.Date(2024, time.January, 23, 0, 0, 0, 0, time.UTC)
)

var (
	errUnexpected = errors.New("unexpected error")
)

func TestUserController_list(t *testing.T) {
	testCases := []struct {
		name                 string
		mockBehavior         func(s *MockUserService)
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name: "Successful",
			mockBehavior: func(s *MockUserService) {
				s.On("List", mock.Anything).Return([]entity.User{
					{
						ID:        1,
						Username:  "john1967",
						Email:     "john1967@gmail.com",
						FirstName: "John",
						LastName:  "Lennon",
						BirthDate: &defaultBirthday,
						Bio:       "...",
						CreatedAt: defaultCreatedAt,
					},
					{
						ID:        2,
						Username:  "mick49",
						Email:     "mick49@gmail.com",
						CreatedAt: defaultCreatedAt,
					},
				}, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"total":2,"data":[{"id":1,"username":"john1967","email":"john1967@gmail.com","first_name":"John","last_name":"Lennon","birth_date":"1940-10-09","bio":"..."},{"id":2,"username":"mick49","email":"mick49@gmail.com"}]}`,
		},
		{
			name: "Successful with empty list",
			mockBehavior: func(s *MockUserService) {
				s.On("List", mock.Anything).Return(nil, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"total":0,"data":[]}`,
		},
		{
			name: "Internal server error",
			mockBehavior: func(s *MockUserService) {
				s.On("List", mock.Anything).Return(nil, errUnexpected)
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"code":"CM0001","message":"internal server error"}`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			service := NewMockUserService(t)
			if testCase.mockBehavior != nil {
				testCase.mockBehavior(service)
			}

			cnt := NewUserController(UserControllerConfig{
				Service:   service,
				Validator: validator.NewValidator(),
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, userListPath, nil)

			cnt.list(rec, req)
			resp := rec.Result()

			assert.Equal(t, testCase.expectedStatusCode, resp.StatusCode)

			respBody, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedResponseBody, string(respBody))
		})
	}
}

func TestUserController_create(t *testing.T) {
	testCases := []struct {
		name                 string
		requestBody          string
		mockBehavior         func(s *MockUserService)
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:        "Successful with required fields",
			requestBody: `{"username":"john1967","password":"qwerty12345","email":"john1967@gmail.com"}`,
			mockBehavior: func(s *MockUserService) {
				s.On("Create", mock.Anything, dto.UserCreate{
					Username: "john1967",
					Password: "qwerty12345",
					Email:    "john1967@gmail.com",
				}).Return(entity.User{
					ID:        1,
					Username:  "john1967",
					Email:     "john1967@gmail.com",
					CreatedAt: defaultCreatedAt,
				}, nil)
			},
			expectedStatusCode:   http.StatusCreated,
			expectedResponseBody: `{"id":1,"username":"john1967","email":"john1967@gmail.com"}`,
		},
		{
			name:        "Successful with all fields",
			requestBody: `{"id":1,"username":"john1967","password":"qwerty12345","email":"john1967@gmail.com","first_name":"John","last_name":"Lennon","birth_date":"1940-10-09","bio":"..."}`,
			mockBehavior: func(s *MockUserService) {
				s.On("Create", mock.Anything, dto.UserCreate{
					Username:  "john1967",
					Password:  "qwerty12345",
					Email:     "john1967@gmail.com",
					FirstName: "John",
					LastName:  "Lennon",
					BirthDate: &defaultBirthday,
					Bio:       "...",
				}).Return(entity.User{
					ID:        1,
					Username:  "john1967",
					Email:     "john1967@gmail.com",
					FirstName: "John",
					LastName:  "Lennon",
					BirthDate: &defaultBirthday,
					Bio:       "...",
					CreatedAt: defaultCreatedAt,
				}, nil)
			},
			expectedStatusCode:   http.StatusCreated,
			expectedResponseBody: `{"id":1,"username":"john1967","email":"john1967@gmail.com","first_name":"John","last_name":"Lennon","birth_date":"1940-10-09","bio":"..."}`,
		},
		{
			name:                 "Decode body error",
			requestBody:          `{"username":"john1967","password":"qwerty12345","email":"john1967@gmail.com"`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0002","message":"decode body error"}`,
		},
		{
			name:                 "Validation error: username, password and email are required",
			requestBody:          `{}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0005","message":"validation error","data":{"email":"failed on the 'required' tag","password":"failed on the 'required' tag","username":"failed on the 'required' tag"}}`,
		},
		{
			name:                 "Validation error: short password",
			requestBody:          `{"username":"john1967","password":"test","email":"john1967@gmail.com"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0005","message":"validation error","data":{"password":"failed on the 'min' tag"}}`,
		},
		{
			name:                 "Validation error: invalid email address",
			requestBody:          `{"username":"john1967","password":"qwerty12345","email":"12345"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0005","message":"validation error","data":{"email":"failed on the 'email' tag"}}`,
		},
		{
			name:                 "Validation error: invalid birth date",
			requestBody:          `{"username":"john1967","password":"qwerty12345","email":"john1967@gmail.com","birth_date":"1999-02-30"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0005","message":"validation error","data":{"birth_date":"failed on the 'datetime' tag"}}`,
		},
		{
			name:        "User with such username or email already exists",
			requestBody: `{"username":"john1967","password":"qwerty12345","email":"john1967@gmail.com"}`,
			mockBehavior: func(s *MockUserService) {
				s.On("Create", mock.Anything, dto.UserCreate{
					Username: "john1967",
					Password: "qwerty12345",
					Email:    "john1967@gmail.com",
				}).Return(entity.User{}, entity.ErrSuchUserAlreadyExists)
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"US0002","message":"user with such username or email already exists"}`,
		},
		{
			name:        "Internal server error",
			requestBody: `{"username":"john1967","password":"qwerty12345","email":"john1967@gmail.com"}`,
			mockBehavior: func(s *MockUserService) {
				s.On("Create", mock.Anything, dto.UserCreate{
					Username: "john1967",
					Password: "qwerty12345",
					Email:    "john1967@gmail.com",
				}).Return(entity.User{}, errUnexpected)
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"code":"CM0001","message":"internal server error"}`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			service := NewMockUserService(t)
			if testCase.mockBehavior != nil {
				testCase.mockBehavior(service)
			}

			cnt := NewUserController(UserControllerConfig{
				Service:   service,
				Validator: validator.NewValidator(),
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, userListPath, strings.NewReader(testCase.requestBody))

			cnt.create(rec, req)
			resp := rec.Result()

			assert.Equal(t, testCase.expectedStatusCode, resp.StatusCode)

			respBody, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedResponseBody, string(respBody))
		})
	}
}

func TestUserController_detail(t *testing.T) {
	testCases := []struct {
		name                 string
		userIDPathParam      string
		mockBehavior         func(s *MockUserService)
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:            "Success with required fields",
			userIDPathParam: "1",
			mockBehavior: func(s *MockUserService) {
				s.On("GetByID", mock.Anything, 1).Return(entity.User{
					ID:        1,
					Username:  "mick49",
					Email:     "mick49@gmail.com",
					CreatedAt: defaultCreatedAt,
				}, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"id":1,"username":"mick49","email":"mick49@gmail.com"}`,
		},
		{
			name:            "Success with all fields",
			userIDPathParam: "1",
			mockBehavior: func(s *MockUserService) {
				s.On("GetByID", mock.Anything, 1).Return(entity.User{
					ID:        1,
					Username:  "john1967",
					Email:     "john1967@gmail.com",
					FirstName: "John",
					LastName:  "Lennon",
					BirthDate: &defaultBirthday,
					Bio:       "...",
					CreatedAt: defaultCreatedAt,
				}, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"id":1,"username":"john1967","email":"john1967@gmail.com","first_name":"John","last_name":"Lennon","birth_date":"1940-10-09","bio":"..."}`,
		},
		{
			name:                 "Decode path param error",
			userIDPathParam:      uuid.New().String(),
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0003","message":"decode path params error"}`,
		},
		{
			name:            "User is not found",
			userIDPathParam: "1",
			mockBehavior: func(s *MockUserService) {
				s.On("GetByID", mock.Anything, 1).Return(entity.User{}, entity.ErrUserNotFound)
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `{"code":"US0001","message":"user is not found"}`,
		},
		{
			name:            "Internal server error",
			userIDPathParam: "1",
			mockBehavior: func(s *MockUserService) {
				s.On("GetByID", mock.Anything, 1).Return(entity.User{}, errUnexpected)
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"code":"CM0001","message":"internal server error"}`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			service := NewMockUserService(t)
			if testCase.mockBehavior != nil {
				testCase.mockBehavior(service)
			}

			cnt := NewUserController(UserControllerConfig{
				Service:   service,
				Validator: validator.NewValidator(),
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, userDetailPath, nil)
			ctx := context.WithValue(
				req.Context(),
				httprouter.ParamsKey,
				httprouter.Params{{Key: "user_id", Value: testCase.userIDPathParam}},
			)
			req = req.WithContext(ctx)

			cnt.detail(rec, req)
			resp := rec.Result()

			assert.Equal(t, testCase.expectedStatusCode, resp.StatusCode)

			respBody, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedResponseBody, string(respBody))
		})
	}
}

func TestUserController_update(t *testing.T) {
	testCases := []struct {
		name                 string
		requestBody          string
		mockBehavior         func(s *MockUserService)
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:        "Successful with required fields",
			requestBody: `{"username":"john1967","email":"john1967@gmail.com"}`,
			mockBehavior: func(s *MockUserService) {
				s.On("Update", mock.Anything, dto.UserUpdate{
					ID:       1,
					Username: "john1967",
					Email:    "john1967@gmail.com",
				}).Return(entity.User{
					ID:        1,
					Username:  "john1967",
					Email:     "john1967@gmail.com",
					CreatedAt: defaultCreatedAt,
				}, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"id":1,"username":"john1967","email":"john1967@gmail.com"}`,
		},
		{
			name:        "Successful with all fields",
			requestBody: `{"username":"john1967","email":"john1967@gmail.com","first_name":"John","last_name":"Lennon","birth_date":"1940-10-09","bio":"..."}`,
			mockBehavior: func(s *MockUserService) {
				s.On("Update", mock.Anything, dto.UserUpdate{
					ID:        1,
					Username:  "john1967",
					Email:     "john1967@gmail.com",
					FirstName: "John",
					LastName:  "Lennon",
					BirthDate: &defaultBirthday,
					Bio:       "...",
				}).Return(entity.User{
					ID:        1,
					Username:  "john1967",
					Email:     "john1967@gmail.com",
					FirstName: "John",
					LastName:  "Lennon",
					BirthDate: &defaultBirthday,
					Bio:       "...",
					CreatedAt: defaultCreatedAt,
				}, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"id":1,"username":"john1967","email":"john1967@gmail.com","first_name":"John","last_name":"Lennon","birth_date":"1940-10-09","bio":"..."}`,
		},
		{
			name:                 "Decode body error",
			requestBody:          `{"username":"john1967","email":"john1967@gmail.com"`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0002","message":"decode body error"}`,
		},
		{
			name:                 "Validation error: username and email are required",
			requestBody:          `{}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0005","message":"validation error","data":{"email":"failed on the 'required' tag","username":"failed on the 'required' tag"}}`,
		},
		{
			name:                 "Validation error: invalid email address",
			requestBody:          `{"username":"john1967","email":"12345"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0005","message":"validation error","data":{"email":"failed on the 'email' tag"}}`,
		},
		{
			name:                 "Validation error: invalid birth date",
			requestBody:          `{"username":"john1967","email":"john1967@gmail.com","birth_date":"1999-02-30"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0005","message":"validation error","data":{"birth_date":"failed on the 'datetime' tag"}}`,
		},
		{
			name:        "User is not found",
			requestBody: `{"username":"john1967","email":"john1967@gmail.com"}`,
			mockBehavior: func(s *MockUserService) {
				s.On("Update", mock.Anything, dto.UserUpdate{
					ID:       1,
					Username: "john1967",
					Email:    "john1967@gmail.com",
				}).Return(entity.User{}, entity.ErrUserNotFound)
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `{"code":"US0001","message":"user is not found"}`,
		},
		{
			name:        "User with such username or email already exists",
			requestBody: `{"username":"john1967","email":"john1967@gmail.com"}`,
			mockBehavior: func(s *MockUserService) {
				s.On("Update", mock.Anything, dto.UserUpdate{
					ID:       1,
					Username: "john1967",
					Email:    "john1967@gmail.com",
				}).Return(entity.User{}, entity.ErrSuchUserAlreadyExists)
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"US0002","message":"user with such username or email already exists"}`,
		},
		{
			name:        "Internal server error",
			requestBody: `{"username":"john1967","password":"qwerty12345","email":"john1967@gmail.com"}`,
			mockBehavior: func(s *MockUserService) {
				s.On("Update", mock.Anything, dto.UserUpdate{
					ID:       1,
					Username: "john1967",
					Email:    "john1967@gmail.com",
				}).Return(entity.User{}, errUnexpected)
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"code":"CM0001","message":"internal server error"}`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			service := NewMockUserService(t)
			if testCase.mockBehavior != nil {
				testCase.mockBehavior(service)
			}

			cnt := NewUserController(UserControllerConfig{
				Service:   service,
				Validator: validator.NewValidator(),
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, userMePath, strings.NewReader(testCase.requestBody))
			ctx := ctxutil.WithUserID(req.Context(), "1")
			req = req.WithContext(ctx)

			cnt.update(rec, req)
			resp := rec.Result()

			assert.Equal(t, testCase.expectedStatusCode, resp.StatusCode)

			respBody, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedResponseBody, string(respBody))
		})
	}
}

func TestUserController_updatePassword(t *testing.T) {
	testCases := []struct {
		name                 string
		requestBody          string
		mockBehavior         func(s *MockUserService)
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:        "Successful",
			requestBody: `{"current_password":"qwerty12345","new_password":"qwerty123456"}`,
			mockBehavior: func(s *MockUserService) {
				s.On("UpdatePassword", mock.Anything, dto.UserUpdatePassword{
					UserID:      1,
					CurPassword: "qwerty12345",
					NewPassword: "qwerty123456",
				}).Return(nil)
			},
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name:                 "Decode body error",
			requestBody:          `{"username":"john1967","email":"john1967@gmail.com"`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0002","message":"decode body error"}`,
		},
		{
			name:                 "Validation error: current and new password are required",
			requestBody:          `{}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0005","message":"validation error","data":{"current_password":"failed on the 'required' tag","new_password":"failed on the 'required' tag"}}`,
		},
		{
			name:        "Wrong current password",
			requestBody: `{"current_password":"qwerty12345","new_password":"qwerty123456"}`,
			mockBehavior: func(s *MockUserService) {
				s.On("UpdatePassword", mock.Anything, dto.UserUpdatePassword{
					UserID:      1,
					CurPassword: "qwerty12345",
					NewPassword: "qwerty123456",
				}).Return(entity.ErrWrongCurrentPassword)
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"US0003","message":"wrong current password"}`,
		},
		{
			name:        "User is not found",
			requestBody: `{"current_password":"qwerty12345","new_password":"qwerty123456"}`,
			mockBehavior: func(s *MockUserService) {
				s.On("UpdatePassword", mock.Anything, dto.UserUpdatePassword{
					UserID:      1,
					CurPassword: "qwerty12345",
					NewPassword: "qwerty123456",
				}).Return(entity.ErrUserNotFound)
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `{"code":"US0001","message":"user is not found"}`,
		},
		{
			name:        "Internal server error",
			requestBody: `{"current_password":"qwerty12345","new_password":"qwerty123456"}`,
			mockBehavior: func(s *MockUserService) {
				s.On("UpdatePassword", mock.Anything, dto.UserUpdatePassword{
					UserID:      1,
					CurPassword: "qwerty12345",
					NewPassword: "qwerty123456",
				}).Return(errUnexpected)
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"code":"CM0001","message":"internal server error"}`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			service := NewMockUserService(t)
			if testCase.mockBehavior != nil {
				testCase.mockBehavior(service)
			}

			cnt := NewUserController(UserControllerConfig{
				Service:   service,
				Validator: validator.NewValidator(),
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, userMePath, strings.NewReader(testCase.requestBody))
			ctx := ctxutil.WithUserID(req.Context(), "1")
			req = req.WithContext(ctx)

			cnt.updatePassword(rec, req)
			resp := rec.Result()

			assert.Equal(t, testCase.expectedStatusCode, resp.StatusCode)

			respBody, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedResponseBody, string(respBody))
		})
	}
}

func TestUserController_delete(t *testing.T) {
	testCases := []struct {
		name                 string
		mockBehavior         func(s *MockUserService)
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name: "Successful",
			mockBehavior: func(s *MockUserService) {
				s.On("Delete", mock.Anything, 1).Return(nil)
			},
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name: "User is not found",
			mockBehavior: func(s *MockUserService) {
				s.On("Delete", mock.Anything, 1).Return(entity.ErrUserNotFound)
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `{"code":"US0001","message":"user is not found"}`,
		},
		{
			name: "Internal server error",
			mockBehavior: func(s *MockUserService) {
				s.On("Delete", mock.Anything, 1).Return(errUnexpected)
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"code":"CM0001","message":"internal server error"}`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			service := NewMockUserService(t)
			if testCase.mockBehavior != nil {
				testCase.mockBehavior(service)
			}

			cnt := NewUserController(UserControllerConfig{
				Service:   service,
				Validator: validator.NewValidator(),
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, userDetailPath, nil)
			ctx := ctxutil.WithUserID(req.Context(), "1")
			req = req.WithContext(ctx)

			cnt.delete(rec, req)
			resp := rec.Result()

			assert.Equal(t, testCase.expectedStatusCode, resp.StatusCode)

			respBody, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedResponseBody, string(respBody))
		})
	}
}
