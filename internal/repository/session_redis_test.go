//go:build unit
// +build unit

package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/go-redis/redismock/v8"
)

const refreshTokenTTL = 15 * 24 * time.Hour // 15 days

var exampleSession = domain.Session{
	UserID:       "1",
	RefreshToken: "qGVFLRQw37TnSmG0LKFN",
	Fingerprint:  "5dc49b7a-6153-4eae-9c0f-297655c45f08",
	CreatedAt:    time.Date(2021, time.October, 2, 14, 47, 12, 0, time.Local),
	ExpiresAt:    time.Date(2021, time.November, 1, 14, 47, 12, 0, time.Local),
}

func TestSessionRedisRepository_Get(t *testing.T) {
	type mockBehavior func(mock redismock.ClientMock, refreshToken, value string)

	logging.InitLogger(
		logging.LogConfig{
			LoggerKind: "mock",
		},
	)

	testTable := []struct {
		name            string
		refreshToken    string
		value           string
		mockBehavior    mockBehavior
		expectedSession domain.Session
		expectedErr     error
	}{
		{
			name:         "Success",
			refreshToken: "qGVFLRQw37TnSmG0LKFN",
			value:        `{"UserID":"1","RefreshToken":"qGVFLRQw37TnSmG0LKFN","Fingerprint":"5dc49b7a-6153-4eae-9c0f-297655c45f08","CreatedAt":"2021-10-02T14:47:12.00+03:00","ExpiresAt":"2021-11-01T14:47:12.00+03:00"}`,
			mockBehavior: func(mock redismock.ClientMock, refreshToken, value string) {
				mock.ExpectGet("session:" + refreshToken).SetVal(value)
			},
			expectedSession: exampleSession,
			expectedErr:     nil,
		},
		{
			name:         "Not found",
			refreshToken: "qGVFLRQw37TnSmG0LKFN",
			mockBehavior: func(mock redismock.ClientMock, refreshToken, value string) {
				mock.ExpectGet("session:" + refreshToken).RedisNil()
			},
			expectedErr: domain.ErrSessionNotFound,
		},
		{
			name:         "Unexpected error while getting session",
			refreshToken: "qGVFLRQw37TnSmG0LKFN",
			mockBehavior: func(mock redismock.ClientMock, refreshToken, value string) {
				mock.ExpectGet("session:" + refreshToken).SetErr(errUnexpected)
			},
			expectedErr: errUnexpected,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			client, mock := redismock.NewClientMock()
			sessionRepo := NewSessionRedisRepository(client)

			if testCase.mockBehavior != nil {
				testCase.mockBehavior(mock, testCase.refreshToken, testCase.value)
			}

			session, err := sessionRepo.Get(context.Background(), testCase.refreshToken)

			if testCase.expectedErr == nil {
				assert.NoError(t, err)
			}

			if testCase.expectedErr != nil {
				assert.Error(t, err)
				if testCase.expectedErr != errUnexpected {
					assert.ErrorIs(t, err, testCase.expectedErr)
				}
			}

			assert.Equal(t, testCase.expectedSession, session)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err, "There were unfulfilled expectations")
		})
	}
}

func TestSessionRedisRepository_Set(t *testing.T) {
	type mockBehavior func(mock redismock.ClientMock, session domain.Session, ttl time.Duration)

	logging.InitLogger(
		logging.LogConfig{
			LoggerKind: "mock",
		},
	)

	testTable := []struct {
		name         string
		session      domain.Session
		mockBehavior mockBehavior
		expectedErr  error
	}{
		{
			name:    "Success",
			session: exampleSession,
			mockBehavior: func(mock redismock.ClientMock, session domain.Session, ttl time.Duration) {
				sessionKey := "session:" + session.RefreshToken
				userSessionsKey := fmt.Sprintf("user:%s:sessions", session.UserID)

				payload, _ := json.Marshal(session)
				mock.ExpectSet(sessionKey, payload, ttl).SetVal("OK")
				mock.ExpectRPush(userSessionsKey, sessionKey).SetVal(1)
			},
			expectedErr: nil,
		},
		{
			name:    "Unexpected error while setting session",
			session: exampleSession,
			mockBehavior: func(mock redismock.ClientMock, session domain.Session, ttl time.Duration) {
				sessionKey := "session:" + session.RefreshToken

				payload, _ := json.Marshal(session)
				mock.ExpectSet(sessionKey, payload, ttl).SetErr(errUnexpected)
			},
			expectedErr: errUnexpected,
		},
		{
			name:    "Unexpected error while pushing session to the list of user's sessions",
			session: exampleSession,
			mockBehavior: func(mock redismock.ClientMock, session domain.Session, ttl time.Duration) {
				sessionKey := "session:" + session.RefreshToken
				userSessionsKey := fmt.Sprintf("user:%s:sessions", session.UserID)

				payload, _ := json.Marshal(session)
				mock.ExpectSet(sessionKey, payload, ttl).SetVal("OK")
				mock.ExpectRPush(userSessionsKey, sessionKey).SetErr(errUnexpected)
			},
			expectedErr: errUnexpected,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			client, mock := redismock.NewClientMock()
			sessionRepo := NewSessionRedisRepository(client)

			if testCase.mockBehavior != nil {
				testCase.mockBehavior(mock, testCase.session, refreshTokenTTL)
			}

			err := sessionRepo.Set(context.Background(), testCase.session, refreshTokenTTL)

			if testCase.expectedErr == nil {
				assert.NoError(t, err)
			}

			if testCase.expectedErr != nil {
				assert.Error(t, err)
				if testCase.expectedErr != errUnexpected {
					assert.ErrorIs(t, err, testCase.expectedErr)
				}
			}

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err, "There were unfulfilled expectations")
		})
	}
}

func TestSessionRedisRepository_Delete(t *testing.T) {
	type mockBehavior func(mock redismock.ClientMock, refreshToken, userID string)

	logging.InitLogger(
		logging.LogConfig{
			LoggerKind: "mock",
		},
	)

	testTable := []struct {
		name         string
		refreshToken string
		userID       string
		mockBehavior mockBehavior
		expectedErr  error
	}{
		{
			name:         "Success",
			refreshToken: "qGVFLRQw37TnSmG0LKFN",
			userID:       "1",
			mockBehavior: func(mock redismock.ClientMock, refreshToken, userID string) {
				sessionKey := "session:" + refreshToken
				userSessionsKey := fmt.Sprintf("user:%s:sessions", userID)

				mock.ExpectDel(sessionKey).SetVal(1)
				mock.ExpectLRem(userSessionsKey, 0, sessionKey).SetVal(1)
			},
			expectedErr: nil,
		},
		{
			name:         "Session is not found",
			refreshToken: "qGVFLRQw37TnSmG0LKFN",
			userID:       "1",
			mockBehavior: func(mock redismock.ClientMock, refreshToken, userID string) {
				sessionKey := "session:" + refreshToken

				mock.ExpectDel(sessionKey).SetVal(0)
			},
			expectedErr: domain.ErrSessionNotFound,
		},
		{
			name:         "Unexpected error while deleting session by key",
			refreshToken: "qGVFLRQw37TnSmG0LKFN",
			userID:       "1",
			mockBehavior: func(mock redismock.ClientMock, refreshToken, userID string) {
				sessionKey := "session:" + refreshToken

				mock.ExpectDel(sessionKey).SetErr(errUnexpected)
			},
			expectedErr: errUnexpected,
		},
		{
			name:         "Unexpected error while deleting session key from list of user's sessions",
			refreshToken: "qGVFLRQw37TnSmG0LKFN",
			userID:       "1",
			mockBehavior: func(mock redismock.ClientMock, refreshToken, userID string) {
				sessionKey := "session:" + refreshToken
				userSessionsKey := fmt.Sprintf("user:%s:sessions", userID)

				mock.ExpectDel(sessionKey).SetVal(1)
				mock.ExpectLRem(userSessionsKey, 0, sessionKey).SetErr(errUnexpected)
			},
			expectedErr: errUnexpected,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			client, mock := redismock.NewClientMock()
			sessionRepo := NewSessionRedisRepository(client)

			if testCase.mockBehavior != nil {
				testCase.mockBehavior(mock, testCase.refreshToken, testCase.userID)
			}

			err := sessionRepo.Delete(context.Background(), testCase.refreshToken, testCase.userID)

			if testCase.expectedErr == nil {
				assert.NoError(t, err)
			}

			if testCase.expectedErr != nil {
				assert.Error(t, err)
				if testCase.expectedErr != errUnexpected {
					assert.ErrorIs(t, err, testCase.expectedErr)
				}
			}

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err, "There were unfulfilled expectations")
		})
	}
}

func TestSessionRedisRepository_DeleteAllByUserID(t *testing.T) {
	type mockBehavior func(mock redismock.ClientMock, userID string)

	logging.InitLogger(
		logging.LogConfig{
			LoggerKind: "mock",
		},
	)

	testTable := []struct {
		name         string
		userID       string
		mockBehavior mockBehavior
		expectedErr  error
	}{
		{
			name:   "Success",
			userID: "1",
			mockBehavior: func(mock redismock.ClientMock, userID string) {
				userSessionsKey := fmt.Sprintf("user:%s:sessions", userID)
				returnedKeys := []string{"session:qGVFLRQw37TnSmG0LKFN", "session:TnSmG0LKFNqGVFLRQw37"}

				mock.ExpectLRange(userSessionsKey, 0, -1).SetVal(returnedKeys)

				returnedKeys = append(returnedKeys, userSessionsKey)

				mock.ExpectDel(returnedKeys...).SetVal(int64(len(returnedKeys)))
			},
			expectedErr: nil,
		},
		{
			name:   "Unexpected error while range user's session keys",
			userID: "1",
			mockBehavior: func(mock redismock.ClientMock, userID string) {
				userSessionsKey := fmt.Sprintf("user:%s:sessions", userID)

				mock.ExpectLRange(userSessionsKey, 0, -1).SetErr(errUnexpected)
			},
			expectedErr: errUnexpected,
		},
		{
			name:   "Unexpected error while deleting session keys and key which aggregated them",
			userID: "1",
			mockBehavior: func(mock redismock.ClientMock, userID string) {
				userSessionsKey := fmt.Sprintf("user:%s:sessions", userID)
				returnedKeys := []string{"session:qGVFLRQw37TnSmG0LKFN", "session:TnSmG0LKFNqGVFLRQw37"}

				mock.ExpectLRange(userSessionsKey, 0, -1).SetVal(returnedKeys)

				returnedKeys = append(returnedKeys, userSessionsKey)

				mock.ExpectDel(returnedKeys...).SetErr(errUnexpected)
			},
			expectedErr: errUnexpected,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			client, mock := redismock.NewClientMock()
			sessionRepo := NewSessionRedisRepository(client)

			if testCase.mockBehavior != nil {
				testCase.mockBehavior(mock, testCase.userID)
			}

			err := sessionRepo.DeleteAllByUserID(context.Background(), testCase.userID)

			if testCase.expectedErr == nil {
				assert.NoError(t, err)
			}

			if testCase.expectedErr != nil {
				assert.Error(t, err)
				if testCase.expectedErr != errUnexpected {
					assert.ErrorIs(t, err, testCase.expectedErr)
				}
			}

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err, "There were unfulfilled expectations")
		})
	}
}
