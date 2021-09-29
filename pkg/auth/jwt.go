package auth

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/dgrijalva/jwt-go/v4"
)

type JWTTokenManager struct {
	signingKey []byte
}

func NewJWTTokenManager(signingKey string) (*JWTTokenManager, error) {
	if signingKey == "" {
		return nil, fmt.Errorf("empty signing key")
	}

	return &JWTTokenManager{
		signingKey: []byte(signingKey),
	}, nil
}

func (m *JWTTokenManager) NewAccessToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	sigToken, err := token.SignedString(m.signingKey)
	if err != nil {
		return "", err
	}

	return sigToken, nil
}

func (m *JWTTokenManager) NewRefreshToken() (string, error) {
	src := rand.NewSource(time.Now().Unix())
	rnd := rand.New(src)

	buf := make([]byte, 32)
	if _, err := rnd.Read(buf); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", buf), nil
}

func (m *JWTTokenManager) Parse(accessToken string, claims jwt.Claims) error {
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
