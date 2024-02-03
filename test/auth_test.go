package test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Chatyx/backend/pkg/auth"
)

const defaultFingerprint = "12345"

func (s *AppTestSuite) authenticate(username string) (auth.TokenPair, error) {
	bodyStr := fmt.Sprintf(`{"username":"%s","password":"%s"}`, username, "qwerty12345")
	req, err := http.NewRequest(http.MethodPost, s.apiURLFromPath("/api/v1/auth/login").String(), strings.NewReader(bodyStr))
	if err != nil {
		return auth.TokenPair{}, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("X-Fingerprint", defaultFingerprint)

	resp, err := s.httpCli.Do(req)
	if err != nil {
		return auth.TokenPair{}, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return auth.TokenPair{}, fmt.Errorf("got %d response status code", resp.StatusCode)
	}

	var tokens struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		return auth.TokenPair{}, fmt.Errorf("decode response body: %w", err)
	}

	return auth.TokenPair{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}
