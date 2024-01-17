package middleware

import (
	"net/http"

	"github.com/Chatyx/backend/pkg/ctxutil"
	"github.com/Chatyx/backend/pkg/log"
	"github.com/google/uuid"
)

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		id := uuid.New().String()
		logger := log.FromContext(req.Context())
		logger.With("request_id", id)

		ctx := log.WithLogger(ctxutil.WithRequestID(req.Context(), id), logger)
		next.ServeHTTP(w, req.WithContext(ctx))
	})
}
