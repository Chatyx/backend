// +build unit

package repository

import (
	"context"
	"encoding/json"
	"errors"
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
	type mockBehavior func(mock redismock.ClientMock, key, value string)

	logging.InitLogger(
		logging.LogConfig{
			LoggerKind: "mock",
		},
	)

	testTable := []struct {
		name              string
		key               string
		value             string
		mockBehavior      mockBehavior
		strictCheckErrors bool
		expectedSession   domain.Session
		expectedErr       error
	}{
		{
			name:  "Success",
			key:   "qGVFLRQw37TnSmG0LKFN",
			value: `{"UserID":"1","RefreshToken":"qGVFLRQw37TnSmG0LKFN","Fingerprint":"5dc49b7a-6153-4eae-9c0f-297655c45f08","CreatedAt":"2021-10-02T14:47:12.00+03:00","ExpiresAt":"2021-11-01T14:47:12.00+03:00"}`,
			mockBehavior: func(mock redismock.ClientMock, key, value string) {
				mock.ExpectGet(key).SetVal(value)
			},
			expectedSession: exampleSession,
			expectedErr:     nil,
		},
		{
			name: "Not found",
			key:  "qGVFLRQw37TnSmG0LKFN",
			mockBehavior: func(mock redismock.ClientMock, key, value string) {
				mock.ExpectGet(key).RedisNil()
			},
			strictCheckErrors: true,
			expectedErr:       domain.ErrSessionNotFound,
		},
		{
			name: "Unexpected error while getting session",
			key:  "qGVFLRQw37TnSmG0LKFN",
			mockBehavior: func(mock redismock.ClientMock, key, value string) {
				mock.ExpectGet(key).SetErr(errUnexpected)
			},
			strictCheckErrors: true,
			expectedErr:       errUnexpected,
		},
		{
			name:  "Unexpected error while unmarshaling",
			key:   "qGVFLRQw37TnSmG0LKFN",
			value: `{"UserID":"1","RefreshToken":"qGVFLRQw37TnSmG0LKFN"`,
			mockBehavior: func(mock redismock.ClientMock, key, value string) {
				mock.ExpectGet(key).SetVal(value)
			},
			strictCheckErrors: false,
			expectedErr:       errors.New("unmarshaling error"),
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			client, mock := redismock.NewClientMock()
			sessionRepo := NewSessionRedisRepository(client)

			if testCase.mockBehavior != nil {
				testCase.mockBehavior(mock, testCase.key, testCase.value)
			}

			session, err := sessionRepo.Get(context.Background(), testCase.key)

			if testCase.expectedErr == nil {
				assert.NoError(t, err)
			}

			if testCase.expectedErr != nil {
				assert.Error(t, err)
				if testCase.strictCheckErrors {
					assert.EqualError(t, err, testCase.expectedErr.Error())
				}
			}

			assert.Equal(t, testCase.expectedSession, session)

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err, "There were unfulfilled expectations")
		})
	}
}

func TestSessionRedisRepository_Set(t *testing.T) {
	type mockBehavior func(mock redismock.ClientMock, key string, session domain.Session, ttl time.Duration)

	logging.InitLogger(
		logging.LogConfig{
			LoggerKind: "mock",
		},
	)

	testTable := []struct {
		name         string
		key          string
		session      domain.Session
		mockBehavior mockBehavior
		expectedErr  error
	}{
		{
			name:    "Success",
			key:     "qGVFLRQw37TnSmG0LKFN",
			session: exampleSession,
			mockBehavior: func(mock redismock.ClientMock, key string, session domain.Session, ttl time.Duration) {
				payload, _ := json.Marshal(session)
				mock.ExpectSet(key, payload, ttl).SetVal("OK")
				mock.ExpectRPush(session.UserID, key).SetVal(1)
			},
			expectedErr: nil,
		},
		{
			name:    "Unexpected error while setting session",
			key:     "qGVFLRQw37TnSmG0LKFN",
			session: exampleSession,
			mockBehavior: func(mock redismock.ClientMock, key string, session domain.Session, ttl time.Duration) {
				payload, _ := json.Marshal(session)
				mock.ExpectSet(key, payload, ttl).SetErr(errUnexpected)
			},
			expectedErr: errUnexpected,
		},
		{
			name:    "Unexpected error while pushing session to the list of user's sessions",
			key:     "qGVFLRQw37TnSmG0LKFN",
			session: exampleSession,
			mockBehavior: func(mock redismock.ClientMock, key string, session domain.Session, ttl time.Duration) {
				payload, _ := json.Marshal(session)
				mock.ExpectSet(key, payload, ttl).SetVal("OK")
				mock.ExpectRPush(session.UserID, key).SetErr(errUnexpected)
			},
			expectedErr: errUnexpected,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			client, mock := redismock.NewClientMock()
			sessionRepo := NewSessionRedisRepository(client)

			if testCase.mockBehavior != nil {
				testCase.mockBehavior(mock, testCase.key, testCase.session, refreshTokenTTL)
			}

			err := sessionRepo.Set(context.Background(), testCase.key, testCase.session, refreshTokenTTL)

			if testCase.expectedErr == nil {
				assert.NoError(t, err)
			}

			if testCase.expectedErr != nil {
				assert.EqualError(t, err, testCase.expectedErr.Error())
			}

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err, "There were unfulfilled expectations")
		})
	}
}

func TestSessionRedisRepository_Delete(t *testing.T) {
	type mockBehavior func(mock redismock.ClientMock, key, userID string)

	logging.InitLogger(
		logging.LogConfig{
			LoggerKind: "mock",
		},
	)

	testTable := []struct {
		name         string
		key          string
		userID       string
		mockBehavior mockBehavior
		expectedErr  error
	}{
		{
			name:   "Success",
			key:    "qGVFLRQw37TnSmG0LKFN",
			userID: "1",
			mockBehavior: func(mock redismock.ClientMock, key, userID string) {
				mock.ExpectDel(key).SetVal(1)
				mock.ExpectLRem(userID, 0, key).SetVal(1)
			},
			expectedErr: nil,
		},
		{
			name:   "Session is not found",
			key:    "qGVFLRQw37TnSmG0LKFN",
			userID: "1",
			mockBehavior: func(mock redismock.ClientMock, key, userID string) {
				mock.ExpectDel(key).SetVal(0)
			},
			expectedErr: domain.ErrSessionNotFound,
		},
		{
			name:   "Unexpected error while deleting session by key",
			key:    "qGVFLRQw37TnSmG0LKFN",
			userID: "1",
			mockBehavior: func(mock redismock.ClientMock, key, userID string) {
				mock.ExpectDel(key).SetErr(errUnexpected)
			},
			expectedErr: errUnexpected,
		},
		{
			name:   "Unexpected error while deleting session key from list of user's sessions",
			key:    "qGVFLRQw37TnSmG0LKFN",
			userID: "1",
			mockBehavior: func(mock redismock.ClientMock, key, userID string) {
				mock.ExpectDel(key).SetVal(1)
				mock.ExpectLRem(userID, 0, key).SetErr(errUnexpected)
			},
			expectedErr: errUnexpected,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			client, mock := redismock.NewClientMock()
			sessionRepo := NewSessionRedisRepository(client)

			if testCase.mockBehavior != nil {
				testCase.mockBehavior(mock, testCase.key, testCase.userID)
			}

			err := sessionRepo.Delete(context.Background(), testCase.key, testCase.userID)

			if testCase.expectedErr == nil {
				assert.NoError(t, err)
			}

			if testCase.expectedErr != nil {
				assert.EqualError(t, err, testCase.expectedErr.Error())
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
				returnedKeys := []string{"qGVFLRQw37TnSmG0LKFN", "TnSmG0LKFNqGVFLRQw37"}
				mock.ExpectLRange(userID, 0, -1).SetVal(returnedKeys)

				returnedKeys = append(returnedKeys, userID)

				mock.ExpectDel(returnedKeys...).SetVal(int64(len(returnedKeys)))
			},
			expectedErr: nil,
		},
		{
			name:   "Unexpected error while range user's session keys",
			userID: "1",
			mockBehavior: func(mock redismock.ClientMock, userID string) {
				mock.ExpectLRange(userID, 0, -1).SetErr(errUnexpected)
			},
			expectedErr: errUnexpected,
		},
		{
			name:   "Unexpected error while deleting session keys and key which aggregated them",
			userID: "1",
			mockBehavior: func(mock redismock.ClientMock, userID string) {
				returnedKeys := []string{"qGVFLRQw37TnSmG0LKFN", "TnSmG0LKFNqGVFLRQw37"}
				mock.ExpectLRange(userID, 0, -1).SetVal(returnedKeys)

				returnedKeys = append(returnedKeys, userID)

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
				assert.EqualError(t, err, testCase.expectedErr.Error())
			}

			err = mock.ExpectationsWereMet()
			assert.NoError(t, err, "There were unfulfilled expectations")
		})
	}
}
