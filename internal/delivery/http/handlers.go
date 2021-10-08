package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/rs/cors"

	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/Mort4lis/scht-backend/docs"
	"github.com/Mort4lis/scht-backend/internal/config"
	"github.com/Mort4lis/scht-backend/internal/service"
	"github.com/Mort4lis/scht-backend/internal/utils"
	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/go-playground/validator/v10"
	"github.com/julienschmidt/httprouter"
)

type baseHandler struct {
	logger   logging.Logger
	validate *validator.Validate
}

func (h *baseHandler) decodeJSONFromBody(body io.ReadCloser, decoder utils.JSONDecoder) error {
	if err := decoder.DecodeFrom(body); err != nil {
		h.logger.WithError(err).Debug("Invalid json body")
		return errInvalidJSON
	}

	defer func() {
		if err := body.Close(); err != nil {
			h.logger.WithError(err).Error("An error occurred while closing body")
		}
	}()

	return nil
}

func (h *baseHandler) validateStruct(s interface{}) error {
	if err := h.validate.Struct(s); err != nil {
		fields := ErrorFields{}
		for _, err := range err.(validator.ValidationErrors) {
			fields[err.Field()] = fmt.Sprintf(
				"field validation for '%s' failed on the '%s' tag",
				err.Field(), err.Tag(),
			)
		}

		h.logger.Debugf("fields validation failed: %v", fields)

		return ResponseError{
			StatusCode: http.StatusBadRequest,
			Message:    "validation error",
			Fields:     fields,
		}
	}

	return nil
}

func extractTokenFromHeader(header string) (string, error) {
	if header == "" {
		logging.GetLogger().Debug("authorization header is empty")
		return "", errInvalidAuthorizationHeader
	}

	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 {
		logging.GetLogger().Debug("authorization header must contains with two parts")
		return "", errInvalidAuthorizationHeader
	}

	if headerParts[0] != "Bearer" {
		logging.GetLogger().Debug("authorization header doesn't begin with Bearer")
		return "", errInvalidAuthorizationHeader
	}

	if headerParts[1] == "" {
		logging.GetLogger().Debug("authorization header value is empty")
		return "", errInvalidAuthorizationHeader
	}

	return headerParts[1], nil
}

func respondSuccess(statusCode int, w http.ResponseWriter, encoder utils.JSONEncoder) {
	if encoder == nil {
		w.WriteHeader(statusCode)
		return
	}

	respBody, err := encoder.Encode()
	if err != nil {
		logging.GetLogger().WithError(err).Error("An error occurred while encoding response structure")
		respondError(w, errInternalServer)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if _, err = w.Write(respBody); err != nil {
		logging.GetLogger().WithError(err).Error("An error occurred while writing response body")
		return
	}
}

func respondError(w http.ResponseWriter, err error) {
	appErr, ok := err.(ResponseError)
	if !ok {
		respondError(w, errInternalServer)
		return
	}

	respBody, err := json.Marshal(appErr)
	if err != nil {
		logging.GetLogger().WithError(err).Error("An error occurred while marshalling application error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.StatusCode)

	if _, err = w.Write(respBody); err != nil {
		logging.GetLogger().WithError(err).Error("An error occurred while writing response body")
	}
}

func Init(container service.ServiceContainer, cfg *config.Config, validate *validator.Validate) http.Handler {
	router := httprouter.New()

	newUserHandler(container.User, container.Auth, validate).register(router)
	newAuthHandler(container.Auth, validate, cfg.Domain, cfg.Auth.RefreshTokenTTL).register(router)

	router.HandlerFunc(http.MethodGet, "/docs/:any", httpSwagger.WrapHandler)

	router.PanicHandler = func(w http.ResponseWriter, req *http.Request, i interface{}) {
		logging.GetLogger().Errorf("There was a panic: %v", i)
		respondError(w, errInternalServer)
	}
	router.GlobalOPTIONS = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	corsHandler := cors.New(cors.Options{
		AllowedOrigins: cfg.Cors.AllowedOrigins,
		AllowedMethods: []string{
			http.MethodHead, http.MethodGet, http.MethodPost,
			http.MethodPut, http.MethodPatch, http.MethodDelete,
		},
		AllowedHeaders:   []string{"X-Fingerprint"},
		MaxAge:           cfg.Cors.MaxAge,
		AllowCredentials: true,
		Debug:            cfg.IsDebug,
	})

	return loggingMiddleware(corsHandler.Handler(router))
}
