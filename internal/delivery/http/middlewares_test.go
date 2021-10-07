// +build unit

package http

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Mort4lis/scht-backend/internal/domain"
	mockservice "github.com/Mort4lis/scht-backend/internal/service/mocks"
	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/dgrijalva/jwt-go/v4"
	"github.com/golang/mock/gomock"
	"github.com/julienschmidt/httprouter"
	"github.com/stretchr/testify/assert"
)

func TestAuthorizationMiddleware(t *testing.T) {
	type mockBehaviour func(as *mockservice.MockAuthService, accessToken string, claims domain.Claims)

	testTable := []struct {
		name                     string
		authorizationHeaderKey   string
		authorizationHeaderValue string
		accessToken              string
		claims                   domain.Claims
		mockBehaviour            mockBehaviour
		expectedStatusCode       int
		expectedResponseBody     string
	}{
		{
			name:                     "Success",
			authorizationHeaderKey:   "Authorization",
			authorizationHeaderValue: "Bearer header.payload.sign",
			accessToken:              "header.payload.sign",
			claims: domain.Claims{
				StandardClaims: jwt.StandardClaims{
					Subject: "1",
				},
			},
			mockBehaviour: func(as *mockservice.MockAuthService, accessToken string, claims domain.Claims) {
				as.EXPECT().Authorize(accessToken).Return(claims, nil)
			},
			expectedStatusCode:   http.StatusOK,
			expectedResponseBody: `{"id":"1"}`,
		},
		{
			name:                 "Missed Authorization header",
			expectedStatusCode:   http.StatusBadRequest,
			expectedResponseBody: `{"message":"invalid Authorization header"}`,
		},
		{
			name:                     "Wrong Authorization header key",
			authorizationHeaderKey:   "Authorize",
			authorizationHeaderValue: "Bearer header.payload.sign",
			expectedStatusCode:       http.StatusBadRequest,
			expectedResponseBody:     `{"message":"invalid Authorization header"}`,
		},
		{
			name:                     "Wrong Authorization header value",
			authorizationHeaderKey:   "Authorization",
			authorizationHeaderValue: "Berer header.payload.sign",
			expectedStatusCode:       http.StatusBadRequest,
			expectedResponseBody:     `{"message":"invalid Authorization header"}`,
		},
		{
			name:                     "Empty Authorization header value",
			authorizationHeaderKey:   "Authorization",
			authorizationHeaderValue: "Bearer ",
			expectedStatusCode:       http.StatusBadRequest,
			expectedResponseBody:     `{"message":"invalid Authorization header"}`,
		},
		{
			name:                     "Invalid access token",
			authorizationHeaderKey:   "Authorization",
			authorizationHeaderValue: "Bearer header.payload.sign",
			accessToken:              "header.payload.sign",
			mockBehaviour: func(as *mockservice.MockAuthService, accessToken string, claims domain.Claims) {
				as.EXPECT().Authorize(accessToken).Return(domain.Claims{}, domain.ErrInvalidAccessToken)
			},
			expectedStatusCode:   http.StatusUnauthorized,
			expectedResponseBody: `{"message":"invalid access token"}`,
		},
		{
			name:                     "Unexpected error",
			authorizationHeaderKey:   "Authorization",
			authorizationHeaderValue: "Bearer header.payload.sign",
			accessToken:              "header.payload.sign",
			mockBehaviour: func(as *mockservice.MockAuthService, accessToken string, claims domain.Claims) {
				as.EXPECT().Authorize(accessToken).Return(domain.Claims{}, errors.New("unexpected error"))
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedResponseBody: `{"message":"internal server error"}`,
		},
	}

	logging.InitLogger(logging.LogConfig{
		LoggerKind: "mock",
	})

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			c := gomock.NewController(t)
			defer c.Finish()

			as := mockservice.NewMockAuthService(c)

			if testCase.mockBehaviour != nil {
				testCase.mockBehaviour(as, testCase.accessToken, testCase.claims)
			}

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)

			if testCase.authorizationHeaderKey != "" {
				req.Header.Add(testCase.authorizationHeaderKey, testCase.authorizationHeaderValue)
			}

			handler := authorizationMiddleware(func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
				id := domain.UserIDFromContext(req.Context())
				_, _ = fmt.Fprintf(w, `{"id":"%s"}`, id)
			}, as)

			handler(rec, req, httprouter.Params{})

			resp := rec.Result()

			respBodyPayload, err := ioutil.ReadAll(resp.Body)
			assert.NoError(t, err, "Unexpected error while reading response body")

			assert.Equal(t, testCase.expectedStatusCode, resp.StatusCode)
			assert.Equal(t, testCase.expectedResponseBody, string(respBodyPayload))
		})
	}
}
