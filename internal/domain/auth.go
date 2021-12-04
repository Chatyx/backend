package domain

import (
	"time"

	"github.com/dgrijalva/jwt-go/v4"
)

type AuthUser struct {
	UserID string
}

type Session struct {
	UserID       string
	RefreshToken string
	Fingerprint  string
	ExpiresAt    time.Time
	CreatedAt    time.Time
}

type Claims struct {
	jwt.StandardClaims
}

type SignInDTO struct {
	Username    string `json:"username" validate:"required"`
	Password    string `json:"password" validate:"required"`
	Fingerprint string `json:"-"`
}

type JWTPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshSessionDTO struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
	Fingerprint  string `json:"-"`
}
