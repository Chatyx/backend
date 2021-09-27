package domain

import "errors"

var (
	ErrUserNotFound        = errors.New("user is not found")
	ErrUserUniqueViolation = errors.New("user with such username or email already exists")
	ErrUserNoNeedUpdate    = errors.New("no need to update user")
	ErrWrongCredentials    = errors.New("wrong username or password")
	ErrInvalidAccessToken  = errors.New("invalid access token")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrSessionNotFound     = errors.New("refresh session is not found")
)
