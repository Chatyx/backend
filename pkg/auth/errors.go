package auth

import "errors"

var (
	ErrUserNotFound        = errors.New("user is not found")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
)
