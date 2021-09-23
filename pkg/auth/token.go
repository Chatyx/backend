package auth

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/dgrijalva/jwt-go/v4"
)

var ErrInvalidTokenParse = fmt.Errorf("invalid token parse")

type TokenManager struct {
	signingKey []byte
}

func NewTokenManager(signingKey string) (*TokenManager, error) {
	if signingKey == "" {
		return nil, fmt.Errorf("empty signing key")
	}

	return &TokenManager{
		signingKey: []byte(signingKey),
	}, nil
}

func (m *TokenManager) NewAccessToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	sigToken, err := token.SignedString(m.signingKey)
	if err != nil {
		return "", err
	}

	return sigToken, nil
}

func (m *TokenManager) NewRefreshToken() (string, error) {
	src := rand.NewSource(time.Now().Unix())
	rnd := rand.New(src)

	buf := make([]byte, 32)
	if _, err := rnd.Read(buf); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", buf), nil
}

func (m *TokenManager) Parse(accessToken string, claims jwt.Claims) error {
	token, err := jwt.ParseWithClaims(accessToken, claims, func(token *jwt.Token) (interface{}, error) {
		return m.signingKey, nil
	})
	if err != nil {
		return ErrInvalidTokenParse
	}

	if err = token.Claims.Valid(jwt.DefaultValidationHelper); err != nil {
		return ErrInvalidTokenParse
	}

	return nil
}
