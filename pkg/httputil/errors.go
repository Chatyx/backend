package httputil

import (
	"fmt"
	"net/http"
	"strings"
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
	if e.Err != nil {
		errString := e.Err.Error()
		if strings.HasPrefix(errString, e.Message) {
			return fmt.Sprintf("[%s] - %s", e.Code, errString)
		}

		return fmt.Sprintf("[%s] - %s: %s", e.Code, e.Message, errString)
	}
	return fmt.Sprintf("[%s] - %s", e.Code, e.Message)
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
	ErrDecodeHeaderParamsFailed = Error{
		Code:       "CM0004",
		Message:    "decode header params error",
		StatusCode: http.StatusBadRequest,
	}
	ErrDecodeQueryParamsFailed = Error{
		Code:       "CM0005",
		Message:    "decode query params error",
		StatusCode: http.StatusBadRequest,
	}
	ErrValidationFailed = Error{
		Code:       "CM0006",
		Message:    "validation error",
		StatusCode: http.StatusBadRequest,
	}
	ErrInvalidAuthorization = Error{
		Code:       "CM0007",
		Message:    "invalid Authorization header or query param",
		StatusCode: http.StatusBadRequest,
	}
	ErrForbiddenPerformAction = Error{
		Code:       "CM0008",
		Message:    "it's forbidden to perform this action",
		StatusCode: http.StatusForbidden,
	}
)
