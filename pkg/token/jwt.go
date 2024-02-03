package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims map[string]any

type JWT struct {
	issuer    string
	ttl       time.Duration
	signedKey any
}

func NewJWT(issuer string, signedKey any, ttl time.Duration) JWT {
	return JWT{
		issuer:    issuer,
		ttl:       ttl,
		signedKey: signedKey,
	}
}

func (j JWT) Token(subject string, extra Claims) (string, error) {
	claims := jwt.MapClaims{}
	for k, v := range extra {
		claims[k] = v
	}

	issuedAt := time.Now()
	claims["jti"] = uuid.New().String()
	claims["iss"] = j.issuer
	claims["sub"] = subject
	claims["iat"] = float64(issuedAt.Unix())
	claims["exp"] = float64(issuedAt.Add(j.ttl).Unix())

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(j.signedKey)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	return signedToken, nil
}
