package auth

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/Chatyx/backend/pkg/token"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	issuer = "test"
)

var (
	errUnexpected = errors.New("unexpected error")
)

func TestService_Login(t *testing.T) {
	correctCred := Credentials{
		Username:    "test",
		Password:    "test",
		Fingerprint: "12345",
	}

	testCases := []struct {
		name          string
		cred          Credentials
		mockBehavior  func(m *MockSessionStorage)
		expectedError error
	}{
		{
			name: "Successful",
			cred: correctCred,
			mockBehavior: func(m *MockSessionStorage) {
				m.On("Set",
					context.Background(),
					mock.AnythingOfType("Session"),
				).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "Wrong credentials",
			cred: Credentials{
				Username:    "test",
				Password:    uuid.New().String(),
				Fingerprint: "12345",
			},
			expectedError: ErrWrongCredentials,
		},
		{
			name: "Check password error",
			cred: Credentials{
				Username:    "error",
				Password:    "error",
				Fingerprint: "12345",
			},
			expectedError: errUnexpected,
		},
		{
			name: "Set session error",
			cred: correctCred,
			mockBehavior: func(m *MockSessionStorage) {
				m.On("Set",
					context.Background(),
					mock.AnythingOfType("Session"),
				).Return(errUnexpected)
			},
			expectedError: errUnexpected,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			storage := NewMockSessionStorage(t)
			if testCase.mockBehavior != nil {
				testCase.mockBehavior(storage)
			}

			signedKey := uuid.New().String()
			service := NewService(
				storage,
				WithIssuer(issuer),
				WithAccessTokenTTL(1*time.Minute),
				WithRefreshTokenTTL(1*time.Minute),
				WithSignedKey([]byte(signedKey)),
				WithEnrichClaims(func(claims token.Claims) {
					claims[issuer] = issuer
				}),
				WithCheckPassword(func(user, password string) (userID string, ok bool, err error) {
					if user == "error" || password == "error" {
						return "", false, errUnexpected
					}
					if user == "test" && password == "test" {
						return "0", true, nil
					}
					return "", false, nil
				}),
			)

			pair, err := service.Login(
				context.Background(),
				testCase.cred,
				WithIP(net.ParseIP("127.0.0.1")),
			)
			if testCase.expectedError == nil {
				require.NoError(t, err)
				assert.NotEmpty(t, pair.RefreshToken)
				checkAccessToken(t, pair.AccessToken, signedKey)
			} else {
				assert.ErrorIs(t, err, testCase.expectedError)
			}
		})
	}
}

func TestService_Logout(t *testing.T) {
	refreshToken := uuid.New().String()

	testCases := []struct {
		name          string
		mockBehavior  func(m *MockSessionStorage)
		expectedError error
	}{
		{
			name: "Successful",
			mockBehavior: func(m *MockSessionStorage) {
				m.On("GetWithDelete",
					context.Background(),
					refreshToken,
				).Return(Session{}, nil)
			},
			expectedError: nil,
		},
		{
			name: "Session is not found",
			mockBehavior: func(m *MockSessionStorage) {
				m.On("GetWithDelete",
					context.Background(),
					refreshToken,
				).Return(Session{}, ErrSessionNotFound)
			},
			expectedError: ErrInvalidRefreshToken,
		},
		{
			name: "Get with delete session error",
			mockBehavior: func(m *MockSessionStorage) {
				m.On("GetWithDelete",
					context.Background(),
					refreshToken,
				).Return(Session{}, errUnexpected)
			},
			expectedError: errUnexpected,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			storage := NewMockSessionStorage(t)
			if testCase.mockBehavior != nil {
				testCase.mockBehavior(storage)
			}

			service := NewService(storage)
			err := service.Logout(context.Background(), refreshToken)

			if testCase.expectedError == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, testCase.expectedError)
			}
		})
	}
}

func TestService_RefreshSession(t *testing.T) {
	var (
		refreshToken = uuid.New().String()
		fingerprint  = uuid.New().String()
	)

	testCases := []struct {
		name          string
		mockBehavior  func(m *MockSessionStorage)
		expectedError error
	}{
		{
			name: "Successful",
			mockBehavior: func(m *MockSessionStorage) {
				m.On("GetWithDelete",
					context.Background(),
					refreshToken,
				).Return(Session{
					RefreshToken: refreshToken,
					Fingerprint:  fingerprint,
					ExpiresAt:    time.Now().Add(1 * time.Minute),
					CreatedAt:    time.Now(),
				}, nil)

				m.On("Set",
					context.Background(),
					mock.AnythingOfType("Session"),
				).Return(nil)
			},
			expectedError: nil,
		},
		{
			name: "Session is not found",
			mockBehavior: func(m *MockSessionStorage) {
				m.On("GetWithDelete",
					context.Background(),
					refreshToken,
				).Return(Session{}, ErrSessionNotFound)
			},
			expectedError: ErrInvalidRefreshToken,
		},
		{
			name: "Session is expired",
			mockBehavior: func(m *MockSessionStorage) {
				m.On("GetWithDelete",
					context.Background(),
					refreshToken,
				).Return(Session{
					RefreshToken: refreshToken,
					Fingerprint:  fingerprint,
					ExpiresAt:    time.Now(),
					CreatedAt:    time.Now(),
				}, nil)
			},
			expectedError: ErrInvalidRefreshToken,
		},
		{
			name: "Fingerprints don't match",
			mockBehavior: func(m *MockSessionStorage) {
				m.On("GetWithDelete",
					context.Background(),
					refreshToken,
				).Return(Session{
					RefreshToken: refreshToken,
					Fingerprint:  uuid.New().String(),
					ExpiresAt:    time.Now().Add(1 * time.Minute),
					CreatedAt:    time.Now(),
				}, nil)
			},
			expectedError: ErrInvalidRefreshToken,
		},
		{
			name: "Get with delete session error",
			mockBehavior: func(m *MockSessionStorage) {
				m.On("GetWithDelete",
					context.Background(),
					refreshToken,
				).Return(Session{}, errUnexpected)
			},
			expectedError: errUnexpected,
		},
		{
			name: "Set session error",
			mockBehavior: func(m *MockSessionStorage) {
				m.On("GetWithDelete",
					context.Background(),
					refreshToken,
				).Return(Session{
					RefreshToken: refreshToken,
					Fingerprint:  fingerprint,
					ExpiresAt:    time.Now().Add(1 * time.Minute),
					CreatedAt:    time.Now(),
				}, nil)

				m.On("Set",
					context.Background(),
					mock.AnythingOfType("Session"),
				).Return(errUnexpected)
			},
			expectedError: errUnexpected,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			storage := NewMockSessionStorage(t)
			if testCase.mockBehavior != nil {
				testCase.mockBehavior(storage)
			}

			signedKey := uuid.New().String()
			service := NewService(
				storage,
				WithIssuer(issuer),
				WithAccessTokenTTL(1*time.Minute),
				WithRefreshTokenTTL(1*time.Minute),
				WithSignedKey([]byte(signedKey)),
				WithEnrichClaims(func(claims token.Claims) {
					claims[issuer] = issuer
				}),
			)

			pair, err := service.RefreshSession(
				context.Background(),
				RefreshSession{
					RefreshToken: refreshToken,
					Fingerprint:  fingerprint,
				},
				WithIP(net.ParseIP("127.0.0.1")),
			)
			if testCase.expectedError == nil {
				require.NoError(t, err)
				assert.NotEmpty(t, pair.RefreshToken)
				assert.NotEqual(t, refreshToken, pair.RefreshToken)

				checkAccessToken(t, pair.AccessToken, signedKey)
			} else {
				assert.ErrorIs(t, err, testCase.expectedError)
			}
		})
	}
}

func checkAccessToken(t *testing.T, tokenStr, signedKey string) {
	t.Helper()

	tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		return []byte(signedKey), nil
	})
	require.NoError(t, err)
	assert.True(t, tok.Valid, "Access token is invalid")

	iss, err := tok.Claims.GetIssuer()
	assert.NoError(t, err)
	assert.Equal(t, issuer, iss)

	mapClaims, ok := tok.Claims.(jwt.MapClaims)
	assert.True(t, ok)
	assert.Equal(t, issuer, mapClaims[issuer])
}
