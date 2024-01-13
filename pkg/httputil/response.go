package httputil

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Chatyx/backend/pkg/log"
)

func RespondSuccess(ctx context.Context, w http.ResponseWriter, statusCode int, v any) {
	if statusCode == http.StatusNoContent || v == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	respBody, err := json.Marshal(v)
	if err != nil {
		RespondError(ctx, w, ErrInternalServer.Wrap(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, _ = w.Write(respBody)
}

func RespondError(ctx context.Context, w http.ResponseWriter, err error) {
	var respErr Error
	if !errors.As(err, &respErr) {
		RespondError(ctx, w, ErrInternalServer.Wrap(err))
		return
	}

	logger := log.FromContext(ctx)
	if respErr.StatusCode == http.StatusInternalServerError {
		logger.WithError(err).Error("Response error")
	} else {
		logger.WithError(err).Debug("Response error")
	}

	respBody, err := json.Marshal(respErr)
	if err != nil {
		logger.WithError(err).Error("Failed to marshal response body")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(respErr.StatusCode)
	_, _ = w.Write(respBody)
}
