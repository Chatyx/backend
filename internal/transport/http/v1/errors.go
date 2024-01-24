package v1

import (
	"net/http"

	"github.com/Chatyx/backend/pkg/httputil"
)

// user errors.
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
)

// chat (groups/dialogs) and participant errors.
var (
	errGroupNotFound = httputil.Error{
		Code:       "CH0001",
		Message:    "group is not found",
		StatusCode: http.StatusNotFound,
	}
	errDialogNotFound = httputil.Error{
		Code:       "CH0002",
		Message:    "dialog is not found",
		StatusCode: http.StatusNotFound,
	}
	errSuchDialogAlreadyExists = httputil.Error{
		Code:       "CH0003",
		Message:    "such a dialog already exists",
		StatusCode: http.StatusBadRequest,
	}
	errCreatingDialogWithYourself = httputil.Error{
		Code:       "CH0004",
		Message:    "creating a dialog with yourself",
		StatusCode: http.StatusBadRequest,
	}
	errCreatingDialogWithNonExistenceUser = httputil.Error{
		Code:       "CH0005",
		Message:    "creating a dialog with a non-existent user",
		StatusCode: http.StatusBadRequest,
	}
)
