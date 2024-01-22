package v1

import (
	"net/http"

	"github.com/Chatyx/backend/pkg/httputil"
)

var (
	errUserNotFound = httputil.Error{
		Code:       "US0001",
		Message:    "user is not found",
		StatusCode: http.StatusNotFound,
	}
	errSuchUserAlreadyExists = httputil.Error{
		Code:       "US0002",
		Message:    "user with such username or email already exists",
		StatusCode: http.StatusBadRequest,
	}
	errWrongCurrentPassword = httputil.Error{
		Code:       "US0003",
		Message:    "wrong current password",
		StatusCode: http.StatusBadRequest,
	}

	errGroupNotFound = httputil.Error{
		Code:       "CH0001",
		Message:    "group is not found",
		StatusCode: http.StatusNotFound,
	}
)
