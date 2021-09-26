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
	errInternalServer = ResponseError{
		StatusCode: http.StatusInternalServerError,
		Message:    "internal server error",
	}
	errInvalidJSON = ResponseError{
		StatusCode: http.StatusBadRequest,
		Message:    "invalid json body",
	}
	errInvalidAuthorizationToken = ResponseError{
		StatusCode: http.StatusUnauthorized,
		Message:    "invalid authorization token",
	}
	errInvalidRefreshToken = ResponseError{
		StatusCode: http.StatusBadRequest,
		Message:    "invalid refresh token",
	}
	errEmptyFingerprintHeader = ResponseError{
		StatusCode: http.StatusBadRequest,
		Message:    "X-Fingerprint header is required",
	}
)
