package http

import "net/http"

type ErrorFields map[string]string

type ResponseError struct {
	StatusCode int         `json:"-"`
	Message    string      `json:"message"`
	Fields     ErrorFields `json:"fields,omitempty"`
}

func (e ResponseError) Error() string {
	return e.Message
}

var (
	ErrInternalServer = ResponseError{
		StatusCode: http.StatusInternalServerError,
		Message:    "internal server error",
	}
	ErrInvalidJSON = ResponseError{
		StatusCode: http.StatusBadRequest,
		Message:    "invalid json body",
	}
	ErrInvalidAuthorizationToken = ResponseError{
		StatusCode: http.StatusUnauthorized,
		Message:    "invalid authorization token",
	}
	ErrInvalidRefreshToken = ResponseError{
		StatusCode: http.StatusBadRequest,
		Message:    "invalid refresh token",
	}
)
