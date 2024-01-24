package v1

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Chatyx/backend/internal/dto"
	"github.com/Chatyx/backend/internal/entity"
	"github.com/Chatyx/backend/pkg/validator"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDialogController_list(t *testing.T) {
	testCases := []struct {
		name                 string
		mockBehavior         func(s *MockDialogService)
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name: "Successful",
			mockBehavior: func(s *MockDialogService) {
				s.On("List", mock.Anything).Return([]entity.Dialog{
					{
						ID:        1,
						IsBlocked: false,
						Partner: entity.DialogPartner{
							UserID:    2,
							IsBlocked: true,
						},
						CreatedAt: defaultCreatedAt,
					},
					{
						ID:        2,
						IsBlocked: true,
						Partner: entity.DialogPartner{
							UserID:    3,
							IsBlocked: false,
						},
						CreatedAt: defaultCreatedAt,
					},
				}, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"total":2,"data":[{"id":1,"is_blocked":false,"partner":{"user_id":2,"is_blocked":true},"created_at":"2024-01-23T00:00:00Z"},{"id":2,"is_blocked":true,"partner":{"user_id":3,"is_blocked":false},"created_at":"2024-01-23T00:00:00Z"}]}`,
		},
		{
			name: "Successful with empty list",
			mockBehavior: func(s *MockDialogService) {
				s.On("List", mock.Anything).Return(nil, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"total":0,"data":[]}`,
		},
		{
			name: "Internal server error",
			mockBehavior: func(s *MockDialogService) {
				s.On("List", mock.Anything).Return(nil, errUnexpected)
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"code":"CM0001","message":"internal server error"}`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			service := NewMockDialogService(t)
			if testCase.mockBehavior != nil {
				testCase.mockBehavior(service)
			}

			cnt := NewDialogController(DialogControllerConfig{
				Service:   service,
				Validator: validator.NewValidator(),
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, dialogListPath, nil)

			cnt.list(rec, req)
			resp := rec.Result()

			assert.Equal(t, testCase.expectedStatusCode, resp.StatusCode)

			respBody, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedResponseBody, string(respBody))
		})
	}
}

func TestDialogController_create(t *testing.T) {
	testCases := []struct {
		name                 string
		requestBody          string
		mockBehavior         func(s *MockDialogService)
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:        "Successful",
			requestBody: `{"partner":{"user_id":2}}`,
			mockBehavior: func(s *MockDialogService) {
				s.On("Create", mock.Anything, dto.DialogCreate{
					PartnerUserID: 2,
				}).Return(entity.Dialog{
					ID:        1,
					IsBlocked: false,
					Partner: entity.DialogPartner{
						UserID:    2,
						IsBlocked: false,
					},
					CreatedAt: defaultCreatedAt,
				}, nil)
			},
			expectedStatusCode:   http.StatusCreated,
			expectedResponseBody: `{"id":1,"is_blocked":false,"partner":{"user_id":2,"is_blocked":false},"created_at":"2024-01-23T00:00:00Z"}`,
		},
		{
			name:                 "Decode body error",
			requestBody:          `{"partner":{"user_id":2}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0002","message":"decode body error"}`,
		},
		{
			name:                 "Validation error: partner.user_id is required",
			requestBody:          `{}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0005","message":"validation error","data":{"partner.user_id":"failed on the 'required' tag"}}`,
		},
		{
			name:        "Create a dialog that already exists",
			requestBody: `{"partner":{"user_id":2}}`,
			mockBehavior: func(s *MockDialogService) {
				s.On("Create", mock.Anything, dto.DialogCreate{
					PartnerUserID: 2,
				}).Return(entity.Dialog{}, entity.ErrSuchDialogAlreadyExists)
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CH0003","message":"such a dialog already exists"}`,
		},
		{
			name:        "Create a dialog with yourself",
			requestBody: `{"partner":{"user_id":2}}`,
			mockBehavior: func(s *MockDialogService) {
				s.On("Create", mock.Anything, dto.DialogCreate{
					PartnerUserID: 2,
				}).Return(entity.Dialog{}, entity.ErrCreatingDialogWithYourself)
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CH0004","message":"creating a dialog with yourself"}`,
		},
		{
			name:        "Create a dialog with non-existent user",
			requestBody: `{"partner":{"user_id":2}}`,
			mockBehavior: func(s *MockDialogService) {
				s.On("Create", mock.Anything, dto.DialogCreate{
					PartnerUserID: 2,
				}).Return(entity.Dialog{}, entity.ErrCreatingDialogWithNonExistenceUser)
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CH0005","message":"creating a dialog with a non-existent user"}`,
		},
		{
			name:        "Internal server error",
			requestBody: `{"partner":{"user_id":2}}`,
			mockBehavior: func(s *MockDialogService) {
				s.On("Create", mock.Anything, dto.DialogCreate{
					PartnerUserID: 2,
				}).Return(entity.Dialog{}, errUnexpected)
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"code":"CM0001","message":"internal server error"}`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			service := NewMockDialogService(t)
			if testCase.mockBehavior != nil {
				testCase.mockBehavior(service)
			}

			cnt := NewDialogController(DialogControllerConfig{
				Service:   service,
				Validator: validator.NewValidator(),
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, dialogListPath, strings.NewReader(testCase.requestBody))

			cnt.create(rec, req)
			resp := rec.Result()

			assert.Equal(t, testCase.expectedStatusCode, resp.StatusCode)

			respBody, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedResponseBody, string(respBody))
		})
	}
}

func TestDialogController_detail(t *testing.T) {
	testCases := []struct {
		name                 string
		dialogIDPathParam    string
		mockBehavior         func(s *MockDialogService)
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:              "Successful",
			dialogIDPathParam: "1",
			mockBehavior: func(s *MockDialogService) {
				s.On("GetByID", mock.Anything, 1).Return(entity.Dialog{
					ID:        1,
					IsBlocked: false,
					Partner: entity.DialogPartner{
						UserID:    2,
						IsBlocked: true,
					},
					CreatedAt: defaultCreatedAt,
				}, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"id":1,"is_blocked":false,"partner":{"user_id":2,"is_blocked":true},"created_at":"2024-01-23T00:00:00Z"}`,
		},
		{
			name:                 "Decode path param error",
			dialogIDPathParam:    uuid.New().String(),
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0003","message":"decode path params error"}`,
		},
		{
			name:              "Dialog is not found",
			dialogIDPathParam: "1",
			mockBehavior: func(s *MockDialogService) {
				s.On("GetByID", mock.Anything, 1).Return(entity.Dialog{}, entity.ErrDialogNotFound)
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `{"code":"CH0002","message":"dialog is not found"}`,
		},
		{
			name:              "Internal server error",
			dialogIDPathParam: "1",
			mockBehavior: func(s *MockDialogService) {
				s.On("GetByID", mock.Anything, 1).Return(entity.Dialog{}, errUnexpected)
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"code":"CM0001","message":"internal server error"}`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			service := NewMockDialogService(t)
			if testCase.mockBehavior != nil {
				testCase.mockBehavior(service)
			}

			cnt := NewDialogController(DialogControllerConfig{
				Service:   service,
				Validator: validator.NewValidator(),
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, dialogDetailPath, nil)
			ctx := context.WithValue(
				req.Context(),
				httprouter.ParamsKey,
				httprouter.Params{{Key: "dialog_id", Value: testCase.dialogIDPathParam}},
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

func TestDialogController_update(t *testing.T) {
	testCases := []struct {
		name                 string
		dialogIDPathParam    string
		requestBody          string
		mockBehavior         func(s *MockDialogService)
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:              "Successful",
			dialogIDPathParam: "1",
			requestBody:       `{"partner":{"is_blocked":true}}`,
			mockBehavior: func(s *MockDialogService) {
				s.On("Update", mock.Anything, dto.DialogUpdate{
					ID:               1,
					PartnerIsBlocked: boolPtr(true),
				}).Return(nil)
			},
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name:                 "Decode path param error",
			dialogIDPathParam:    uuid.New().String(),
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0003","message":"decode path params error"}`,
		},
		{
			name:                 "Decode body error",
			dialogIDPathParam:    "1",
			requestBody:          `{"name":"Test1","description":"Test1 dialog description"`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0002","message":"decode body error"}`,
		},
		{
			name:              "Dialog is not found",
			dialogIDPathParam: "1",
			requestBody:       `{"partner":{"is_blocked":true}}`,
			mockBehavior: func(s *MockDialogService) {
				s.On("Update", mock.Anything, dto.DialogUpdate{
					ID:               1,
					PartnerIsBlocked: boolPtr(true),
				}).Return(entity.ErrDialogNotFound)
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `{"code":"CH0002","message":"dialog is not found"}`,
		},
		{
			name:              "Internal server error",
			dialogIDPathParam: "1",
			requestBody:       `{"partner":{"is_blocked":true}}`,
			mockBehavior: func(s *MockDialogService) {
				s.On("Update", mock.Anything, dto.DialogUpdate{
					ID:               1,
					PartnerIsBlocked: boolPtr(true),
				}).Return(errUnexpected)
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"code":"CM0001","message":"internal server error"}`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			service := NewMockDialogService(t)
			if testCase.mockBehavior != nil {
				testCase.mockBehavior(service)
			}

			cnt := NewDialogController(DialogControllerConfig{
				Service:   service,
				Validator: validator.NewValidator(),
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPatch, dialogDetailPath, strings.NewReader(testCase.requestBody))
			ctx := context.WithValue(
				req.Context(),
				httprouter.ParamsKey,
				httprouter.Params{{Key: "dialog_id", Value: testCase.dialogIDPathParam}},
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

func boolPtr(b bool) *bool {
	return &b
}
