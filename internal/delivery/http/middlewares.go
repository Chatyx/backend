package http

import (
	"errors"
	"net/http"
	"time"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/internal/service"
	"github.com/Mort4lis/scht-backend/pkg/logging"
)

type Middleware func(handler http.Handler) http.Handler

type StatusCodeRecorder struct {
	http.ResponseWriter
	StatusCode int
}

func (rec *StatusCodeRecorder) WriteHeader(statusCode int) {
	rec.StatusCode = statusCode
	rec.ResponseWriter.WriteHeader(statusCode)
}

func loggingMiddleware(handler http.Handler) http.Handler {
	logger := logging.GetLogger()

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		rec := &StatusCodeRecorder{
			ResponseWriter: w,
		}
		start := time.Now()

		handler.ServeHTTP(rec, req)

		logger.WithFields(logging.Fields{
			"path":        req.URL.EscapedPath(),
			"method":      req.Method,
			"remote_addr": req.RemoteAddr,
			"status":      rec.StatusCode,
			"duration":    time.Since(start),
		}).Info("Request has been handled")
	})
}

func AuthorizationMiddlewareFactory(as service.AuthService) Middleware {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			accessToken, err := extractTokenFromHeader(req.Header.Get("Authorization"))
			if err != nil {
				respondError(w, err)
				return
			}

			claims, err := as.Authorize(accessToken)
			if err != nil {
				switch {
				case errors.Is(err, domain.ErrInvalidAccessToken):
					respondErrorRefactored(req.Context(), w, errInvalidAccessToken.Wrap(err))
				default:
					respondErrorRefactored(req.Context(), w, err)
				}

				return
			}

			authUser := domain.AuthUser{UserID: claims.Subject}
			logger := logging.GetLogger().WithFields(logging.Fields{
				"auth.user_id": authUser.UserID,
			})

			ctx := domain.NewContextFromAuthUser(logging.NewContextFromLogger(req.Context(), logger), authUser)

			handler.ServeHTTP(w, req.WithContext(ctx))
		})
	}
}
