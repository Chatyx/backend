package test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/Mort4lis/scht-backend/internal/domain"
)

const (
	refreshCookieName = "refresh_token"
	fingerprintValue  = "6d918360-3211-4d99-8c5c-0bc0cefd66c2"
)

func (s *AppTestSuite) TestSignInSuccess() {
	refreshTokenTTL := s.cfg.Auth.RefreshTokenTTL

	reqStr := `{"username":"john1967","password":"qwerty12345"}`
	req, err := http.NewRequest(http.MethodPost, s.getURL("/auth/sign-in"), strings.NewReader(reqStr))
	s.NoError(err, "Failed to create request")

	req.Header.Set("X-Fingerprint", fingerprintValue)

	httpClient := &http.Client{Timeout: 5 * time.Second}
	resp, err := httpClient.Do(req)
	s.NoError(err, "Failed to send request")
	s.Require().Equal(http.StatusOK, resp.StatusCode)

	var tokenPair domain.JWTPair
	err = json.NewDecoder(resp.Body).Decode(&tokenPair)
	s.NoError(err, "Failed to decode response body")

	s.Require().NotEqual("", tokenPair.AccessToken)
	s.Require().NotEqual("", tokenPair.RefreshToken)

	var refreshCookie *http.Cookie

	cookies := resp.Cookies()
	for _, cookie := range cookies {
		if cookie.Name == refreshCookieName {
			refreshCookie = cookie
			break
		}
	}

	s.Require().NotNil(refreshCookie)
	s.Require().Equal(refreshCookie.Value, tokenPair.RefreshToken)
	s.Require().Equal(refreshCookie.Path, "/api/auth/refresh")
	s.Require().True(refreshCookie.HttpOnly, "Refresh token cookie must be http only")

	refreshCookieTTL := time.Duration(refreshCookie.MaxAge) * time.Second
	if !refreshCookie.Expires.IsZero() {
		refreshCookieTTL = time.Until(refreshCookie.Expires)
	}

	s.Require().LessOrEqual(
		refreshTokenTTL-refreshCookieTTL, 5*time.Second,
		"Wrong refresh cookie TTL",
	)

	payload, err := s.redisClient.Get(context.Background(), tokenPair.RefreshToken).Result()
	s.NoError(err, "Failed to get session from redis")

	var session domain.Session
	err = json.Unmarshal([]byte(payload), &session)
	s.NoError(err, "Failed to unmarshal session json bytes")

	s.Require().Equal("ba566522-3305-48df-936a-73f47611934b", session.UserID)
	s.Require().Equal(tokenPair.RefreshToken, session.RefreshToken)
	s.Require().Equal(fingerprintValue, session.Fingerprint)

	sessionTTL, err := s.redisClient.TTL(context.Background(), tokenPair.RefreshToken).Result()
	s.NoError(err, "Failed to get session TTL from redis")
	s.Require().LessOrEqual(refreshTokenTTL-sessionTTL, 5*time.Second, "Wrong session TTL")
}

func (s *AppTestSuite) TestSignInFailed() {
	reqStr := `{"username":"john1967","password":"qwerty_qQq"}`
	req, err := http.NewRequest(http.MethodPost, s.getURL("/auth/sign-in"), strings.NewReader(reqStr))
	s.NoError(err, "Failed to create request")

	req.Header.Set("X-Fingerprint", fingerprintValue)

	httpClient := &http.Client{Timeout: 5 * time.Second}
	resp, err := httpClient.Do(req)
	s.NoError(err, "Failed to send request")
	s.Require().Equal(http.StatusUnauthorized, resp.StatusCode)
}
