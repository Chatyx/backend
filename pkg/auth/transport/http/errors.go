package http

import (
	"net/http"

	"github.com/Chatyx/backend/pkg/httputil"
)

var (
	ErrFailedLogin         = httputil.Error{Code: "AU0001", Message: "failed login", StatusCode: http.StatusUnauthorized}
	ErrInvalidRefreshToken = httputil.Error{Code: "AU0002", Message: "invalid refresh token", StatusCode: http.StatusBadRequest}
)
