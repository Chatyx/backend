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

var userCreatedAt = time.Date(2021, time.September, 27, 11, 10, 12, 411, time.Local)

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
				t.Errorf("Wrong response body. Expected %s, got %s", testCase.expectedResponseBody, string(respBodyPayload))
			}
		})
	}
}
