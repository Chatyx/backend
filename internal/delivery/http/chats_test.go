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
	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/Mort4lis/scht-backend/pkg/validator"
	"github.com/golang/mock/gomock"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	chatCreatedAt = time.Date(2021, time.October, 25, 18, 05, 00, 0, time.Local)
	chatUpdatedAt = time.Date(2021, time.November, 17, 23, 0, 42, 142, time.Local)
)

func TestChatHandler_list(t *testing.T) {
	type mockBehaviour func(chs *mockservice.MockChatService, ctx context.Context, authUser domain.AuthUser, returnedChats []domain.Chat)

	logging.InitLogger(
		logging.LogConfig{
			LoggerKind: "mock",
		},
	)

	testTable := []struct {
		name                 string
		authUser             domain.AuthUser
		returnedChats        []domain.Chat
		mockBehaviour        mockBehaviour
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:     "Success",
			authUser: domain.AuthUser{UserID: "1"},
			returnedChats: []domain.Chat{
				{
					ID:          "1",
					Name:        "Test chat name",
					Description: "Test chat description",
					CreatorID:   "1",
					CreatedAt:   &chatCreatedAt,
					UpdatedAt:   &chatUpdatedAt,
				},
				{
					ID:        "2",
					Name:      "Another test chat name",
					CreatorID: "1",
					CreatedAt: &chatCreatedAt,
				},
			},
			mockBehaviour: func(chs *mockservice.MockChatService, ctx context.Context, authUser domain.AuthUser, returnedChats []domain.Chat) {
				chs.EXPECT().List(ctx, authUser).Return(returnedChats, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"list":[{"id":"1","name":"Test chat name","description":"Test chat description","creator_id":"1","created_at":"2021-10-25T18:05:00+03:00","updated_at":"2021-11-17T23:00:42.000000142+03:00"},{"id":"2","name":"Another test chat name","creator_id":"1","created_at":"2021-10-25T18:05:00+03:00"}]}`,
		},
		{
			name:          "Empty list",
			authUser:      domain.AuthUser{UserID: "1"},
			returnedChats: make([]domain.Chat, 0),
			mockBehaviour: func(chs *mockservice.MockChatService, ctx context.Context, authUser domain.AuthUser, returnedChats []domain.Chat) {
				chs.EXPECT().List(ctx, authUser).Return(returnedChats, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"list":[]}`,
		},
		{
			name:     "Unexpected error",
			authUser: domain.AuthUser{UserID: "1"},
			mockBehaviour: func(chs *mockservice.MockChatService, ctx context.Context, authUser domain.AuthUser, returnedChats []domain.Chat) {
				chs.EXPECT().List(ctx, authUser).Return(nil, errors.New("unexpected error"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"message":"internal server error"}`,
		},
	}

	validate, err := validator.New()
	require.NoError(t, err, "Unexpected error while creating validator")

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			chs := mockservice.NewMockChatService(c)
			chh := newChatHandler(chs, validate)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/chats", nil)
			req = req.WithContext(domain.NewContextFromAuthUser(context.Background(), testCase.authUser))

			if testCase.mockBehaviour != nil {
				testCase.mockBehaviour(chs, req.Context(), testCase.authUser, testCase.returnedChats)
			}

			chh.list(rec, req)

			resp := rec.Result()

			respBodyPayload, err := ioutil.ReadAll(resp.Body)
			assert.NoError(t, err, "Unexpected error while reading response body")

			assert.Equal(t, testCase.expectedStatusCode, resp.StatusCode)
			assert.Equal(t, testCase.expectedResponseBody, string(respBodyPayload))
		})
	}
}

func TestChatHandler_detail(t *testing.T) {
	type mockBehaviour func(chs *mockservice.MockChatService, ctx context.Context, chatID string, authUser domain.AuthUser, returnedChat domain.Chat)

	logging.InitLogger(
		logging.LogConfig{
			LoggerKind: "mock",
		},
	)

	testTable := []struct {
		name                 string
		chatID               string
		authUser             domain.AuthUser
		returnedChat         domain.Chat
		mockBehaviour        mockBehaviour
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:     "Success with full fields",
			chatID:   "1",
			authUser: domain.AuthUser{UserID: "123"},
			returnedChat: domain.Chat{
				ID:          "1",
				Name:        "Test chat name",
				Description: "Test chat description",
				CreatorID:   "1",
				CreatedAt:   &chatCreatedAt,
				UpdatedAt:   &chatUpdatedAt,
			},
			mockBehaviour: func(chs *mockservice.MockChatService, ctx context.Context, chatID string, authUser domain.AuthUser, returnedChat domain.Chat) {
				chs.EXPECT().GetByID(ctx, chatID, authUser).Return(returnedChat, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"id":"1","name":"Test chat name","description":"Test chat description","creator_id":"1","created_at":"2021-10-25T18:05:00+03:00","updated_at":"2021-11-17T23:00:42.000000142+03:00"}`,
		},
		{
			name:     "Success with required fields",
			chatID:   "2",
			authUser: domain.AuthUser{UserID: "123"},
			returnedChat: domain.Chat{
				ID:        "2",
				Name:      "Another test chat name",
				CreatorID: "1",
				CreatedAt: &chatCreatedAt,
			},
			mockBehaviour: func(chs *mockservice.MockChatService, ctx context.Context, chatID string, authUser domain.AuthUser, returnedChat domain.Chat) {
				chs.EXPECT().GetByID(ctx, chatID, authUser).Return(returnedChat, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"id":"2","name":"Another test chat name","creator_id":"1","created_at":"2021-10-25T18:05:00+03:00"}`,
		},
		{
			name:     "Not found",
			chatID:   "1",
			authUser: domain.AuthUser{UserID: "123"},
			mockBehaviour: func(chs *mockservice.MockChatService, ctx context.Context, chatID string, authUser domain.AuthUser, returnedChat domain.Chat) {
				chs.EXPECT().GetByID(ctx, chatID, authUser).Return(domain.Chat{}, domain.ErrChatNotFound)
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `{"message":"chat is not found"}`,
		},
		{
			name:     "Unexpected error",
			chatID:   "1",
			authUser: domain.AuthUser{UserID: "123"},
			mockBehaviour: func(chs *mockservice.MockChatService, ctx context.Context, chatID string, authUser domain.AuthUser, returnedChat domain.Chat) {
				chs.EXPECT().GetByID(ctx, chatID, authUser).Return(domain.Chat{}, errors.New("unexpected error"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"message":"internal server error"}`,
		},
	}

	validate, err := validator.New()
	require.NoError(t, err, "Unexpected error while creating validator")

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			chs := mockservice.NewMockChatService(c)
			chh := newChatHandler(chs, validate)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/chats/"+testCase.chatID, nil)
			req = req.WithContext(domain.NewContextFromAuthUser(context.Background(), testCase.authUser))
			ctx := context.WithValue(
				req.Context(),
				httprouter.ParamsKey,
				httprouter.Params{{Key: "chat_id", Value: testCase.chatID}},
			)

			req = req.WithContext(ctx)

			if testCase.mockBehaviour != nil {
				testCase.mockBehaviour(chs, req.Context(), testCase.chatID, testCase.authUser, testCase.returnedChat)
			}

			chh.detail(rec, req)

			resp := rec.Result()

			respBodyPayload, err := ioutil.ReadAll(resp.Body)
			assert.NoError(t, err, "Unexpected error while reading response body")

			assert.Equal(t, testCase.expectedStatusCode, resp.StatusCode)
			assert.Equal(t, testCase.expectedResponseBody, string(respBodyPayload))
		})
	}
}

func TestChatHandler_create(t *testing.T) {
	type mockBehaviour func(chs *mockservice.MockChatService, ctx context.Context, dto domain.CreateChatDTO, createdChat domain.Chat)

	logging.InitLogger(
		logging.LogConfig{
			LoggerKind: "mock",
		},
	)

	testTable := []struct {
		name                 string
		authUser             domain.AuthUser
		requestBody          string
		createChatDTO        domain.CreateChatDTO
		createdChat          domain.Chat
		mockBehaviour        mockBehaviour
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:        "Success with required fields",
			authUser:    domain.AuthUser{UserID: "123"},
			requestBody: `{"name":"Test chat name"}`,
			createChatDTO: domain.CreateChatDTO{
				Name:      "Test chat name",
				CreatorID: "123",
			},
			createdChat: domain.Chat{
				ID:        "1",
				Name:      "Test chat name",
				CreatorID: "123",
				CreatedAt: &chatCreatedAt,
			},
			mockBehaviour: func(chs *mockservice.MockChatService, ctx context.Context, dto domain.CreateChatDTO, createdChat domain.Chat) {
				chs.EXPECT().Create(ctx, dto).Return(createdChat, nil)
			},
			expectedStatusCode:   http.StatusCreated,
			expectedResponseBody: `{"id":"1","name":"Test chat name","creator_id":"123","created_at":"2021-10-25T18:05:00+03:00"}`,
		},
		{
			name:        "Success with full fields",
			authUser:    domain.AuthUser{UserID: "123"},
			requestBody: `{"name":"Test chat name","description":"Test chat description"}`,
			createChatDTO: domain.CreateChatDTO{
				Name:        "Test chat name",
				Description: "Test chat description",
				CreatorID:   "123",
			},
			createdChat: domain.Chat{
				ID:          "1",
				Name:        "Test chat name",
				Description: "Test chat description",
				CreatorID:   "123",
				CreatedAt:   &chatCreatedAt,
			},
			mockBehaviour: func(chs *mockservice.MockChatService, ctx context.Context, dto domain.CreateChatDTO, createdChat domain.Chat) {
				chs.EXPECT().Create(ctx, dto).Return(createdChat, nil)
			},
			expectedStatusCode:   http.StatusCreated,
			expectedResponseBody: `{"id":"1","name":"Test chat name","description":"Test chat description","creator_id":"123","created_at":"2021-10-25T18:05:00+03:00"}`,
		},
		{
			name:                 "Invalid JSON body",
			authUser:             domain.AuthUser{UserID: "123"},
			requestBody:          `{"name":"Test chat"`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"invalid body to decode"}`,
		},
		{
			name:                 "Empty body",
			authUser:             domain.AuthUser{UserID: "123"},
			requestBody:          `{}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"validation error","fields":{"name":"field validation for 'name' failed on the 'required' tag"}}`,
		},
		{
			name:        "Unexpected error",
			authUser:    domain.AuthUser{UserID: "123"},
			requestBody: `{"name":"Test chat name"}`,
			createChatDTO: domain.CreateChatDTO{
				Name:      "Test chat name",
				CreatorID: "123",
			},
			mockBehaviour: func(chs *mockservice.MockChatService, ctx context.Context, dto domain.CreateChatDTO, createdChat domain.Chat) {
				chs.EXPECT().Create(ctx, dto).Return(domain.Chat{}, errors.New("unexpected error"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"message":"internal server error"}`,
		},
	}

	validate, err := validator.New()
	require.NoError(t, err, "Unexpected error while creating validator")

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			chs := mockservice.NewMockChatService(c)
			chh := newChatHandler(chs, validate)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/chats", strings.NewReader(testCase.requestBody))
			req = req.WithContext(domain.NewContextFromAuthUser(context.Background(), testCase.authUser))

			if testCase.mockBehaviour != nil {
				testCase.mockBehaviour(chs, req.Context(), testCase.createChatDTO, testCase.createdChat)
			}

			chh.create(rec, req)

			resp := rec.Result()

			respBodyPayload, err := ioutil.ReadAll(resp.Body)
			assert.NoError(t, err, "Unexpected error while reading response body")

			assert.Equal(t, testCase.expectedStatusCode, resp.StatusCode)
			assert.Equal(t, testCase.expectedResponseBody, string(respBodyPayload))
		})
	}
}

func TestChatHandler_update(t *testing.T) {
	type mockBehaviour func(chs *mockservice.MockChatService, ctx context.Context, dto domain.UpdateChatDTO, updatedChat domain.Chat)

	logging.InitLogger(
		logging.LogConfig{
			LoggerKind: "mock",
		},
	)

	testTable := []struct {
		name                 string
		chatID               string
		authUser             domain.AuthUser
		requestBody          string
		updateChatDTO        domain.UpdateChatDTO
		updatedChat          domain.Chat
		mockBehaviour        mockBehaviour
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:        "Success with required fields",
			chatID:      "1",
			authUser:    domain.AuthUser{UserID: "123"},
			requestBody: `{"name":"Test chat name"}`,
			updateChatDTO: domain.UpdateChatDTO{
				ID:        "1",
				Name:      "Test chat name",
				CreatorID: "123",
			},
			updatedChat: domain.Chat{
				ID:        "1",
				Name:      "Test chat name",
				CreatorID: "123",
				CreatedAt: &chatCreatedAt,
				UpdatedAt: &chatUpdatedAt,
			},
			mockBehaviour: func(chs *mockservice.MockChatService, ctx context.Context, dto domain.UpdateChatDTO, updatedChat domain.Chat) {
				chs.EXPECT().Update(ctx, dto).Return(updatedChat, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"id":"1","name":"Test chat name","creator_id":"123","created_at":"2021-10-25T18:05:00+03:00","updated_at":"2021-11-17T23:00:42.000000142+03:00"}`,
		},
		{
			name:        "Success with full fields",
			chatID:      "2",
			authUser:    domain.AuthUser{UserID: "123"},
			requestBody: `{"name":"Test chat name","description":"Test chat description"}`,
			updateChatDTO: domain.UpdateChatDTO{
				ID:          "2",
				Name:        "Test chat name",
				Description: "Test chat description",
				CreatorID:   "123",
			},
			updatedChat: domain.Chat{
				ID:          "2",
				Name:        "Test chat name",
				Description: "Test chat description",
				CreatorID:   "123",
				CreatedAt:   &chatCreatedAt,
				UpdatedAt:   &chatUpdatedAt,
			},
			mockBehaviour: func(chs *mockservice.MockChatService, ctx context.Context, dto domain.UpdateChatDTO, updatedChat domain.Chat) {
				chs.EXPECT().Update(ctx, dto).Return(updatedChat, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"id":"2","name":"Test chat name","description":"Test chat description","creator_id":"123","created_at":"2021-10-25T18:05:00+03:00","updated_at":"2021-11-17T23:00:42.000000142+03:00"}`,
		},
		{
			name:        "Chat is not found",
			chatID:      "1",
			authUser:    domain.AuthUser{UserID: "123"},
			requestBody: `{"name":"Test chat name"}`,
			updateChatDTO: domain.UpdateChatDTO{
				ID:        "1",
				Name:      "Test chat name",
				CreatorID: "123",
			},
			mockBehaviour: func(chs *mockservice.MockChatService, ctx context.Context, dto domain.UpdateChatDTO, updatedChat domain.Chat) {
				chs.EXPECT().Update(ctx, dto).Return(domain.Chat{}, domain.ErrChatNotFound)
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `{"message":"chat is not found"}`,
		},
		{
			name:                 "Invalid JSON body",
			chatID:               "1",
			authUser:             domain.AuthUser{UserID: "123"},
			requestBody:          `{"name":"Test chat"`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"invalid body to decode"}`,
		},
		{
			name:                 "Empty body",
			chatID:               "1",
			authUser:             domain.AuthUser{UserID: "123"},
			requestBody:          `{}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"validation error","fields":{"name":"field validation for 'name' failed on the 'required' tag"}}`,
		},
		{
			name:        "Unexpected error",
			chatID:      "1",
			authUser:    domain.AuthUser{UserID: "123"},
			requestBody: `{"name":"Test chat name"}`,
			updateChatDTO: domain.UpdateChatDTO{
				ID:        "1",
				Name:      "Test chat name",
				CreatorID: "123",
			},
			mockBehaviour: func(chs *mockservice.MockChatService, ctx context.Context, dto domain.UpdateChatDTO, updatedChat domain.Chat) {
				chs.EXPECT().Update(ctx, dto).Return(domain.Chat{}, errors.New("unexpected error"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"message":"internal server error"}`,
		},
	}

	validate, err := validator.New()
	require.NoError(t, err, "Unexpected error while creating validator")

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			chs := mockservice.NewMockChatService(c)
			chh := newChatHandler(chs, validate)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, "/api/chats/"+testCase.chatID, strings.NewReader(testCase.requestBody))
			ctx := context.WithValue(
				req.Context(),
				httprouter.ParamsKey,
				httprouter.Params{{Key: "chat_id", Value: testCase.chatID}},
			)

			ctx = domain.NewContextFromAuthUser(ctx, testCase.authUser)
			req = req.WithContext(ctx)

			if testCase.mockBehaviour != nil {
				testCase.mockBehaviour(chs, req.Context(), testCase.updateChatDTO, testCase.updatedChat)
			}

			chh.update(rec, req)

			resp := rec.Result()

			respBodyPayload, err := ioutil.ReadAll(resp.Body)
			assert.NoError(t, err, "Unexpected error while reading response body")

			assert.Equal(t, testCase.expectedStatusCode, resp.StatusCode)
			assert.Equal(t, testCase.expectedResponseBody, string(respBodyPayload))
		})
	}
}

func TestChatHandler_delete(t *testing.T) {
	type mockBehaviour func(chs *mockservice.MockChatService, ctx context.Context, chatID string, authUser domain.AuthUser)

	logging.InitLogger(
		logging.LogConfig{
			LoggerKind: "mock",
		},
	)

	testTable := []struct {
		name                 string
		chatID               string
		authUser             domain.AuthUser
		mockBehaviour        mockBehaviour
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:     "Success",
			chatID:   "1",
			authUser: domain.AuthUser{UserID: "123"},
			mockBehaviour: func(chs *mockservice.MockChatService, ctx context.Context, chatID string, authUser domain.AuthUser) {
				chs.EXPECT().Delete(ctx, chatID, authUser).Return(nil)
			},
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name:     "Not found",
			chatID:   "2",
			authUser: domain.AuthUser{UserID: "123"},
			mockBehaviour: func(chs *mockservice.MockChatService, ctx context.Context, chatID string, authUser domain.AuthUser) {
				chs.EXPECT().Delete(ctx, chatID, authUser).Return(domain.ErrChatNotFound)
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedResponseBody: `{"message":"chat is not found"}`,
		},
		{
			name:     "Unexpected error",
			chatID:   "2",
			authUser: domain.AuthUser{UserID: "123"},
			mockBehaviour: func(chs *mockservice.MockChatService, ctx context.Context, chatID string, authUser domain.AuthUser) {
				chs.EXPECT().Delete(ctx, chatID, authUser).Return(errors.New("unexpected error"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"message":"internal server error"}`,
		},
	}

	validate, err := validator.New()
	require.NoError(t, err, "Unexpected error while creating validator")

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			chs := mockservice.NewMockChatService(c)
			chh := newChatHandler(chs, validate)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/api/chats/"+testCase.chatID, nil)
			ctx := context.WithValue(
				req.Context(),
				httprouter.ParamsKey,
				httprouter.Params{{Key: "chat_id", Value: testCase.chatID}},
			)

			ctx = domain.NewContextFromAuthUser(ctx, testCase.authUser)
			req = req.WithContext(ctx)

			if testCase.mockBehaviour != nil {
				testCase.mockBehaviour(chs, req.Context(), testCase.chatID, testCase.authUser)
			}

			chh.delete(rec, req)

			resp := rec.Result()

			respBodyPayload, err := ioutil.ReadAll(resp.Body)
			assert.NoError(t, err, "Unexpected error while reading response body")

			assert.Equal(t, testCase.expectedStatusCode, resp.StatusCode)
			assert.Equal(t, testCase.expectedResponseBody, string(respBodyPayload))
		})
	}
}
