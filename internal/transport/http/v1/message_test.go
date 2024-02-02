package v1

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Chatyx/backend/internal/dto"
	"github.com/Chatyx/backend/internal/entity"
	"github.com/Chatyx/backend/pkg/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMessageController_list(t *testing.T) {
	testCases := []struct {
		name                 string
		queryBehavior        func(q url.Values)
		mockBehavior         func(s *MockMessageService)
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name: "Successful",
			queryBehavior: func(query url.Values) {
				query.Add(chatIDParam, "1")
				query.Add(chatTypeParam, "dialog")
				query.Add(idAfterParam, "1")
				query.Add(limitParam, "30")
				query.Add(sortParam, "asc")
			},
			mockBehavior: func(s *MockMessageService) {
				dtoObj := dto.MessageList{
					ChatID:  entity.ChatID{ID: 1, Type: entity.DialogChatType},
					IDAfter: 1,
					Limit:   30,
					Sort:    dto.AscSort,
				}
				s.On("List", mock.Anything, dtoObj).Return([]entity.Message{
					{
						ID:          2,
						ChatID:      dtoObj.ChatID,
						SenderID:    1,
						Content:     "hello",
						ContentType: entity.TextContentType,
						IsService:   false,
						SentAt:      defaultCreatedAt,
					},
					{
						ID:          3,
						ChatID:      dtoObj.ChatID,
						SenderID:    1,
						Content:     "world",
						ContentType: entity.TextContentType,
						IsService:   false,
						SentAt:      defaultCreatedAt,
					},
				}, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"total":2,"data":[{"id":2,"sender_id":1,"content":"hello","content_type":"text","is_service":false,"sent_at":"2024-01-23T00:00:00Z"},{"id":3,"sender_id":1,"content":"world","content_type":"text","is_service":false,"sent_at":"2024-01-23T00:00:00Z"}]}`,
		},
		{
			name: "Successful with empty list",
			queryBehavior: func(query url.Values) {
				query.Add(chatIDParam, "1")
				query.Add(chatTypeParam, "dialog")
				query.Add(idAfterParam, "32")
				query.Add(limitParam, "100")
				query.Add(sortParam, "desc")
			},
			mockBehavior: func(s *MockMessageService) {
				dtoObj := dto.MessageList{
					ChatID:  entity.ChatID{ID: 1, Type: entity.DialogChatType},
					IDAfter: 32,
					Limit:   100,
					Sort:    dto.DescSort,
				}
				s.On("List", mock.Anything, dtoObj).Return(nil, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"total":0,"data":[]}`,
		},
		{
			name: "Decode chat_id query error",
			queryBehavior: func(query url.Values) {
				query.Add(chatIDParam, "1abc")
				query.Add(chatTypeParam, "dialog")
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0005","message":"decode query params error","data":{"chat_id":"failed to parse int"}}`,
		},
		{
			name: "Decode id_after query error",
			queryBehavior: func(query url.Values) {
				query.Add(chatIDParam, "1")
				query.Add(chatTypeParam, "dialog")
				query.Add(idAfterParam, "abc")
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0005","message":"decode query params error","data":{"id_after":"failed to parse int"}}`,
		},
		{
			name: "Decode limit query error",
			queryBehavior: func(query url.Values) {
				query.Add(chatIDParam, "1")
				query.Add(chatTypeParam, "dialog")
				query.Add(limitParam, "---")
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0005","message":"decode query params error","data":{"limit":"failed to parse int"}}`,
		},
		{
			name: "Validation error: chat_id, chat_type, sort are required",
			queryBehavior: func(query url.Values) {
				query.Add(chatIDParam, "0")
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0006","message":"validation error","data":{"chat_id":"failed on the 'required' tag","chat_type":"failed on the 'required' tag","sort":"failed on the 'required' tag"}}`,
		},
		{
			name: "Validation error: chat_type is dialog or group",
			queryBehavior: func(query url.Values) {
				query.Add(chatIDParam, "1")
				query.Add(chatTypeParam, "channel")
				query.Add(sortParam, "asc")
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0006","message":"validation error","data":{"chat_type":"failed on the 'oneof' tag"}}`,
		},
		{
			name: "Validation error: sort is asc or desc",
			queryBehavior: func(query url.Values) {
				query.Add(chatIDParam, "1")
				query.Add(chatTypeParam, "dialog")
				query.Add(sortParam, "-")
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0006","message":"validation error","data":{"sort":"failed on the 'oneof' tag"}}`,
		},
		{
			name: "Validation error: limit exceeded",
			queryBehavior: func(query url.Values) {
				query.Add(chatIDParam, "1")
				query.Add(chatTypeParam, "dialog")
				query.Add(limitParam, "1000")
				query.Add(sortParam, "asc")
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0006","message":"validation error","data":{"limit":"failed on the 'max' tag"}}`,
		},
		{
			name: "Group is not found",
			queryBehavior: func(query url.Values) {
				query.Add(chatIDParam, "2")
				query.Add(chatTypeParam, "group")
				query.Add(sortParam, "asc")
			},
			mockBehavior: func(s *MockMessageService) {
				dtoObj := dto.MessageList{
					ChatID: entity.ChatID{ID: 2, Type: entity.GroupChatType},
					Limit:  defaultLimit,
					Sort:   "asc",
				}
				s.On("List", mock.Anything, dtoObj).Return(nil, entity.ErrGroupNotFound)
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `{"code":"CH0001","message":"group is not found"}`,
		},
		{
			name: "Dialog is not found",
			queryBehavior: func(query url.Values) {
				query.Add(chatIDParam, "1")
				query.Add(chatTypeParam, "dialog")
				query.Add(sortParam, "asc")
			},
			mockBehavior: func(s *MockMessageService) {
				dtoObj := dto.MessageList{
					ChatID: entity.ChatID{ID: 1, Type: entity.DialogChatType},
					Limit:  defaultLimit,
					Sort:   "asc",
				}
				s.On("List", mock.Anything, dtoObj).Return(nil, entity.ErrDialogNotFound)
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `{"code":"CH0002","message":"dialog is not found"}`,
		},
		{
			name: "Internal server error",
			queryBehavior: func(query url.Values) {
				query.Add(chatIDParam, "1")
				query.Add(chatTypeParam, "dialog")
				query.Add(limitParam, "20")
				query.Add(sortParam, "asc")
			},
			mockBehavior: func(s *MockMessageService) {
				dtoObj := dto.MessageList{
					ChatID: entity.ChatID{ID: 1, Type: entity.DialogChatType},
					Limit:  defaultLimit,
					Sort:   "asc",
				}
				s.On("List", mock.Anything, dtoObj).Return(nil, errUnexpected)
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"code":"CM0001","message":"internal server error"}`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			service := NewMockMessageService(t)
			if testCase.mockBehavior != nil {
				testCase.mockBehavior(service)
			}

			cnt := NewMessageController(MessageControllerConfig{
				Service:   service,
				Validator: validator.NewValidator(),
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, messageListPath, nil)

			query := req.URL.Query()
			if testCase.queryBehavior != nil {
				testCase.queryBehavior(query)
			}
			req.URL.RawQuery = query.Encode()

			cnt.list(rec, req)
			resp := rec.Result()

			assert.Equal(t, testCase.expectedStatusCode, resp.StatusCode)

			respBody, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedResponseBody, string(respBody))
		})
	}
}

func TestMessageController_create(t *testing.T) {
	testCases := []struct {
		name                 string
		requestBody          string
		queryBehavior        func(q url.Values)
		mockBehavior         func(s *MockMessageService)
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:        "Successful",
			requestBody: `{"content":"hello","content_type":"text"}`,
			queryBehavior: func(query url.Values) {
				query.Add(chatIDParam, "1")
				query.Add(chatTypeParam, "dialog")
			},
			mockBehavior: func(s *MockMessageService) {
				s.On("Create", mock.Anything, dto.MessageCreate{
					ChatID:      entity.ChatID{ID: 1, Type: entity.DialogChatType},
					Content:     "hello",
					ContentType: entity.TextContentType,
				}).Return(entity.Message{
					ID:          1,
					ChatID:      entity.ChatID{ID: 1, Type: entity.DialogChatType},
					SenderID:    1,
					Content:     "hello",
					ContentType: entity.TextContentType,
					IsService:   false,
					SentAt:      defaultCreatedAt,
				}, nil)
			},
			expectedStatusCode:   http.StatusCreated,
			expectedResponseBody: `{"id":1,"sender_id":1,"content":"hello","content_type":"text","is_service":false,"sent_at":"2024-01-23T00:00:00Z"}`,
		},
		{
			name:        "Decode body error",
			requestBody: `{"content":"hello","content_type":"text"`,
			queryBehavior: func(query url.Values) {
				query.Add(chatIDParam, "1")
				query.Add(chatTypeParam, "dialog")
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0002","message":"decode body error"}`,
		},
		{
			name:        "Decode chat_id query error",
			requestBody: `{"content":"hello","content_type":"text"}`,
			queryBehavior: func(query url.Values) {
				query.Add(chatIDParam, "1a")
				query.Add(chatTypeParam, "dialog")
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0005","message":"decode query params error","data":{"chat_id":"failed to parse int"}}`,
		},
		{
			name:        "Validation error: chat_id, chat_type are required",
			requestBody: `{"content":"hello","content_type":"text"}`,
			queryBehavior: func(query url.Values) {
				query.Add(chatIDParam, "0")
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0006","message":"validation error","data":{"chat_id":"failed on the 'required' tag","chat_type":"failed on the 'required' tag"}}`,
		},
		{
			name:        "Validation error: chat_type is dialog or group",
			requestBody: `{"content":"hello","content_type":"text"}`,
			queryBehavior: func(query url.Values) {
				query.Add(chatIDParam, "1")
				query.Add(chatTypeParam, "channel")
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0006","message":"validation error","data":{"chat_type":"failed on the 'oneof' tag"}}`,
		},
		{
			name:        "Group is not found",
			requestBody: `{"content":"hello","content_type":"text"}`,
			queryBehavior: func(query url.Values) {
				query.Add(chatIDParam, "2")
				query.Add(chatTypeParam, "group")
			},
			mockBehavior: func(s *MockMessageService) {
				s.On("Create", mock.Anything, dto.MessageCreate{
					ChatID:      entity.ChatID{ID: 2, Type: entity.GroupChatType},
					Content:     "hello",
					ContentType: entity.TextContentType,
				}).Return(entity.Message{}, entity.ErrGroupNotFound)
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `{"code":"CH0001","message":"group is not found"}`,
		},
		{
			name:        "Dialog is not found",
			requestBody: `{"content":"hello","content_type":"text"}`,
			queryBehavior: func(query url.Values) {
				query.Add(chatIDParam, "1")
				query.Add(chatTypeParam, "dialog")
			},
			mockBehavior: func(s *MockMessageService) {
				s.On("Create", mock.Anything, dto.MessageCreate{
					ChatID:      entity.ChatID{ID: 1, Type: entity.DialogChatType},
					Content:     "hello",
					ContentType: entity.TextContentType,
				}).Return(entity.Message{}, entity.ErrDialogNotFound)
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `{"code":"CH0002","message":"dialog is not found"}`,
		},
		{
			name:        "Internal server error",
			requestBody: `{"content":"hello","content_type":"text"}`,
			queryBehavior: func(query url.Values) {
				query.Add(chatIDParam, "1")
				query.Add(chatTypeParam, "dialog")
			},
			mockBehavior: func(s *MockMessageService) {
				s.On("Create", mock.Anything, dto.MessageCreate{
					ChatID:      entity.ChatID{ID: 1, Type: entity.DialogChatType},
					Content:     "hello",
					ContentType: entity.TextContentType,
				}).Return(entity.Message{}, errUnexpected)
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"code":"CM0001","message":"internal server error"}`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			service := NewMockMessageService(t)
			if testCase.mockBehavior != nil {
				testCase.mockBehavior(service)
			}

			cnt := NewMessageController(MessageControllerConfig{
				Service:   service,
				Validator: validator.NewValidator(),
			})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, messageListPath, strings.NewReader(testCase.requestBody))

			query := req.URL.Query()
			if testCase.queryBehavior != nil {
				testCase.queryBehavior(query)
			}
			req.URL.RawQuery = query.Encode()

			cnt.create(rec, req)
			resp := rec.Result()

			assert.Equal(t, testCase.expectedStatusCode, resp.StatusCode)

			respBody, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedResponseBody, string(respBody))
		})
	}
}
