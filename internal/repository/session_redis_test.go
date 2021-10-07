// +build unit

package repository

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

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
			name:  "Unexpected error while unmarshalling",
			key:   "qGVFLRQw37TnSmG0LKFN",
			value: `{"UserID":"1","RefreshToken":"qGVFLRQw37TnSmG0LKFN"`,
			mockBehavior: func(mock redismock.ClientMock, key, value string) {
				mock.ExpectGet(key).SetVal(value)
			},
			strictCheckErrors: false,
			expectedErr:       errors.New("unmarshalling error"),
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

			if testCase.expectedErr == nil && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if testCase.strictCheckErrors && testCase.expectedErr != nil && testCase.expectedErr != err {
				t.Errorf("Wrong returned error. Expected error %v, got %v", testCase.expectedErr, err)
			}

			if !testCase.strictCheckErrors && testCase.expectedErr != nil && err == nil {
				t.Errorf("Expected error, got nil")
			}

			if !reflect.DeepEqual(testCase.expectedSession, session) {
				t.Errorf("Wrong found session. Expected %#v, got %#v", testCase.expectedSession, session)
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				t.Errorf("There were unfulfilled expectations: %v", err)
			}
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
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			client, mock := redismock.NewClientMock()
			sessionRepo := NewSessionRedisRepository(client)

			if testCase.mockBehavior != nil {
				testCase.mockBehavior(mock, testCase.key, testCase.session, refreshTokenTTL)
			}

			err := sessionRepo.Set(context.Background(), testCase.key, testCase.session, refreshTokenTTL)

			if testCase.expectedErr == nil && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if testCase.expectedErr != nil && testCase.expectedErr != err {
				t.Errorf("Wrong returned error. Expected error %v, got %v", testCase.expectedErr, err)
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				t.Errorf("There were unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestSessionRedisRepository_Delete(t *testing.T) {
	type mockBehavior func(mock redismock.ClientMock, key string)

	logging.InitLogger(
		logging.LogConfig{
			LoggerKind: "mock",
		},
	)

	testTable := []struct {
		name         string
		key          string
		mockBehavior mockBehavior
		expectedErr  error
	}{
		{
			name: "Success",
			key:  "qGVFLRQw37TnSmG0LKFN",
			mockBehavior: func(mock redismock.ClientMock, key string) {
				mock.ExpectDel(key).SetVal(1)
			},
			expectedErr: nil,
		},
		{
			name: "Session is not found",
			key:  "qGVFLRQw37TnSmG0LKFN",
			mockBehavior: func(mock redismock.ClientMock, key string) {
				mock.ExpectDel(key).SetVal(0)
			},
			expectedErr: domain.ErrSessionNotFound,
		},
		{
			name: "Unexpected error while deleting the session",
			key:  "qGVFLRQw37TnSmG0LKFN",
			mockBehavior: func(mock redismock.ClientMock, key string) {
				mock.ExpectDel(key).SetErr(errUnexpected)
			},
			expectedErr: errUnexpected,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			client, mock := redismock.NewClientMock()
			sessionRepo := NewSessionRedisRepository(client)

			if testCase.mockBehavior != nil {
				testCase.mockBehavior(mock, testCase.key)
			}

			err := sessionRepo.Delete(context.Background(), testCase.key)

			if testCase.expectedErr == nil && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if testCase.expectedErr != nil && testCase.expectedErr != err {
				t.Errorf("Wrong returned error. Expected error %v, got %v", testCase.expectedErr, err)
			}

			if err = mock.ExpectationsWereMet(); err != nil {
				t.Errorf("There were unfulfilled expectations: %v", err)
			}
		})
	}
}
