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
	errInvalidDecodeBody = ResponseError{
		StatusCode: http.StatusBadRequest,
		Message:    "invalid body to decode",
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
)
