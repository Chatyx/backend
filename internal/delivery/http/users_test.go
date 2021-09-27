package http

import (
	"context"
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
			name:        "OK with required fields",
			requestBody: `{"username":"test_user","password":"qwerty12345","email":"test_user@gmail.com"}`,
			mockInUserDTO: domain.CreateUserDTO{
				Username: "test_user",
				Password: "qwerty12345",
				Email:    "test_user@gmail.com",
			},
			mockOutUser: domain.User{
				ID:       "b2ccb96d-14b2-43d9-aafb-3dacaca8a200",
				Username: "test_user",
				Password: uuid.New().String(),
				Email:    "test_user@gmail.com",
			},
			mockBehavior: func(us *mock_service.MockUserService, ctx context.Context, dto domain.CreateUserDTO, user domain.User) {
				createdAt := time.Date(2021, time.September, 27, 11, 10, 12, 411, time.UTC).Local()
				user.CreatedAt = &createdAt
				us.EXPECT().Create(ctx, dto).Return(user, nil)
			},
			expectedStatusCode:   http.StatusCreated,
			expectedResponseBody: `{"id":"b2ccb96d-14b2-43d9-aafb-3dacaca8a200","username":"test_user","email":"test_user@gmail.com","created_at":"2021-09-27T14:10:12.000000411+03:00"}`,
		},
	}

	validate, err := validator.New()
	if err != nil {
		t.Errorf("Unexpected error while creating validator: %v", err)
	}

	logging.InitLogger(logging.LogConfig{})
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

			testCase.mockBehavior(us, req.Context(), testCase.mockInUserDTO, testCase.mockOutUser)
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
