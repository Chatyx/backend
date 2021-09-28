package http

import (
	"context"
	"errors"
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
)

func TestAuthHandler_signIn(t *testing.T) {
	type mockBehavior func(as *mockservice.MockAuthService, ctx context.Context, dto domain.SignInDTO, jwtPair domain.JWTPair)

	const (
		domainName = "localhost"

		authURLPathPrefix = "/api/auth"
		signInURLPath     = authURLPathPrefix + "/sign-in"

		refreshCookieName = "refresh_token"
		refreshTokenTTL   = 15 * 24 * time.Hour // 15 days
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
			expectedResponseBody:   `{"message":"invalid json body"}`,
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

	logging.InitLogger(logging.LogConfig{
		LoggerKind: "mock",
	})

	validate, err := validator.New()
	if err != nil {
		t.Errorf("Unexpected error while creating validator: %v", err)
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			as := mockservice.NewMockAuthService(c)
			ah := newAuthHandler(as, validate, domainName, refreshTokenTTL)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, signInURLPath, strings.NewReader(testCase.requestBody))

			if testCase.fingerPrintHeaderKey != "" {
				req.Header.Add(testCase.fingerPrintHeaderKey, testCase.fingerPrintHeaderValue)
			}

			if testCase.mockBehavior != nil {
				testCase.mockBehavior(as, req.Context(), testCase.signInDTO, testCase.jwtPair)
			}

			ah.signIn(rec, req, httprouter.Params{})

			resp := rec.Result()
			expectedExpiresAt := time.Now().Add(refreshTokenTTL)

			checkResponseResult(t, resp, testCase.expectedStatusCode, testCase.expectedResponseBody)

			if testCase.expectedStatusCode != http.StatusOK {
				return
			}

			var refreshCookie *http.Cookie
			for _, cookie := range resp.Cookies() {
				if cookie.Name == refreshCookieName {
					refreshCookie = cookie
					break
				}
			}

			if refreshCookie == nil {
				t.Errorf("Expected %s cookie, got nil", refreshCookieName)
				return
			}

			if refreshCookie.Value != testCase.jwtPair.RefreshToken {
				t.Errorf(
					"Wrong %s cookie value. Expected %s, got %s",
					refreshCookieName, testCase.jwtPair.RefreshToken, refreshCookie.Value,
				)
			}

			if !strings.HasPrefix(refreshCookie.Path, authURLPathPrefix) {
				t.Errorf(
					"Wrong %s cookie path. Expected that cookie path has prefix %s, but got %s",
					refreshCookieName, authURLPathPrefix, refreshCookie.Path,
				)
			}

			if refreshCookie.Domain != domainName {
				t.Errorf(
					"Wrong %s cookie domain. Expected %s, got %s",
					refreshCookieName, domainName, refreshCookie.Domain,
				)
			}

			if refreshCookie.MaxAge != 0 {
				if expectedMaxAge := int(refreshTokenTTL.Seconds()); refreshCookie.MaxAge != expectedMaxAge {
					t.Errorf(
						"Wrong %s cookie max age. Expected %d, got %d",
						refreshCookieName, expectedMaxAge, refreshCookie.MaxAge,
					)
				}
			} else if cookieExpires := refreshCookie.Expires.Local(); expectedExpiresAt.Sub(cookieExpires) > time.Second {
				t.Errorf(
					"Wrong %s cookie expires at. Expected difference less then 1 second, got %s",
					refreshCookieName, expectedExpiresAt.Sub(cookieExpires),
				)
			}

			if !refreshCookie.HttpOnly {
				t.Errorf("Refresh cookie must be http only")
			}
		})
	}
}
