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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	domainName      = "localhost"
	refreshTokenTTL = 15 * 24 * time.Hour // 15 days
)

func TestAuthHandler_signIn(t *testing.T) {
	type mockBehavior func(as *mockservice.MockAuthService, ctx context.Context, dto domain.SignInDTO, jwtPair domain.JWTPair)

	logging.InitLogger(
		logging.LogConfig{
			LoggerKind: "mock",
		},
	)

	testTable := []struct {
		name                   string
		requestBody            string
		fingerPrintHeaderKey   string
		fingerPrintHeaderValue string
		signInDTO              domain.SignInDTO
		jwtPair                domain.JWTPair
		mockBehavior           mockBehavior
		expectedStatusCode     int
		expectedResponseBody   string
	}{
		{
			name:                   "Success",
			requestBody:            `{"username":"john1967","password":"qwerty12345"}`,
			fingerPrintHeaderKey:   "X-Fingerprint",
			fingerPrintHeaderValue: "5dc49b7a-6153-4eae-9c0f-297655c45f08",
			signInDTO: domain.SignInDTO{
				Username:    "john1967",
				Password:    "qwerty12345",
				Fingerprint: "5dc49b7a-6153-4eae-9c0f-297655c45f08",
			},
			jwtPair: domain.JWTPair{
				AccessToken:  "header.payload.sign",
				RefreshToken: "qGVFLRQw37TnSmG0LKFN",
			},
			mockBehavior: func(as *mockservice.MockAuthService, ctx context.Context, dto domain.SignInDTO, jwtPair domain.JWTPair) {
				as.EXPECT().SignIn(ctx, dto).Return(jwtPair, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"access_token":"header.payload.sign","refresh_token":"qGVFLRQw37TnSmG0LKFN"}`,
		},
		{
			name:                   "Wrong credentials",
			requestBody:            `{"username":"john1967","password":"cash91822849572"}`,
			fingerPrintHeaderKey:   "X-Fingerprint",
			fingerPrintHeaderValue: "5dc49b7a-6153-4eae-9c0f-297655c45f08",
			signInDTO: domain.SignInDTO{
				Username:    "john1967",
				Password:    "cash91822849572",
				Fingerprint: "5dc49b7a-6153-4eae-9c0f-297655c45f08",
			},
			mockBehavior: func(as *mockservice.MockAuthService, ctx context.Context, dto domain.SignInDTO, jwtPair domain.JWTPair) {
				as.EXPECT().SignIn(ctx, dto).Return(jwtPair, domain.ErrWrongCredentials)
			},
			expectedStatusCode:   http.StatusUnauthorized,
			expectedResponseBody: `{"message":"wrong username or password"}`,
		},
		{
			name:                   "Invalid JSON Body",
			requestBody:            `{"username":"john1967","password":"qwerty12345"`,
			fingerPrintHeaderKey:   "X-Fingerprint",
			fingerPrintHeaderValue: "5dc49b7a-6153-4eae-9c0f-297655c45f08",
			expectedStatusCode:     http.StatusBadRequest,
			expectedResponseBody:   `{"message":"invalid body to decode"}`,
		},
		{
			name:                   "Empty body",
			requestBody:            `{}`,
			fingerPrintHeaderKey:   "X-Fingerprint",
			fingerPrintHeaderValue: "5dc49b7a-6153-4eae-9c0f-297655c45f08",
			expectedStatusCode:     http.StatusBadRequest,
			expectedResponseBody:   `{"message":"validation error","fields":{"password":"field validation for 'password' failed on the 'required' tag","username":"field validation for 'username' failed on the 'required' tag"}}`,
		},
		{
			name:                 "Empty fingerprint header",
			requestBody:          `{"username":"john1967","password":"qwerty12345"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"X-Fingerprint header is required"}`,
		},
		{
			name:                   "Wrong fingerprint header key",
			requestBody:            `{"username":"john1967","password":"qwerty12345"}`,
			fingerPrintHeaderKey:   "Fingerprint",
			fingerPrintHeaderValue: "5dc49b7a-6153-4eae-9c0f-297655c45f08",
			expectedStatusCode:     http.StatusBadRequest,
			expectedResponseBody:   `{"message":"X-Fingerprint header is required"}`,
		},
		{
			name:                   "Empty fingerprint header value",
			requestBody:            `{"username":"john1967","password":"qwerty12345"}`,
			fingerPrintHeaderKey:   "X-Fingerprint",
			fingerPrintHeaderValue: "",
			expectedStatusCode:     http.StatusBadRequest,
			expectedResponseBody:   `{"message":"X-Fingerprint header is required"}`,
		},
		{
			name:                   "Unexpected error",
			requestBody:            `{"username":"john1967","password":"qwerty12345"}`,
			fingerPrintHeaderKey:   "X-Fingerprint",
			fingerPrintHeaderValue: "5dc49b7a-6153-4eae-9c0f-297655c45f08",
			signInDTO: domain.SignInDTO{
				Username:    "john1967",
				Password:    "qwerty12345",
				Fingerprint: "5dc49b7a-6153-4eae-9c0f-297655c45f08",
			},
			mockBehavior: func(as *mockservice.MockAuthService, ctx context.Context, dto domain.SignInDTO, jwtPair domain.JWTPair) {
				as.EXPECT().SignIn(ctx, dto).Return(domain.JWTPair{}, errors.New("unexpected error"))
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

			as := mockservice.NewMockAuthService(c)
			ah := newAuthHandler(as, validate, domainName, refreshTokenTTL)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, signInURI, strings.NewReader(testCase.requestBody))

			if testCase.fingerPrintHeaderKey != "" {
				req.Header.Add(testCase.fingerPrintHeaderKey, testCase.fingerPrintHeaderValue)
			}

			if testCase.mockBehavior != nil {
				testCase.mockBehavior(as, req.Context(), testCase.signInDTO, testCase.jwtPair)
			}

			ah.signIn(rec, req)

			resp := rec.Result()

			respBodyPayload, err := ioutil.ReadAll(resp.Body)
			assert.NoError(t, err, "Unexpected error while reading response body")

			assert.Equal(t, testCase.expectedStatusCode, resp.StatusCode)
			assert.Equal(t, testCase.expectedResponseBody, string(respBodyPayload))

			if testCase.expectedStatusCode != http.StatusOK {
				return
			}

			checkRefreshCookie(t, resp, testCase.jwtPair.RefreshToken)
		})
	}
}

func TestAuthHandler_refresh(t *testing.T) {
	type mockBehavior func(as *mockservice.MockAuthService, ctx context.Context, dto domain.RefreshSessionDTO, jwtPair domain.JWTPair)

	logging.InitLogger(
		logging.LogConfig{
			LoggerKind: "mock",
		},
	)

	testTable := []struct {
		name                   string
		requestBody            string
		requestCookie          *http.Cookie
		fingerPrintHeaderKey   string
		fingerPrintHeaderValue string
		refreshSessionDTO      domain.RefreshSessionDTO
		jwtPair                domain.JWTPair
		mockBehavior           mockBehavior
		expectedStatusCode     int
		expectedResponseBody   string
	}{
		{
			name: "Success",
			requestCookie: &http.Cookie{
				Name:  "refresh_token",
				Value: "qGVFLRQw37TnSmG0LKFN",
			},
			fingerPrintHeaderKey:   "X-Fingerprint",
			fingerPrintHeaderValue: "5dc49b7a-6153-4eae-9c0f-297655c45f08",
			refreshSessionDTO: domain.RefreshSessionDTO{
				RefreshToken: "qGVFLRQw37TnSmG0LKFN",
				Fingerprint:  "5dc49b7a-6153-4eae-9c0f-297655c45f08",
			},
			jwtPair: domain.JWTPair{
				AccessToken:  "header.payload.sign",
				RefreshToken: "Bulto5iG1kxFmt8VGkPw",
			},
			mockBehavior: func(as *mockservice.MockAuthService, ctx context.Context, dto domain.RefreshSessionDTO, jwtPair domain.JWTPair) {
				as.EXPECT().Refresh(ctx, dto).Return(jwtPair, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"access_token":"header.payload.sign","refresh_token":"Bulto5iG1kxFmt8VGkPw"}`,
		},
		{
			name:                   "Invalid refresh token",
			requestBody:            `{"refresh_token":"qGVFLRQw37TnSmG0LKFN"}`,
			fingerPrintHeaderKey:   "X-Fingerprint",
			fingerPrintHeaderValue: "5dc49b7a-6153-4eae-9c0f-297655c45f08",
			refreshSessionDTO: domain.RefreshSessionDTO{
				RefreshToken: "qGVFLRQw37TnSmG0LKFN",
				Fingerprint:  "5dc49b7a-6153-4eae-9c0f-297655c45f08",
			},
			mockBehavior: func(as *mockservice.MockAuthService, ctx context.Context, dto domain.RefreshSessionDTO, jwtPair domain.JWTPair) {
				as.EXPECT().Refresh(ctx, dto).Return(domain.JWTPair{}, domain.ErrInvalidRefreshToken)
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"invalid refresh token"}`,
		},
		{
			name:                   "Invalid JSON body",
			requestBody:            `{"refresh_token":"`,
			fingerPrintHeaderKey:   "X-Fingerprint",
			fingerPrintHeaderValue: "5dc49b7a-6153-4eae-9c0f-297655c45f08",
			expectedStatusCode:     http.StatusBadRequest,
			expectedResponseBody:   `{"message":"invalid body to decode"}`,
		},
		{
			name:                   "Empty body",
			requestBody:            `{}`,
			fingerPrintHeaderKey:   "X-Fingerprint",
			fingerPrintHeaderValue: "5dc49b7a-6153-4eae-9c0f-297655c45f08",
			expectedStatusCode:     http.StatusBadRequest,
			expectedResponseBody:   `{"message":"validation error","fields":{"refresh_token":"field validation for 'refresh_token' failed on the 'required' tag"}}`,
		},
		{
			name:                 "Empty fingerprint header",
			requestBody:          `{"refresh_token":"qGVFLRQw37TnSmG0LKFN"}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"X-Fingerprint header is required"}`,
		},
		{
			name:                   "Wrong fingerprint header key",
			requestBody:            `{"refresh_token":"qGVFLRQw37TnSmG0LKFN"}`,
			fingerPrintHeaderKey:   "Fingerprint",
			fingerPrintHeaderValue: "5dc49b7a-6153-4eae-9c0f-297655c45f08",
			expectedStatusCode:     http.StatusBadRequest,
			expectedResponseBody:   `{"message":"X-Fingerprint header is required"}`,
		},
		{
			name:                   "Empty fingerprint header value",
			requestBody:            `{"refresh_token":"qGVFLRQw37TnSmG0LKFN"}`,
			fingerPrintHeaderKey:   "X-Fingerprint",
			fingerPrintHeaderValue: "",
			expectedStatusCode:     http.StatusBadRequest,
			expectedResponseBody:   `{"message":"X-Fingerprint header is required"}`,
		},
		{
			name: "Unexpected error",
			requestCookie: &http.Cookie{
				Name:  "refresh_token",
				Value: "qGVFLRQw37TnSmG0LKFN",
			},
			fingerPrintHeaderKey:   "X-Fingerprint",
			fingerPrintHeaderValue: "5dc49b7a-6153-4eae-9c0f-297655c45f08",
			refreshSessionDTO: domain.RefreshSessionDTO{
				RefreshToken: "qGVFLRQw37TnSmG0LKFN",
				Fingerprint:  "5dc49b7a-6153-4eae-9c0f-297655c45f08",
			},
			mockBehavior: func(as *mockservice.MockAuthService, ctx context.Context, dto domain.RefreshSessionDTO, jwtPair domain.JWTPair) {
				as.EXPECT().Refresh(ctx, dto).Return(domain.JWTPair{}, errors.New("unexpected error"))
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

			as := mockservice.NewMockAuthService(c)
			ah := newAuthHandler(as, validate, domainName, refreshTokenTTL)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, refreshURI, strings.NewReader(testCase.requestBody))

			if testCase.requestCookie != nil {
				req.AddCookie(testCase.requestCookie)
			}

			if testCase.fingerPrintHeaderKey != "" {
				req.Header.Add(testCase.fingerPrintHeaderKey, testCase.fingerPrintHeaderValue)
			}

			if testCase.mockBehavior != nil {
				testCase.mockBehavior(as, req.Context(), testCase.refreshSessionDTO, testCase.jwtPair)
			}

			ah.refresh(rec, req)

			resp := rec.Result()

			respBodyPayload, err := ioutil.ReadAll(resp.Body)
			assert.NoError(t, err, "Unexpected error while reading response body")

			assert.Equal(t, testCase.expectedStatusCode, resp.StatusCode)
			assert.Equal(t, testCase.expectedResponseBody, string(respBodyPayload))

			if testCase.expectedStatusCode != http.StatusOK {
				return
			}

			checkRefreshCookie(t, resp, testCase.jwtPair.RefreshToken)
		})
	}
}

func checkRefreshCookie(t *testing.T, resp *http.Response, expectedRefreshToken string) {
	t.Helper()

	expectedExpiresAt := time.Now().Add(refreshTokenTTL)

	var refreshCookie *http.Cookie

	for _, cookie := range resp.Cookies() {
		if cookie.Name == refreshCookieName {
			refreshCookie = cookie
			break
		}
	}

	assert.NotNilf(t, refreshCookie, "Expected %s cookie, got nil", refreshCookieName)
	assert.Equal(t, expectedRefreshToken, refreshCookie.Value, "Wrong refresh cookie value")
	assert.Equal(t, refreshURI, refreshCookie.Path, "Wrong refresh cookie path")
	assert.Equal(t, domainName, refreshCookie.Domain, "Wrong refresh cookie domain")
	assert.True(t, refreshCookie.HttpOnly, "Refresh cookie must be http only")

	if refreshCookie.MaxAge != 0 {
		assert.Equal(t, int(refreshTokenTTL.Seconds()), refreshCookie.MaxAge, "Wrong refresh cookie max age")
	} else {
		ttl := expectedExpiresAt.Sub(refreshCookie.Expires.Local())

		assert.GreaterOrEqual(t, ttl, time.Duration(0), "Wrong refresh cookie expires at")
		assert.LessOrEqual(t, ttl, time.Second, "Wrong refresh cookie expires at")
	}
}
