package http

import (
	"fmt"
	"net/http"
)

type ErrorFields map[string]string

type ResponseError struct {
	Err        error       `json:"-"`
	StatusCode int         `json:"-"`
	Message    string      `json:"message"`
	Fields     ErrorFields `json:"fields,omitempty"`
}

type ResponseErrorOption func(e *ResponseError)

func WithFields(fields ErrorFields) ResponseErrorOption {
	return func(e *ResponseError) {
		e.Fields = fields
	}
}

func NewResponseError(statusCode int, message string, opts ...ResponseErrorOption) ResponseError {
	respErr := ResponseError{
		StatusCode: statusCode,
		Message:    message,
	}

	for _, opt := range opts {
		opt(&respErr)
	}

	return respErr
}

func WrapInResponseError(err error, statusCode int, message string, opts ...ResponseErrorOption) ResponseError {
	respErr := NewResponseError(statusCode, message, opts...)
	respErr.Err = err

	return respErr
}

func (e ResponseError) Wrap(err error, opts ...ResponseErrorOption) error {
	return WrapInResponseError(err, e.StatusCode, e.Message, opts...)
}

func (e ResponseError) Unwrap() error {
	return e.Err
}

func (e ResponseError) Error() string {
	return fmt.Sprintf("%s: %v", e.Message, e.Err)
}

var (
	errInternalServer = ResponseError{
		StatusCode: http.StatusInternalServerError,
		Message:    "internal server error",
	}
	errInvalidDecodeBody = ResponseError{
		StatusCode: http.StatusBadRequest,
		Message:    "invalid body to decode",
	}
	errValidationError = ResponseError{
		StatusCode: http.StatusBadRequest,
		Message:    "validation error",
	}
	errInvalidAuthorizationHeader = ResponseError{
		StatusCode: http.StatusBadRequest,
		Message:    "invalid Authorization header",
	}
	errInvalidAccessToken = ResponseError{
		StatusCode: http.StatusUnauthorized,
		Message:    "invalid access token",
	}
	errInvalidRefreshToken = ResponseError{
		StatusCode: http.StatusBadRequest,
		Message:    "invalid refresh token",
	}
	errEmptyFingerprintHeader = ResponseError{
		StatusCode: http.StatusBadRequest,
		Message:    "X-Fingerprint header is required",
	}

	errUserNotFound = ResponseError{
		StatusCode: http.StatusNotFound,
		Message:    "user is not found",
	}
	errUserUniqueViolation = ResponseError{
		StatusCode: http.StatusBadRequest,
		Message:    "user with such username or email already exists",
	}
	errWrongCurrentPassword = ResponseError{
		StatusCode: http.StatusBadRequest,
		Message:    "wrong current password",
	}
)
