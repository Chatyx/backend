package domain

import "errors"

var (
	ErrUserNotFound        = errors.New("user is not found")
	ErrUserUniqueViolation = errors.New("user with such username or email is already exist")
	ErrUserNoNeedUpdate    = errors.New("no need to update user")
	ErrInvalidCredentials  = errors.New("invalid username or password")
	ErrInvalidAccessToken  = errors.New("invalid access token")
	ErrSessionNotFound     = errors.New("refresh session is not found")
)
