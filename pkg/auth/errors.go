package auth

import "errors"

var (
	ErrWrongCredentials    = errors.New("wrong username or password")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrSessionNotFound     = errors.New("session is not found")
)
