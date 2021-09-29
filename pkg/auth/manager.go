package auth

import (
	"fmt"

	"github.com/dgrijalva/jwt-go/v4"
)

var ErrInvalidTokenParse = fmt.Errorf("invalid token parse")

//go:generate mockgen -source=manager.go -destination=mocks/mock.go

type TokenManager interface {
	NewAccessToken(claims jwt.Claims) (string, error)
	NewRefreshToken() (string, error)
	Parse(accessToken string, claims jwt.Claims) error
}
