package middleware

import (
	"net/http"
	"time"

	"github.com/Chatyx/backend/pkg/log"
)

type statusCodeRecorder struct {
	http.ResponseWriter
	StatusCode int
}

func (rec *statusCodeRecorder) WriteHeader(statusCode int) {
	rec.StatusCode = statusCode
	rec.ResponseWriter.WriteHeader(statusCode)
}

func Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		rec := &statusCodeRecorder{ResponseWriter: w}

		begin := time.Now()
		next.ServeHTTP(rec, req)

		logger := log.FromContext(req.Context())
		logger.With(
			"path", req.URL.EscapedPath(),
			"user_agent", req.UserAgent(),
			"method", req.Method,
			"ip", req.RemoteAddr,
			"host", req.Host,
			"request_size", req.ContentLength,
			"duration", time.Since(begin).String(),
			"status_code", rec.StatusCode,
		).Info("Request handled")
	})
}
