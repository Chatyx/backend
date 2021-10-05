package test

import (
	"context"
	"encoding/json"
	"fmt"
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
	resp := s.authenticateResponse("john1967", "qwerty12345", fingerprintValue)

	var tokenPair domain.JWTPair
	err := json.NewDecoder(resp.Body).Decode(&tokenPair)
	s.NoError(err, "Failed to decode response body")

	s.Require().NotEqual("", tokenPair.AccessToken)
	s.Require().NotEqual("", tokenPair.RefreshToken)

	s.checkRefreshTokenCookie(resp.Cookies(), tokenPair.RefreshToken)
	s.checkSession(domain.Session{
		UserID:       "ba566522-3305-48df-936a-73f47611934b",
		RefreshToken: tokenPair.RefreshToken,
		Fingerprint:  fingerprintValue,
	})
}

func (s *AppTestSuite) TestSignInFailed() {
	bodyStr := `{"username":"john1967","password":"qwerty_qQq"}`
	req, err := http.NewRequest(http.MethodPost, s.getURL("/auth/sign-in"), strings.NewReader(bodyStr))
	s.NoError(err, "Failed to create request")

	req.Header.Set("X-Fingerprint", fingerprintValue)

	httpClient := &http.Client{Timeout: 5 * time.Second}
	resp, err := httpClient.Do(req)
	s.NoError(err, "Failed to send request")
	s.Require().Equal(http.StatusUnauthorized, resp.StatusCode)
}

func (s *AppTestSuite) TestRefreshWithBodySuccess() {
	tokenPair := s.authenticate("john1967", "qwerty12345", fingerprintValue)

	bodyStr := fmt.Sprintf(`{"refresh_token":"%s"}`, tokenPair.RefreshToken)
	req, err := http.NewRequest(http.MethodPost, s.getURL("/auth/refresh"), strings.NewReader(bodyStr))
	s.NoError(err, "Failed to create request")

	req.Header.Set("X-Fingerprint", fingerprintValue)

	httpClient := &http.Client{Timeout: 5 * time.Second}
	resp, err := httpClient.Do(req)
	s.NoError(err, "Failed to send request")

	defer func() { _ = resp.Body.Close() }()
	s.Require().Equal(http.StatusOK, resp.StatusCode)

	var newTokenPair domain.JWTPair
	err = json.NewDecoder(resp.Body).Decode(&newTokenPair)
	s.NoError(err, "Failed to decode response body")

	s.Require().NotEqual("", newTokenPair.AccessToken)
	s.Require().NotEqual("", newTokenPair.RefreshToken)

	s.checkRefreshTokenCookie(resp.Cookies(), newTokenPair.RefreshToken)
	s.checkSession(domain.Session{
		UserID:       "ba566522-3305-48df-936a-73f47611934b",
		RefreshToken: newTokenPair.RefreshToken,
		Fingerprint:  fingerprintValue,
	})

	val, err := s.redisClient.Exists(context.Background(), tokenPair.RefreshToken).Result()
	s.NoError(err, "Failed to check exist old session")
	s.Require().Equal(val, int64(0), "Old session exists")
}

func (s *AppTestSuite) TestRefreshWithCookieSuccess() {
	resp := s.authenticateResponse("john1967", "qwerty12345", fingerprintValue)
	refreshCookie := s.getRefreshCookie(resp.Cookies())

	req, err := http.NewRequest(http.MethodPost, s.getURL("/auth/refresh"), nil)
	s.NoError(err, "Failed to create request")

	req.AddCookie(refreshCookie)
	req.Header.Set("X-Fingerprint", fingerprintValue)

	httpClient := &http.Client{Timeout: 5 * time.Second}
	resp, err = httpClient.Do(req)
	s.NoError(err, "Failed to send request")

	defer func() { _ = resp.Body.Close() }()
	s.Require().Equal(http.StatusOK, resp.StatusCode)

	var newTokenPair domain.JWTPair
	err = json.NewDecoder(resp.Body).Decode(&newTokenPair)
	s.NoError(err, "Failed to decode response body")

	s.Require().NotEqual("", newTokenPair.AccessToken)
	s.Require().NotEqual("", newTokenPair.RefreshToken)

	s.checkRefreshTokenCookie(resp.Cookies(), newTokenPair.RefreshToken)
	s.checkSession(domain.Session{
		UserID:       "ba566522-3305-48df-936a-73f47611934b",
		RefreshToken: newTokenPair.RefreshToken,
		Fingerprint:  fingerprintValue,
	})

	val, err := s.redisClient.Exists(context.Background(), refreshCookie.Value).Result()
	s.NoError(err, "Failed to check exist old session")
	s.Require().Equal(val, int64(0), "Old session exists")
}

func (s *AppTestSuite) checkRefreshTokenCookie(cookies []*http.Cookie, expectedRefreshToken string) {
	refreshCookie := s.getRefreshCookie(cookies)
	s.Require().Equal(refreshCookie.Value, expectedRefreshToken)
	s.Require().Equal(refreshCookie.Path, "/api/auth/refresh")
	s.Require().True(refreshCookie.HttpOnly, "Refresh token cookie must be http only")

	refreshCookieTTL := time.Duration(refreshCookie.MaxAge) * time.Second
	if !refreshCookie.Expires.IsZero() {
		refreshCookieTTL = time.Until(refreshCookie.Expires)
	}

	ttlDiff := s.cfg.Auth.RefreshTokenTTL - refreshCookieTTL
	s.Require().GreaterOrEqual(ttlDiff, time.Duration(0), "Wrong refresh cookie TTL")
	s.Require().LessOrEqual(ttlDiff, 5*time.Second, "Wrong refresh cookie TTL")
}

func (s *AppTestSuite) getRefreshCookie(cookies []*http.Cookie) *http.Cookie {
	var refreshCookie *http.Cookie

	for _, cookie := range cookies {
		if cookie.Name == refreshCookieName {
			refreshCookie = cookie
			break
		}
	}

	s.Require().NotNil(refreshCookie, "Refresh cookie doesn't exist")

	return refreshCookie
}

func (s *AppTestSuite) checkSession(expectedSession domain.Session) {
	payload, err := s.redisClient.Get(context.Background(), expectedSession.RefreshToken).Result()
	s.NoError(err, "Failed to get session from redis")

	var session domain.Session
	err = json.Unmarshal([]byte(payload), &session)
	s.NoError(err, "Failed to unmarshal session json bytes")

	s.Require().Equal(expectedSession.UserID, session.UserID)
	s.Require().Equal(expectedSession.RefreshToken, session.RefreshToken)
	s.Require().Equal(expectedSession.Fingerprint, session.Fingerprint)

	if !expectedSession.CreatedAt.IsZero() {
		s.Require().Equal(expectedSession.CreatedAt, session.CreatedAt)
	}

	if !expectedSession.ExpiresAt.IsZero() {
		s.Require().Equal(expectedSession.ExpiresAt, session.ExpiresAt)
	}

	sessionTTL, err := s.redisClient.TTL(context.Background(), expectedSession.RefreshToken).Result()
	s.NoError(err, "Failed to get session TTL from redis")

	expectedSessionTTL := s.cfg.Auth.RefreshTokenTTL
	ttlDiff := expectedSessionTTL - sessionTTL
	s.Require().GreaterOrEqual(ttlDiff, time.Duration(0), "Wrong session TTL")
	s.Require().LessOrEqual(expectedSessionTTL-sessionTTL, 5*time.Second, "Wrong session TTL")
}

func (s *AppTestSuite) authenticate(username, password, fingerprint string) domain.JWTPair {
	resp := s.authenticateResponse(username, password, fingerprint)
	defer func() { _ = resp.Body.Close() }()

	var tokenPair domain.JWTPair
	err := json.NewDecoder(resp.Body).Decode(&tokenPair)
	s.NoError(err, "Failed to decode response body")

	s.Require().NotEqual("", tokenPair.AccessToken)
	s.Require().NotEqual("", tokenPair.RefreshToken)

	return tokenPair
}

func (s *AppTestSuite) authenticateResponse(username, password, fingerprint string) *http.Response {
	bodyStr := fmt.Sprintf(`{"username":"%s","password":"%s"}`, username, password)
	req, err := http.NewRequest(http.MethodPost, s.getURL("/auth/sign-in"), strings.NewReader(bodyStr))
	s.NoError(err, "Failed to create authenticate request")

	req.Header.Set("X-Fingerprint", fingerprint)

	httpClient := &http.Client{Timeout: 5 * time.Second}
	resp, err := httpClient.Do(req)
	s.NoError(err, "Failed to authenticate send request")
	s.Require().Equal(http.StatusOK, resp.StatusCode)

	return resp
}
