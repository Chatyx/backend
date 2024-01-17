package http

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Chatyx/backend/pkg/auth"
	"github.com/Chatyx/backend/pkg/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	rtCookieName   = "refresh_token"
	rtCookieDomain = "localhost"
	rtCookiePath   = "/auth"
	rtCookieTTL    = 1 * time.Minute
)

const (
	defaultFingerprint  = "test1234"
	defaultRefreshToken = "qGVFLRQw37TnSmG0LKFN"
)

var (
	errUnexpected = errors.New("unexpected error")
)

func TestController_login(t *testing.T) {
	testCases := []struct {
		name                 string
		requestBody          string
		setFingerprint       bool
		mockBehavior         func(s *MockService)
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:           "Successful",
			requestBody:    `{"username":"root","password":"root1234"}`,
			setFingerprint: true,
			mockBehavior: func(s *MockService) {
				s.On("Login",
					mock.Anything,
					auth.Credentials{
						Username:    "root",
						Password:    "root1234",
						Fingerprint: defaultFingerprint,
					},
					mock.AnythingOfType("auth.MetaOption"),
				).Return(auth.TokenPair{
					AccessToken:  "{header}.{payload}.{sign}",
					RefreshToken: defaultRefreshToken,
				}, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"access_token":"{header}.{payload}.{sign}","refresh_token":"qGVFLRQw37TnSmG0LKFN"}`,
		},
		{
			name:           "Wrong credentials",
			requestBody:    `{"username":"root","password":"root1234"}`,
			setFingerprint: true,
			mockBehavior: func(s *MockService) {
				s.On("Login",
					mock.Anything,
					auth.Credentials{
						Username:    "root",
						Password:    "root1234",
						Fingerprint: defaultFingerprint,
					},
					mock.AnythingOfType("auth.MetaOption"),
				).Return(auth.TokenPair{}, auth.ErrWrongCredentials)
			},
			expectedStatusCode:   http.StatusUnauthorized,
			expectedResponseBody: `{"code":"AU0001","message":"login failed"}`,
		},
		{
			name:                 "Decode body error",
			requestBody:          `{"username":"root","password":"root1234`,
			setFingerprint:       true,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0002","message":"decode body error"}`,
		},
		{
			name:                 "Validation error: username and password are required",
			requestBody:          `{}`,
			setFingerprint:       true,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0003","message":"validation error","data":{"password":"failed on the 'required' tag","username":"failed on the 'required' tag"}}`,
		},
		{
			name:                 "Validation error: empty fingerprint",
			requestBody:          `{"username":"root","password":"root1234"}`,
			setFingerprint:       false,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0003","message":"validation error","data":{"X-Fingerprint":"failed on the 'required' tag"}}`,
		},
		{
			name:           "Internal server error",
			requestBody:    `{"username":"root","password":"root1234"}`,
			setFingerprint: true,
			mockBehavior: func(s *MockService) {
				s.On("Login",
					mock.Anything,
					auth.Credentials{
						Username:    "root",
						Password:    "root1234",
						Fingerprint: defaultFingerprint,
					},
					mock.AnythingOfType("auth.MetaOption"),
				).Return(auth.TokenPair{}, errUnexpected)
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"code":"CM0001","message":"internal server error"}`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			service := NewMockService(t)
			if testCase.mockBehavior != nil {
				testCase.mockBehavior(service)
			}

			cnt := NewController(
				service,
				validator.NewValidator(),
				WithPrefixPath("/"),
				WithRTCookieName(rtCookieName),
				WithRTCookieDomain(rtCookieDomain),
				WithRTCookieTTL(rtCookieTTL),
			)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, loginPath, strings.NewReader(testCase.requestBody))
			if testCase.setFingerprint {
				req.Header.Add(fingerprintHeaderKey, defaultFingerprint)
			}

			cnt.login(rec, req)
			resp := rec.Result()

			assert.Equal(t, testCase.expectedStatusCode, resp.StatusCode)

			respBody, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedResponseBody, string(respBody))

			if testCase.expectedStatusCode == http.StatusOK {
				checkValidRefreshTokenCookie(t, resp)
			}
		})
	}
}

func TestController_logout(t *testing.T) {
	testCases := []struct {
		name                 string
		requestBody          string
		rtCookie             *http.Cookie
		mockBehavior         func(s *MockService)
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:        "Successful with refresh token in body",
			requestBody: `{"refresh_token":"qGVFLRQw37TnSmG0LKFN"}`,
			mockBehavior: func(s *MockService) {
				s.On("Logout",
					mock.Anything,
					defaultRefreshToken,
				).Return(nil)
			},
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name: "Successful with refresh token in cookie",
			rtCookie: &http.Cookie{
				Name:  rtCookieName,
				Value: defaultRefreshToken,
			},
			mockBehavior: func(s *MockService) {
				s.On("Logout",
					mock.Anything,
					defaultRefreshToken,
				).Return(nil)
			},
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name:        "Invalid refresh token",
			requestBody: `{"refresh_token":"qGVFLRQw37TnSmG0LKFN"}`,
			mockBehavior: func(s *MockService) {
				s.On("Logout",
					mock.Anything,
					defaultRefreshToken,
				).Return(auth.ErrInvalidRefreshToken)
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"AU0002","message":"invalid refresh token"}`,
		},
		{
			name:                 "Decode body error",
			requestBody:          `{"refresh_token":"qGVFLRQw37TnSmG0LKFN`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0002","message":"decode body error"}`,
		},
		{
			name:                 "Validation error: refresh_token is required",
			requestBody:          `{}`,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0003","message":"validation error","data":{"refresh_token":"failed on the 'required' tag"}}`,
		},
		{
			name:        "Internal server error",
			requestBody: `{"refresh_token":"qGVFLRQw37TnSmG0LKFN"}`,
			mockBehavior: func(s *MockService) {
				s.On("Logout",
					mock.Anything,
					defaultRefreshToken,
				).Return(errUnexpected)
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"code":"CM0001","message":"internal server error"}`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			service := NewMockService(t)
			if testCase.mockBehavior != nil {
				testCase.mockBehavior(service)
			}

			cnt := NewController(
				service,
				validator.NewValidator(),
				WithPrefixPath("/"),
				WithRTCookieName(rtCookieName),
				WithRTCookieDomain(rtCookieDomain),
				WithRTCookieTTL(rtCookieTTL),
			)
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, refreshTokensPath, strings.NewReader(testCase.requestBody))
			if testCase.rtCookie != nil {
				req.AddCookie(testCase.rtCookie)
			}

			cnt.logout(rec, req)
			resp := rec.Result()

			assert.Equal(t, testCase.expectedStatusCode, resp.StatusCode)

			respBody, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedResponseBody, string(respBody))

			if testCase.expectedStatusCode == http.StatusNoContent {
				checkRemovedRefreshTokenCookie(t, resp)
			}
		})
	}
}

func TestController_refreshTokens(t *testing.T) {
	var (
		oldRefreshToken = "Bulto5iG1kxFmt8VGkPw"
		newRefreshToken = defaultRefreshToken
	)

	testCases := []struct {
		name                 string
		requestBody          string
		setFingerprint       bool
		rtCookie             *http.Cookie
		mockBehavior         func(s *MockService)
		expectedStatusCode   int
		expectedResponseBody string
	}{
		{
			name:           "Successful with refresh token in body",
			requestBody:    `{"refresh_token":"Bulto5iG1kxFmt8VGkPw"}`,
			setFingerprint: true,
			mockBehavior: func(s *MockService) {
				s.On("RefreshSession",
					mock.Anything,
					auth.RefreshSession{
						RefreshToken: oldRefreshToken,
						Fingerprint:  defaultFingerprint,
					},
					mock.AnythingOfType("auth.MetaOption"),
				).Return(auth.TokenPair{
					AccessToken:  "{header}.{payload}.{sign}",
					RefreshToken: newRefreshToken,
				}, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"access_token":"{header}.{payload}.{sign}","refresh_token":"qGVFLRQw37TnSmG0LKFN"}`,
		},
		{
			name: "Successful with refresh token in cookie",
			rtCookie: &http.Cookie{
				Name:  rtCookieName,
				Value: oldRefreshToken,
			},
			setFingerprint: true,
			mockBehavior: func(s *MockService) {
				s.On("RefreshSession",
					mock.Anything,
					auth.RefreshSession{
						RefreshToken: oldRefreshToken,
						Fingerprint:  defaultFingerprint,
					},
					mock.AnythingOfType("auth.MetaOption"),
				).Return(auth.TokenPair{
					AccessToken:  "{header}.{payload}.{sign}",
					RefreshToken: newRefreshToken,
				}, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"access_token":"{header}.{payload}.{sign}","refresh_token":"qGVFLRQw37TnSmG0LKFN"}`,
		},
		{
			name:           "Invalid refresh token",
			requestBody:    `{"refresh_token":"Bulto5iG1kxFmt8VGkPw"}`,
			setFingerprint: true,
			mockBehavior: func(s *MockService) {
				s.On("RefreshSession",
					mock.Anything,
					auth.RefreshSession{
						RefreshToken: oldRefreshToken,
						Fingerprint:  defaultFingerprint,
					},
					mock.AnythingOfType("auth.MetaOption"),
				).Return(auth.TokenPair{}, auth.ErrInvalidRefreshToken)
			},
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"AU0002","message":"invalid refresh token"}`,
		},
		{
			name:                 "Decode body error",
			requestBody:          `{"refresh_token":"Bulto5iG1kxFmt8VGkPw`,
			setFingerprint:       true,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0002","message":"decode body error"}`,
		},
		{
			name:                 "Validation error: refresh_token is required",
			requestBody:          `{}`,
			setFingerprint:       true,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0003","message":"validation error","data":{"refresh_token":"failed on the 'required' tag"}}`,
		},
		{
			name:                 "Validation error: empty fingerprint",
			requestBody:          `{"refresh_token":"Bulto5iG1kxFmt8VGkPw"}`,
			setFingerprint:       false,
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"code":"CM0003","message":"validation error","data":{"X-Fingerprint":"failed on the 'required' tag"}}`,
		},
		{
			name:           "Internal server error",
			requestBody:    `{"refresh_token":"Bulto5iG1kxFmt8VGkPw"}`,
			setFingerprint: true,
			mockBehavior: func(s *MockService) {
				s.On("RefreshSession",
					mock.Anything,
					auth.RefreshSession{
						RefreshToken: oldRefreshToken,
						Fingerprint:  defaultFingerprint,
					},
					mock.AnythingOfType("auth.MetaOption"),
				).Return(auth.TokenPair{}, errUnexpected)
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"code":"CM0001","message":"internal server error"}`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			service := NewMockService(t)
			if testCase.mockBehavior != nil {
				testCase.mockBehavior(service)
			}

			cnt := NewController(
				service,
				validator.NewValidator(),
				WithPrefixPath("/"),
				WithRTCookieName(rtCookieName),
				WithRTCookieDomain(rtCookieDomain),
				WithRTCookieTTL(rtCookieTTL),
			)
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, refreshTokensPath, strings.NewReader(testCase.requestBody))
			if testCase.setFingerprint {
				req.Header.Add(fingerprintHeaderKey, defaultFingerprint)
			}
			if testCase.rtCookie != nil {
				req.AddCookie(testCase.rtCookie)
			}

			cnt.refreshTokens(rec, req)
			resp := rec.Result()

			assert.Equal(t, testCase.expectedStatusCode, resp.StatusCode)

			respBody, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedResponseBody, string(respBody))

			if testCase.expectedStatusCode == http.StatusOK {
				checkValidRefreshTokenCookie(t, resp)
			}
		})
	}
}

func checkValidRefreshTokenCookie(t *testing.T, resp *http.Response) {
	t.Helper()

	var rtCookie *http.Cookie
	for _, cookie := range resp.Cookies() {
		if cookie.Name == rtCookieName {
			rtCookie = cookie
			break
		}
	}

	require.NotNilf(t, rtCookie, "Expected %s cookie, got nil", rtCookieName)
	assert.Equal(t, defaultRefreshToken, rtCookie.Value, "Wrong refresh token cookie value")
	assert.Equal(t, rtCookiePath, rtCookie.Path, "Wrong refresh token cookie path")
	assert.Equal(t, rtCookieDomain, rtCookie.Domain, "Wrong refresh token cookie domain")
	assert.True(t, rtCookie.HttpOnly, "Refresh cookie token must be http only")

	if rtCookie.MaxAge != 0 {
		assert.Equal(t, int(rtCookieTTL.Seconds()), rtCookie.MaxAge, "Wrong refresh token cookie max age")
	} else {
		ttl := time.Until(rtCookie.Expires)
		assert.GreaterOrEqual(t, ttl, time.Duration(0), "Wrong refresh token cookie expires at")
		assert.LessOrEqual(t, ttl, rtCookieTTL, "Wrong refresh token cookie expires at")
	}
}

func checkRemovedRefreshTokenCookie(t *testing.T, resp *http.Response) {
	t.Helper()

	var rtCookie *http.Cookie
	for _, cookie := range resp.Cookies() {
		if cookie.Name == rtCookieName {
			rtCookie = cookie
			break
		}
	}

	require.NotNilf(t, rtCookie, "Expected %s cookie, got nil", rtCookieName)
	assert.Equal(t, rtCookiePath, rtCookie.Path, "Wrong refresh token cookie path")
	assert.Equal(t, rtCookieDomain, rtCookie.Domain, "Wrong refresh token cookie domain")
	assert.True(t, rtCookie.HttpOnly, "Refresh cookie token must be http only")
	if rtCookie.MaxAge != 0 {
		assert.Equal(t, -1, rtCookie.MaxAge, "Wrong refresh token cookie max age")
	} else {
		assert.Less(t, rtCookie.Expires, time.Now(), "Wrong refresh token cookie expires at")
	}
}
