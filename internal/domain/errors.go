package domain

import (
	"net/http"
	"reflect"
)

type ErrorFields map[string]string

type AppError struct {
	StatusCode int         `json:"-"`
	Message    string      `json:"message"`
	Fields     ErrorFields `json:"fields,omitempty"`
}

func (e AppError) Error() string {
	return e.Message
}

func (e AppError) Is(err error) bool {
	target, ok := err.(AppError)
	if !ok {
		return false
	}

	return reflect.DeepEqual(e, target)
}

func NewValidationError(fields ErrorFields) AppError {
	return NewValidationMessageError("validation error", fields)
}

func NewValidationMessageError(message string, fields ErrorFields) AppError {
	return AppError{
		StatusCode: http.StatusBadRequest,
		Message:    message,
		Fields:     fields,
	}
}

var (
	ErrInternalServer = AppError{
		StatusCode: http.StatusInternalServerError,
		Message:    "internal server error",
	}
	ErrInvalidJSON = AppError{
		StatusCode: http.StatusBadRequest,
		Message:    "invalid json body",
	}
	ErrUserNotFound = AppError{
		StatusCode: http.StatusNotFound,
		Message:    "user is not found",
	}
	ErrUserUniqueViolation = AppError{
		StatusCode: http.StatusBadRequest,
		Message:    "user with such username or email is already exist",
	}
	ErrUserNoNeedUpdate = AppError{
		StatusCode: http.StatusBadRequest,
		Message:    "no need to update user",
	}
	ErrInvalidCredentials = AppError{
		StatusCode: http.StatusBadRequest,
		Message:    "invalid username or password",
	}
	ErrInvalidToken = AppError{
		StatusCode: http.StatusUnauthorized,
		Message:    "invalid token",
	}
)
