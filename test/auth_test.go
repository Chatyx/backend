// +build integration

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
	userID := "ba566522-3305-48df-936a-73f47611934b"
	resp := s.authenticateResponse("john1967", "qwerty12345", fingerprintValue)

	var tokenPair domain.JWTPair
	err := json.NewDecoder(resp.Body).Decode(&tokenPair)
	s.NoError(err, "Failed to decode response body")

	s.Require().NotEqual("", tokenPair.AccessToken)
	s.Require().NotEqual("", tokenPair.RefreshToken)

	s.checkRefreshTokenCookie(resp.Cookies(), tokenPair.RefreshToken)
	s.checkSession(domain.Session{
		UserID:       userID,
		RefreshToken: tokenPair.RefreshToken,
		Fingerprint:  fingerprintValue,
	})

	userSessionsKey := fmt.Sprintf("user:%s:sessions", userID)
	sessionKeys, err := s.redisClient.LRange(context.Background(), userSessionsKey, 0, -1).Result()
	s.NoError(err, "Failed to range user's sessions from redis")

	s.Require().Equal([]string{"session:" + tokenPair.RefreshToken}, sessionKeys)
}

func (s *AppTestSuite) TestSignInFailed() {
	userID := "ba566522-3305-48df-936a-73f47611934b"
	bodyStr := `{"username":"john1967","password":"qwerty_qQq"}`

	req, err := http.NewRequest(http.MethodPost, s.buildURL("/auth/sign-in"), strings.NewReader(bodyStr))
	s.NoError(err, "Failed to create request")

	req.Header.Set("X-Fingerprint", fingerprintValue)

	resp, err := s.httpClient.Do(req)
	s.Require().NoError(err, "Failed to send request")

	defer resp.Body.Close()
	s.Require().Equal(http.StatusUnauthorized, resp.StatusCode)

	userSessionsKey := fmt.Sprintf("user:%s:sessions", userID)
	val, err := s.redisClient.Exists(context.Background(), userSessionsKey).Result()
	s.NoError(err, "Failed to check exist userID key")
	s.Require().Equal(int64(0), val, "userID key must not exist")
}

func (s *AppTestSuite) TestRefreshWithBodySuccess() {
	userID := "ba566522-3305-48df-936a-73f47611934b"
	tokenPair := s.authenticate("john1967", "qwerty12345", fingerprintValue)

	bodyStr := fmt.Sprintf(`{"refresh_token":"%s"}`, tokenPair.RefreshToken)
	req, err := http.NewRequest(http.MethodPost, s.buildURL("/auth/refresh"), strings.NewReader(bodyStr))
	s.NoError(err, "Failed to create request")

	req.Header.Set("X-Fingerprint", fingerprintValue)

	resp, err := s.httpClient.Do(req)
	s.Require().NoError(err, "Failed to send request")

	defer resp.Body.Close()
	s.Require().Equal(http.StatusOK, resp.StatusCode)

	var newTokenPair domain.JWTPair
	err = json.NewDecoder(resp.Body).Decode(&newTokenPair)
	s.NoError(err, "Failed to decode response body")

	s.Require().NotEqual("", newTokenPair.AccessToken)
	s.Require().NotEqual("", newTokenPair.RefreshToken)

	s.checkRefreshTokenCookie(resp.Cookies(), newTokenPair.RefreshToken)
	s.checkSession(domain.Session{
		UserID:       userID,
		RefreshToken: newTokenPair.RefreshToken,
		Fingerprint:  fingerprintValue,
	})

	userSessionsKey := fmt.Sprintf("user:%s:sessions", userID)
	sessionKeys, err := s.redisClient.LRange(context.Background(), userSessionsKey, 0, -1).Result()
	s.NoError(err, "Failed to range user's sessions from redis")

	s.Require().Equal([]string{"session:" + newTokenPair.RefreshToken}, sessionKeys)

	val, err := s.redisClient.Exists(context.Background(), tokenPair.RefreshToken).Result()
	s.NoError(err, "Failed to check exist old session")
	s.Require().Equal(int64(0), val, "Old session exists")
}

func (s *AppTestSuite) TestRefreshWithCookieSuccess() {
	userID := "ba566522-3305-48df-936a-73f47611934b"
	resp := s.authenticateResponse("john1967", "qwerty12345", fingerprintValue)
	refreshCookie := s.getRefreshCookie(resp.Cookies())

	req, err := http.NewRequest(http.MethodPost, s.buildURL("/auth/refresh"), nil)
	s.NoError(err, "Failed to create request")

	req.AddCookie(refreshCookie)
	req.Header.Set("X-Fingerprint", fingerprintValue)

	resp, err = s.httpClient.Do(req)
	s.Require().NoError(err, "Failed to send request")

	defer resp.Body.Close()
	s.Require().Equal(http.StatusOK, resp.StatusCode)

	var newTokenPair domain.JWTPair
	err = json.NewDecoder(resp.Body).Decode(&newTokenPair)
	s.NoError(err, "Failed to decode response body")

	s.Require().NotEqual("", newTokenPair.AccessToken)
	s.Require().NotEqual("", newTokenPair.RefreshToken)

	s.checkRefreshTokenCookie(resp.Cookies(), newTokenPair.RefreshToken)
	s.checkSession(domain.Session{
		UserID:       userID,
		RefreshToken: newTokenPair.RefreshToken,
		Fingerprint:  fingerprintValue,
	})

	userSessionsKey := fmt.Sprintf("user:%s:sessions", userID)
	sessionKeys, err := s.redisClient.LRange(context.Background(), userSessionsKey, 0, -1).Result()
	s.NoError(err, "Failed to range user's sessions from redis")

	s.Require().Equal([]string{"session:" + newTokenPair.RefreshToken}, sessionKeys)

	val, err := s.redisClient.Exists(context.Background(), refreshCookie.Value).Result()
	s.NoError(err, "Failed to check exist old session")
	s.Require().Equal(int64(0), val, "Old session exists")
}

func (s *AppTestSuite) TestRefreshInvalidFingerprint() {
	tokenPair := s.authenticate("john1967", "qwerty12345", fingerprintValue)

	bodyStr := fmt.Sprintf(`{"refresh_token":"%s"}`, tokenPair.RefreshToken)
	req, err := http.NewRequest(http.MethodPost, s.buildURL("/auth/refresh"), strings.NewReader(bodyStr))
	s.NoError(err, "Failed to create request")

	req.Header.Add("X-Fingerprint", "invalid_fingerprint")

	resp, err := s.httpClient.Do(req)
	s.Require().NoError(err, "Failed to send request")

	defer resp.Body.Close()
	s.Require().Equal(http.StatusBadRequest, resp.StatusCode)
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
	payload, err := s.redisClient.Get(context.Background(), "session:"+expectedSession.RefreshToken).Result()
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

	sessionTTL, err := s.redisClient.TTL(context.Background(), "session:"+expectedSession.RefreshToken).Result()
	s.NoError(err, "Failed to get session TTL from redis")

	expectedSessionTTL := s.cfg.Auth.RefreshTokenTTL
	ttlDiff := expectedSessionTTL - sessionTTL
	s.Require().GreaterOrEqual(ttlDiff, time.Duration(0), "Wrong session TTL")
	s.Require().LessOrEqual(expectedSessionTTL-sessionTTL, 5*time.Second, "Wrong session TTL")
}

func (s *AppTestSuite) authenticate(username, password, fingerprint string) domain.JWTPair {
	resp := s.authenticateResponse(username, password, fingerprint)
	defer resp.Body.Close()

	var tokenPair domain.JWTPair
	err := json.NewDecoder(resp.Body).Decode(&tokenPair)
	s.NoError(err, "Failed to decode response body")

	s.Require().NotEqual("", tokenPair.AccessToken)
	s.Require().NotEqual("", tokenPair.RefreshToken)

	return tokenPair
}

func (s *AppTestSuite) authenticateResponse(username, password, fingerprint string) *http.Response {
	bodyStr := fmt.Sprintf(`{"username":"%s","password":"%s"}`, username, password)
	req, err := http.NewRequest(http.MethodPost, s.buildURL("/auth/sign-in"), strings.NewReader(bodyStr))
	s.NoError(err, "Failed to create authenticate request")

	req.Header.Set("X-Fingerprint", fingerprint)

	resp, err := s.httpClient.Do(req)
	s.Require().NoError(err, "Failed to authenticate send request")
	s.Require().Equal(http.StatusOK, resp.StatusCode)

	return resp
}
