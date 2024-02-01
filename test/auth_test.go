package test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Chatyx/backend/pkg/auth"
)

const defaultFingerprint = "12345"

func (s *AppTestSuite) authenticate(username, password string) auth.TokenPair {
	bodyStr := fmt.Sprintf(`{"username":"%s","password":"%s"}`, username, password)
	req, err := http.NewRequest(http.MethodPost, s.apiURLFromPath("/api/v1/auth/login"), strings.NewReader(bodyStr))
	s.Require().NoError(err, "Failed to create authentication request")

	req.Header.Set("X-Fingerprint", defaultFingerprint)

	resp, err := s.httpCli.Do(req)
	s.Require().NoError(err, "Failed to authenticate send request")
	s.Require().Equal(http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()

	var tokens struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	err = json.NewDecoder(resp.Body).Decode(&tokens)
	s.NoError(err, "Failed to decode response body")

	s.NotEmpty(tokens.AccessToken, "Access token is empty")
	s.NotEmpty(tokens.RefreshToken, "Refresh token is empty")

	return auth.TokenPair{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}
}
