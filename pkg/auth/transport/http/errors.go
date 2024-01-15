package http

import (
	"net/http"

	"github.com/Chatyx/backend/pkg/httputil"
)

var (
	ErrLoginFailed         = httputil.Error{Code: "AU0001", Message: "login failed", StatusCode: http.StatusUnauthorized}
	ErrInvalidRefreshToken = httputil.Error{Code: "AU0002", Message: "invalid refresh token", StatusCode: http.StatusBadRequest}
)
