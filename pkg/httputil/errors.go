package httputil

import (
	"fmt"
	"net/http"
)

type Error struct {
	Err        error          `json:"-"`
	Code       string         `json:"code"`
	StatusCode int            `json:"-"`
	Message    string         `json:"message"`
	Data       map[string]any `json:"data,omitempty"`
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

func (e Error) WithData(data map[string]any) Error {
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
	ErrInternalServer     = Error{Code: "CM0001", Message: "internal server error", StatusCode: http.StatusInternalServerError}
	ErrDecodeParamsFailed = Error{Code: "CM0002", Message: "decode params error", StatusCode: http.StatusBadRequest}
	ErrDecodeBodyFailed   = Error{Code: "CM0003", Message: "decode body error", StatusCode: http.StatusBadRequest}
	ErrValidationFailed   = Error{Code: "CM0004", Message: "validation error", StatusCode: http.StatusBadRequest}
)