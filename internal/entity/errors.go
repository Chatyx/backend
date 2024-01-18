package entity

import "errors"

var (
	ErrUserNotFound          = errors.New("user is not found")
	ErrSuchUserAlreadyExists = errors.New("user with such username or email already exists")
	ErrWrongCurrentPassword  = errors.New("wrong current password")
)
