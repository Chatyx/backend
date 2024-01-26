package v1

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Chatyx/backend/internal/entity"
	"github.com/Chatyx/backend/pkg/validator"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGroupParticipantController_list(t *testing.T) {
	testCases := []struct {
		name                 string
		groupIDPathParam     string
		mockBehavior         func(s *MockGroupParticipantService)
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:             "Successful",
			groupIDPathParam: "1",
			mockBehavior: func(s *MockGroupParticipantService) {
				s.On("List", mock.Anything, 1).Return([]entity.GroupParticipant{
					{
						GroupID: 1,
						UserID:  1,
						IsAdmin: true,
						Status:  entity.JoinedStatus,
					},
					{
						GroupID: 1,
						UserID:  2,
						IsAdmin: false,
						Status:  entity.KickedStatus,
					},
					{
						GroupID: 1,
						UserID:  3,
						IsAdmin: false,
						Status:  entity.LeftStatus,
					},
				}, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"total":3,"data":[{"user_id":1,"status":"joined","is_admin":true},{"user_id":2,"status":"kicked","is_admin":false},{"user_id":3,"status":"left","is_admin":false}]}`,
		},
		{
			name:             "Successful with empty list",
			groupIDPathParam: "1",
			mockBehavior: func(s *MockGroupParticipantService) {
				s.On("List", mock.Anything, 1).Return(nil, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"total":0,"data":[]}`,
		},
		{
			name:                 "Decode path param error",
			groupIDPathParam:     uuid.New().String(),
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0003","message":"decode path params error"}`,
		},
		{
			name:             "Group is not found",
			groupIDPathParam: "1",
			mockBehavior: func(s *MockGroupParticipantService) {
				s.On("List", mock.Anything, 1).Return(nil, entity.ErrGroupNotFound)
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `{"code":"CH0001","message":"group is not found"}`,
		},
		{
			name:             "Internal server error",
			groupIDPathParam: "1",
			mockBehavior: func(s *MockGroupParticipantService) {
				s.On("List", mock.Anything, 1).Return(nil, errUnexpected)
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"code":"CM0001","message":"internal server error"}`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			service := NewMockGroupParticipantService(t)
			if testCase.mockBehavior != nil {
				testCase.mockBehavior(service)
			}

			cnt := NewGroupParticipantController(GroupParticipantControllerConfig{
				Service:   service,
				Validator: validator.NewValidator(),
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, groupDetailPath+groupParticipantListPath, nil)
			ctx := context.WithValue(
				req.Context(),
				httprouter.ParamsKey,
				httprouter.Params{{Key: "group_id", Value: testCase.groupIDPathParam}},
			)
			req = req.WithContext(ctx)

			cnt.list(rec, req)
			resp := rec.Result()

			assert.Equal(t, testCase.expectedStatusCode, resp.StatusCode)

			respBody, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedResponseBody, string(respBody))
		})
	}
}

func TestGroupParticipantController_invite(t *testing.T) {
	testCases := []struct {
		name                 string
		groupIDPathParam     string
		userIDQueryParam     string
		mockBehavior         func(s *MockGroupParticipantService)
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:             "Successful",
			groupIDPathParam: "1",
			userIDQueryParam: "2",
			mockBehavior: func(s *MockGroupParticipantService) {
				s.On("Invite", mock.Anything, 1, 2).Return(entity.GroupParticipant{
					GroupID: 1,
					UserID:  2,
					IsAdmin: false,
					Status:  entity.JoinedStatus,
				}, nil)
			},
			expectedStatusCode:   http.StatusCreated,
			expectedResponseBody: `{"user_id":2,"status":"joined","is_admin":false}`,
		},
		{
			name:                 "Decode path param error",
			groupIDPathParam:     uuid.New().String(),
			userIDQueryParam:     "2",
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0003","message":"decode path params error"}`,
		},
		{
			name:                 "Decode query param error",
			groupIDPathParam:     "1",
			userIDQueryParam:     uuid.New().String(),
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0004","message":"decode query params error"}`,
		},
		{
			name:             "Group is not found",
			groupIDPathParam: "1",
			userIDQueryParam: "2",
			mockBehavior: func(s *MockGroupParticipantService) {
				s.On("Invite", mock.Anything, 1, 2).Return(entity.GroupParticipant{}, entity.ErrGroupNotFound)
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `{"code":"CH0001","message":"group is not found"}`,
		},
		{
			name:             "Invite non-existent user",
			groupIDPathParam: "1",
			userIDQueryParam: "2",
			mockBehavior: func(s *MockGroupParticipantService) {
				s.On("Invite", mock.Anything, 1, 2).Return(entity.GroupParticipant{}, entity.ErrAddNonExistentUserToGroup)
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CH0007","message":"inviting non-existent user to group"}`,
		},
		{
			name:             "Invite user that already exists",
			groupIDPathParam: "1",
			userIDQueryParam: "2",
			mockBehavior: func(s *MockGroupParticipantService) {
				s.On("Invite", mock.Anything, 1, 2).Return(entity.GroupParticipant{}, entity.ErrSuchGroupParticipantAlreadyExists)
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CH0008","message":"such a group participant already exists"}`,
		},
		{
			name:             "Invite user without admin permission",
			groupIDPathParam: "1",
			userIDQueryParam: "2",
			mockBehavior: func(s *MockGroupParticipantService) {
				s.On("Invite", mock.Anything, 1, 2).Return(entity.GroupParticipant{}, entity.ErrForbiddenPerformAction)
			},
			expectedStatusCode:   http.StatusForbidden,
			expectedResponseBody: `{"code":"CM0007","message":"it's forbidden to perform this action"}`,
		},
		{
			name:             "Internal server error",
			groupIDPathParam: "1",
			userIDQueryParam: "2",
			mockBehavior: func(s *MockGroupParticipantService) {
				s.On("Invite", mock.Anything, 1, 2).Return(entity.GroupParticipant{}, errUnexpected)
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"code":"CM0001","message":"internal server error"}`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			service := NewMockGroupParticipantService(t)
			if testCase.mockBehavior != nil {
				testCase.mockBehavior(service)
			}

			cnt := NewGroupParticipantController(GroupParticipantControllerConfig{
				Service:   service,
				Validator: validator.NewValidator(),
			})

			query := url.Values{}
			query.Add("user_id", testCase.userIDQueryParam)
			urlPath := fmt.Sprintf("%s?%s", groupDetailPath+groupParticipantListPath, query.Encode())

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, urlPath, nil)
			ctx := context.WithValue(
				req.Context(),
				httprouter.ParamsKey,
				httprouter.Params{{Key: "group_id", Value: testCase.groupIDPathParam}},
			)
			req = req.WithContext(ctx)

			cnt.invite(rec, req)
			resp := rec.Result()

			assert.Equal(t, testCase.expectedStatusCode, resp.StatusCode)

			respBody, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedResponseBody, string(respBody))
		})
	}
}

func TestGroupParticipantController_detail(t *testing.T) {
	testCases := []struct {
		name                 string
		groupIDPathParam     string
		userIDPathParam      string
		mockBehavior         func(s *MockGroupParticipantService)
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:             "Successful",
			groupIDPathParam: "1",
			userIDPathParam:  "2",
			mockBehavior: func(s *MockGroupParticipantService) {
				s.On("Get", mock.Anything, 1, 2).Return(entity.GroupParticipant{
					GroupID: 1,
					UserID:  2,
					IsAdmin: false,
					Status:  entity.JoinedStatus,
				}, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"user_id":2,"status":"joined","is_admin":false}`,
		},
		{
			name:                 "Decode group_id path param error",
			groupIDPathParam:     uuid.New().String(),
			userIDPathParam:      "2",
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0003","message":"decode path params error"}`,
		},
		{
			name:                 "Decode user_id path param error",
			groupIDPathParam:     "1",
			userIDPathParam:      uuid.New().String(),
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0003","message":"decode path params error"}`,
		},
		{
			name:             "Group is not found",
			groupIDPathParam: "1",
			userIDPathParam:  "2",
			mockBehavior: func(s *MockGroupParticipantService) {
				s.On("Get", mock.Anything, 1, 2).Return(entity.GroupParticipant{}, entity.ErrGroupNotFound)
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `{"code":"CH0001","message":"group is not found"}`,
		},
		{
			name:             "Group participant is not found",
			groupIDPathParam: "1",
			userIDPathParam:  "2",
			mockBehavior: func(s *MockGroupParticipantService) {
				s.On("Get", mock.Anything, 1, 2).Return(entity.GroupParticipant{}, entity.ErrGroupParticipantNotFound)
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `{"code":"CH0006","message":"group participant is not found"}`,
		},
		{
			name:             "Internal server error",
			groupIDPathParam: "1",
			userIDPathParam:  "2",
			mockBehavior: func(s *MockGroupParticipantService) {
				s.On("Get", mock.Anything, 1, 2).Return(entity.GroupParticipant{}, errUnexpected)
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"code":"CM0001","message":"internal server error"}`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			service := NewMockGroupParticipantService(t)
			if testCase.mockBehavior != nil {
				testCase.mockBehavior(service)
			}

			cnt := NewGroupParticipantController(GroupParticipantControllerConfig{
				Service:   service,
				Validator: validator.NewValidator(),
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, groupDetailPath+groupParticipantListPath, nil)
			ctx := context.WithValue(
				req.Context(),
				httprouter.ParamsKey,
				httprouter.Params{
					{Key: "group_id", Value: testCase.groupIDPathParam},
					{Key: "user_id", Value: testCase.userIDPathParam},
				},
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

func TestGroupParticipantController_update(t *testing.T) {
	testCases := []struct {
		name                 string
		groupIDPathParam     string
		userIDPathParam      string
		requestBody          string
		mockBehavior         func(s *MockGroupParticipantService)
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:             "Successful",
			groupIDPathParam: "1",
			userIDPathParam:  "2",
			requestBody:      `{"status":"kicked"}`,
			mockBehavior: func(s *MockGroupParticipantService) {
				s.On("UpdateStatus", mock.Anything, 1, 2, entity.KickedStatus).Return(nil)
			},
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name:               "Successful with empty body",
			groupIDPathParam:   "1",
			userIDPathParam:    "2",
			requestBody:        `{}`,
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name:                 "Decode group_id path param error",
			groupIDPathParam:     uuid.New().String(),
			userIDPathParam:      "2",
			requestBody:          `{"status":"kicked"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0003","message":"decode path params error"}`,
		},
		{
			name:                 "Decode user_id path param error",
			groupIDPathParam:     "1",
			userIDPathParam:      uuid.New().String(),
			requestBody:          `{"status":"kicked"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0003","message":"decode path params error"}`,
		},
		{
			name:                 "Decode body error",
			groupIDPathParam:     "1",
			userIDPathParam:      "2",
			requestBody:          `{"status":"kicked"`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0002","message":"decode body error"}`,
		},
		{
			name:                 "Validation error: unknown status",
			groupIDPathParam:     "1",
			userIDPathParam:      "2",
			requestBody:          `{"status":"test"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0005","message":"validation error","data":{"status":"failed on the 'oneof' tag"}}`,
		},
		{
			name:             "Group is not found",
			groupIDPathParam: "1",
			userIDPathParam:  "2",
			requestBody:      `{"status":"kicked"}`,
			mockBehavior: func(s *MockGroupParticipantService) {
				s.On("UpdateStatus", mock.Anything, 1, 2, entity.KickedStatus).Return(entity.ErrGroupNotFound)
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `{"code":"CH0001","message":"group is not found"}`,
		},
		{
			name:             "Group participant is not found",
			groupIDPathParam: "1",
			userIDPathParam:  "2",
			requestBody:      `{"status":"kicked"}`,
			mockBehavior: func(s *MockGroupParticipantService) {
				s.On("UpdateStatus", mock.Anything, 1, 2, entity.KickedStatus).Return(entity.ErrGroupParticipantNotFound)
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `{"code":"CH0006","message":"group participant is not found"}`,
		},
		{
			name:             "Kick user without admin permission",
			groupIDPathParam: "1",
			userIDPathParam:  "2",
			requestBody:      `{"status":"kicked"}`,
			mockBehavior: func(s *MockGroupParticipantService) {
				s.On("UpdateStatus", mock.Anything, 1, 2, entity.KickedStatus).Return(entity.ErrForbiddenPerformAction)
			},
			expectedStatusCode:   http.StatusForbidden,
			expectedResponseBody: `{"code":"CM0007","message":"it's forbidden to perform this action"}`,
		},
		{
			name:             "Incorrect status transit",
			groupIDPathParam: "1",
			userIDPathParam:  "2",
			requestBody:      `{"status":"kicked"}`,
			mockBehavior: func(s *MockGroupParticipantService) {
				s.On("UpdateStatus", mock.Anything, 1, 2, entity.KickedStatus).Return(entity.ErrIncorrectGroupParticipantStatusTransit)
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CH0009","message":"incorrect group participant status transit"}`,
		},
		{
			name:             "Internal server error",
			groupIDPathParam: "1",
			userIDPathParam:  "2",
			requestBody:      `{"status":"kicked"}`,
			mockBehavior: func(s *MockGroupParticipantService) {
				s.On("UpdateStatus", mock.Anything, 1, 2, entity.KickedStatus).Return(errUnexpected)
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"code":"CM0001","message":"internal server error"}`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			service := NewMockGroupParticipantService(t)
			if testCase.mockBehavior != nil {
				testCase.mockBehavior(service)
			}

			cnt := NewGroupParticipantController(GroupParticipantControllerConfig{
				Service:   service,
				Validator: validator.NewValidator(),
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, groupDetailPath+groupParticipantListPath, strings.NewReader(testCase.requestBody))
			ctx := context.WithValue(
				req.Context(),
				httprouter.ParamsKey,
				httprouter.Params{
					{Key: "group_id", Value: testCase.groupIDPathParam},
					{Key: "user_id", Value: testCase.userIDPathParam},
				},
			)
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
