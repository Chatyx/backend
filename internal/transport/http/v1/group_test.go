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

func TestGroupController_list(t *testing.T) {
	testCases := []struct {
		name                 string
		mockBehavior         func(s *MockGroupService)
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name: "Successful",
			mockBehavior: func(s *MockGroupService) {
				s.On("List", mock.Anything).Return([]entity.Group{
					{
						ID:          1,
						Name:        "Test1",
						Description: "Test1 group description",
						CreatedAt:   defaultCreatedAt,
					},
					{
						ID:          2,
						Name:        "Test2",
						Description: "Test2 group description",
						CreatedAt:   defaultCreatedAt,
					},
				}, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"total":2,"data":[{"id":1,"name":"Test1","description":"Test1 group description","created_at":"2024-01-23T00:00:00Z"},{"id":2,"name":"Test2","description":"Test2 group description","created_at":"2024-01-23T00:00:00Z"}]}`,
		},
		{
			name: "Successful with empty list",
			mockBehavior: func(s *MockGroupService) {
				s.On("List", mock.Anything).Return(nil, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"total":0,"data":[]}`,
		},
		{
			name: "Internal server error",
			mockBehavior: func(s *MockGroupService) {
				s.On("List", mock.Anything).Return(nil, errUnexpected)
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"code":"CM0001","message":"internal server error"}`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			service := NewMockGroupService(t)
			if testCase.mockBehavior != nil {
				testCase.mockBehavior(service)
			}

			cnt := NewGroupController(GroupControllerConfig{
				Service:   service,
				Validator: validator.NewValidator(),
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, groupListPath, nil)

			cnt.list(rec, req)
			resp := rec.Result()

			assert.Equal(t, testCase.expectedStatusCode, resp.StatusCode)

			respBody, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedResponseBody, string(respBody))
		})
	}
}

func TestGroupController_create(t *testing.T) {
	testCases := []struct {
		name                 string
		requestBody          string
		mockBehavior         func(s *MockGroupService)
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:        "Successful with required fields",
			requestBody: `{"name":"Test1"}`,
			mockBehavior: func(s *MockGroupService) {
				s.On("Create", mock.Anything, dto.GroupCreate{
					Name: "Test1",
				}).Return(entity.Group{
					ID:        1,
					Name:      "Test1",
					CreatedAt: defaultCreatedAt,
				}, nil)
			},
			expectedStatusCode:   http.StatusCreated,
			expectedResponseBody: `{"id":1,"name":"Test1","created_at":"2024-01-23T00:00:00Z"}`,
		},
		{
			name:        "Successful with all fields",
			requestBody: `{"name":"Test1","description":"Test1 group description"}`,
			mockBehavior: func(s *MockGroupService) {
				s.On("Create", mock.Anything, dto.GroupCreate{
					Name:        "Test1",
					Description: "Test1 group description",
				}).Return(entity.Group{
					ID:          1,
					Name:        "Test1",
					Description: "Test1 group description",
					CreatedAt:   defaultCreatedAt,
				}, nil)
			},
			expectedStatusCode:   http.StatusCreated,
			expectedResponseBody: `{"id":1,"name":"Test1","description":"Test1 group description","created_at":"2024-01-23T00:00:00Z"}`,
		},
		{
			name:                 "Decode body error",
			requestBody:          `{"name":"Test1",`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0002","message":"decode body error"}`,
		},
		{
			name:                 "Validation error: name is required",
			requestBody:          `{}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0006","message":"validation error","data":{"name":"failed on the 'required' tag"}}`,
		},
		{
			name:        "Internal server error",
			requestBody: `{"name":"Test1","description":"Test1 group description"}`,
			mockBehavior: func(s *MockGroupService) {
				s.On("Create", mock.Anything, dto.GroupCreate{
					Name:        "Test1",
					Description: "Test1 group description",
				}).Return(entity.Group{}, errUnexpected)
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"code":"CM0001","message":"internal server error"}`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			service := NewMockGroupService(t)
			if testCase.mockBehavior != nil {
				testCase.mockBehavior(service)
			}

			cnt := NewGroupController(GroupControllerConfig{
				Service:   service,
				Validator: validator.NewValidator(),
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, groupListPath, strings.NewReader(testCase.requestBody))

			cnt.create(rec, req)
			resp := rec.Result()

			assert.Equal(t, testCase.expectedStatusCode, resp.StatusCode)

			respBody, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedResponseBody, string(respBody))
		})
	}
}

func TestGroupController_detail(t *testing.T) {
	testCases := []struct {
		name                 string
		groupIDPathParam     string
		mockBehavior         func(s *MockGroupService)
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:             "Success with required fields",
			groupIDPathParam: "1",
			mockBehavior: func(s *MockGroupService) {
				s.On("GetByID", mock.Anything, 1).Return(entity.Group{
					ID:        1,
					Name:      "Test1",
					CreatedAt: defaultCreatedAt,
				}, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"id":1,"name":"Test1","created_at":"2024-01-23T00:00:00Z"}`,
		},
		{
			name:             "Success with all fields",
			groupIDPathParam: "1",
			mockBehavior: func(s *MockGroupService) {
				s.On("GetByID", mock.Anything, 1).Return(entity.Group{
					ID:          1,
					Name:        "Test1",
					Description: "Test1 group description",
					CreatedAt:   defaultCreatedAt,
				}, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"id":1,"name":"Test1","description":"Test1 group description","created_at":"2024-01-23T00:00:00Z"}`,
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
			mockBehavior: func(s *MockGroupService) {
				s.On("GetByID", mock.Anything, 1).Return(entity.Group{}, entity.ErrGroupNotFound)
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `{"code":"CH0001","message":"group is not found"}`,
		},
		{
			name:             "Internal server error",
			groupIDPathParam: "1",
			mockBehavior: func(s *MockGroupService) {
				s.On("GetByID", mock.Anything, 1).Return(entity.Group{}, errUnexpected)
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"code":"CM0001","message":"internal server error"}`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			service := NewMockGroupService(t)
			if testCase.mockBehavior != nil {
				testCase.mockBehavior(service)
			}

			cnt := NewGroupController(GroupControllerConfig{
				Service:   service,
				Validator: validator.NewValidator(),
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, groupDetailPath, nil)
			ctx := context.WithValue(
				req.Context(),
				httprouter.ParamsKey,
				httprouter.Params{{Key: "group_id", Value: testCase.groupIDPathParam}},
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

func TestGroupController_update(t *testing.T) {
	testCases := []struct {
		name                 string
		groupIDPathParam     string
		requestBody          string
		mockBehavior         func(s *MockGroupService)
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:             "Successful with required fields",
			groupIDPathParam: "1",
			requestBody:      `{"name":"Test1"}`,
			mockBehavior: func(s *MockGroupService) {
				s.On("Update", mock.Anything, dto.GroupUpdate{
					ID:   1,
					Name: "Test1",
				}).Return(entity.Group{
					ID:        1,
					Name:      "Test1",
					CreatedAt: defaultCreatedAt,
				}, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"id":1,"name":"Test1","created_at":"2024-01-23T00:00:00Z"}`,
		},
		{
			name:             "Successful with all fields",
			groupIDPathParam: "1",
			requestBody:      `{"name":"Test1","description":"Test1 group description"}`,
			mockBehavior: func(s *MockGroupService) {
				s.On("Update", mock.Anything, dto.GroupUpdate{
					ID:          1,
					Name:        "Test1",
					Description: "Test1 group description",
				}).Return(entity.Group{
					ID:          1,
					Name:        "Test1",
					Description: "Test1 group description",
					CreatedAt:   defaultCreatedAt,
				}, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"id":1,"name":"Test1","description":"Test1 group description","created_at":"2024-01-23T00:00:00Z"}`,
		},
		{
			name:                 "Decode path param error",
			groupIDPathParam:     uuid.New().String(),
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0003","message":"decode path params error"}`,
		},
		{
			name:                 "Decode body error",
			groupIDPathParam:     "1",
			requestBody:          `{"name":"Test1","description":"Test1 group description"`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0002","message":"decode body error"}`,
		},
		{
			name:                 "Validation error: name is required",
			groupIDPathParam:     "1",
			requestBody:          `{}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0006","message":"validation error","data":{"name":"failed on the 'required' tag"}}`,
		},
		{
			name:             "Group is not found",
			groupIDPathParam: "1",
			requestBody:      `{"name":"Test1","description":"Test1 group description"}`,
			mockBehavior: func(s *MockGroupService) {
				s.On("Update", mock.Anything, dto.GroupUpdate{
					ID:          1,
					Name:        "Test1",
					Description: "Test1 group description",
				}).Return(entity.Group{}, entity.ErrGroupNotFound)
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `{"code":"CH0001","message":"group is not found"}`,
		},
		{
			name:             "Internal server error",
			groupIDPathParam: "1",
			requestBody:      `{"name":"Test1","description":"Test1 group description"}`,
			mockBehavior: func(s *MockGroupService) {
				s.On("Update", mock.Anything, dto.GroupUpdate{
					ID:          1,
					Name:        "Test1",
					Description: "Test1 group description",
				}).Return(entity.Group{}, errUnexpected)
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"code":"CM0001","message":"internal server error"}`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			service := NewMockGroupService(t)
			if testCase.mockBehavior != nil {
				testCase.mockBehavior(service)
			}

			cnt := NewGroupController(GroupControllerConfig{
				Service:   service,
				Validator: validator.NewValidator(),
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, groupDetailPath, strings.NewReader(testCase.requestBody))
			ctx := context.WithValue(
				req.Context(),
				httprouter.ParamsKey,
				httprouter.Params{{Key: "group_id", Value: testCase.groupIDPathParam}},
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

func TestGroupController_delete(t *testing.T) {
	testCases := []struct {
		name                 string
		groupIDPathParam     string
		mockBehavior         func(s *MockGroupService)
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:             "Successful",
			groupIDPathParam: "1",
			mockBehavior: func(s *MockGroupService) {
				s.On("Delete", mock.Anything, 1).Return(nil)
			},
			expectedStatusCode: http.StatusNoContent,
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
			mockBehavior: func(s *MockGroupService) {
				s.On("Delete", mock.Anything, 1).Return(entity.ErrGroupNotFound)
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `{"code":"CH0001","message":"group is not found"}`,
		},
		{
			name:             "Internal server error",
			groupIDPathParam: "1",
			mockBehavior: func(s *MockGroupService) {
				s.On("Delete", mock.Anything, 1).Return(errUnexpected)
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"code":"CM0001","message":"internal server error"}`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			service := NewMockGroupService(t)
			if testCase.mockBehavior != nil {
				testCase.mockBehavior(service)
			}

			cnt := NewGroupController(GroupControllerConfig{
				Service:   service,
				Validator: validator.NewValidator(),
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, groupDetailPath, nil)
			ctx := context.WithValue(
				req.Context(),
				httprouter.ParamsKey,
				httprouter.Params{{Key: "group_id", Value: testCase.groupIDPathParam}},
			)
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
