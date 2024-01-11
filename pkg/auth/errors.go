package auth

import "errors"

var (
	ErrUserNotFound        = errors.New("user is not found")
	ErrSessionNotFound     = errors.New("session is not found")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
)
