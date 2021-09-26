package domain

import (
	"encoding/json"
	"io"
	"time"

	"github.com/dgrijalva/jwt-go/v4"
)

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

func (s *SignInDTO) Decode(payload []byte) error {
	return json.Unmarshal(payload, s)
}

func (s *SignInDTO) DecodeFrom(r io.Reader) error {
	return json.NewDecoder(r).Decode(s)
}

type JWTPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (t JWTPair) Encode() ([]byte, error) {
	return json.Marshal(t)
}

type RefreshSessionDTO struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
	Fingerprint  string `json:"-"`
}

func (rt *RefreshSessionDTO) Decode(payload []byte) error {
	return json.Unmarshal(payload, rt)
}

func (rt *RefreshSessionDTO) DecodeFrom(r io.Reader) error {
	return json.NewDecoder(r).Decode(rt)
}
