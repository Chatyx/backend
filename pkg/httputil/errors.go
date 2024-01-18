package httputil

import (
	"fmt"
	"net/http"
)

type Error struct {
	Err        error             `json:"-"`
	Code       string            `json:"code"`
	StatusCode int               `json:"-"`
	Message    string            `json:"message"`
	Data       map[string]string `json:"data,omitempty"`
}

func (e Error) Wrap(err error) Error {
	return Error{
		Err:        err,
		Code:       e.Code,
		StatusCode: e.StatusCode,
		Message:    e.Message,
		Data:       e.Data,
	}
}

func (e Error) WithData(data map[string]string) Error {
	return Error{
		Err:        e.Err,
		Code:       e.Code,
		StatusCode: e.StatusCode,
		Message:    e.Message,
		Data:       data,
	}
}

func (e Error) Unwrap() error {
	return e.Err
}

func (e Error) Error() string {
	res := fmt.Sprintf("[%s] - %s", e.Code, e.Message)
	if e.Err != nil {
		res += ": " + e.Err.Error()
	}

	return res
}

var (
	ErrInternalServer = Error{
		Code:       "CM0001",
		Message:    "internal server error",
		StatusCode: http.StatusInternalServerError,
	}
	ErrDecodeBodyFailed = Error{
		Code:       "CM0002",
		Message:    "decode body error",
		StatusCode: http.StatusBadRequest,
	}
	ErrDecodePathParamsFailed = Error{
		Code:       "CM0003",
		Message:    "decode path params error",
		StatusCode: http.StatusBadRequest,
	}
	ErrDecodeQueryParamsFailed = Error{
		Code:       "CM0004",
		Message:    "decode query params error",
		StatusCode: http.StatusBadRequest,
	}
	ErrValidationFailed = Error{
		Code:       "CM0005",
		Message:    "validation error",
		StatusCode: http.StatusBadRequest,
	}
	ErrInvalidAuthorization = Error{
		Code:       "CM0006",
		Message:    "invalid Authorization header or query param",
		StatusCode: http.StatusBadRequest,
	}
)
