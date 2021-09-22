package domain

import "net/http"

type ErrorFields map[string]string

type AppError struct {
	StatusCode int         `json:"-"`
	Message    string      `json:"message"`
	Fields     ErrorFields `json:"fields,omitempty"`
}

func (e AppError) Error() string {
	return e.Message
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
)