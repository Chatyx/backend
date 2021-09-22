package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Mort4lis/scht-backend/internal/utils"

	"github.com/Mort4lis/scht-backend/internal/domain"
	"github.com/Mort4lis/scht-backend/pkg/logging"

	"github.com/go-playground/validator/v10"

	"github.com/Mort4lis/scht-backend/internal/services"
	"github.com/julienschmidt/httprouter"
)

type Handler struct {
	logger   logging.Logger
	validate *validator.Validate
}

func (h *Handler) DecodeJSONFromBody(body io.ReadCloser, decoder utils.JSONDecoder) error {
	if err := decoder.DecodeFrom(body); err != nil {
		h.logger.WithError(err).Debug("Invalid json body")

		return domain.ErrInvalidJSON
	}

	defer func() {
		if err := body.Close(); err != nil {
			h.logger.WithError(err).Error("Error occurred while closing body")
		}
	}()

	return nil
}

func (h *Handler) Validate(s interface{}) error {
	if err := h.validate.Struct(s); err != nil {
		fields := domain.ErrorFields{}
		for _, err := range err.(validator.ValidationErrors) {
			fields[err.Field()] = fmt.Sprintf(
				"field validation for '%s' failed on the '%s' tag",
				err.Field(), err.Tag(),
			)
		}

		return domain.NewValidationError(fields)
	}

	return nil
}

func (h *Handler) RespondSuccess(statusCode int, w http.ResponseWriter, encoder utils.JSONEncoder) {
	respBody, err := encoder.Encode()
	if err != nil {
		h.logger.WithError(err).Error("Error occurred while encoding response structure")
		h.RespondError(w, err)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if _, err = w.Write(respBody); err != nil {
		h.logger.WithError(err).Error("Error occurred while writing response body")

		return
	}
}

func (h *Handler) RespondError(w http.ResponseWriter, err error) {
	appErr, ok := err.(domain.AppError)
	if !ok {
		h.RespondError(w, domain.ErrInternalServer)

		return
	}

	respBody, err := json.Marshal(appErr)
	if err != nil {
		h.logger.WithError(err).Error("Error occurred while marshalling application error")

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.StatusCode)

	if _, err = w.Write(respBody); err != nil {
		h.logger.WithError(err).Error("Error occurred while writing response body")
	}
}

func Init(container services.ServiceContainer, validate *validator.Validate) http.Handler {
	router := httprouter.New()

	NewUserHandler(container.User, validate).Register(router)

	return router
}
