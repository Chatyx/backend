package http

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/Mort4lis/scht-backend/internal/encoding"

	"github.com/rs/cors"

	httpSwagger "github.com/swaggo/http-swagger"

	_ "github.com/Mort4lis/scht-backend/docs"
	"github.com/Mort4lis/scht-backend/internal/config"
	"github.com/Mort4lis/scht-backend/internal/service"
	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/go-playground/validator/v10"
	"github.com/julienschmidt/httprouter"
)

type baseHandler struct {
	logger   logging.Logger
	validate *validator.Validate
}

func (h *baseHandler) decodeBody(body io.ReadCloser, unmarshaler encoding.Unmarshaler) error {
	defer body.Close()

	payload, err := ioutil.ReadAll(body)
	if err != nil {
		h.logger.WithError(err).Error("An error occurred while reading body")
		return err
	}

	if err = unmarshaler.Unmarshal(payload); err != nil {
		h.logger.WithError(err).Debug("failed to unmarshal body")
		return errInvalidDecodeBody
	}

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

func respondSuccess(statusCode int, w http.ResponseWriter, marshaler encoding.Marshaler) {
	if marshaler == nil {
		w.WriteHeader(statusCode)
		return
	}

	respBody, err := marshaler.Marshal()
	if err != nil {
		logging.GetLogger().WithError(err).Error("An error occurred while marshaling response structure")
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
		logging.GetLogger().WithError(err).Error("An error occurred while marshaling application error")
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
	authMid := AuthorizationMiddlewareFactory(container.Auth)

	newUserHandler(container.User, validate).register(router, authMid)
	newChatHandler(container.Chat, validate).register(router, authMid)
	newChatMemberHandler(container.ChatMember, validate).register(router, authMid)
	newMessageHandler(container.Message, validate).register(router, authMid)
	newAuthHandler(container.Auth, validate, cfg.Domain, cfg.Auth.RefreshTokenTTL).register(router)

	router.HandlerFunc(http.MethodGet, "/docs/:any", httpSwagger.WrapHandler)
	router.Handler(http.MethodGet, "/", http.RedirectHandler("/docs/index.html", http.StatusMovedPermanently))
	router.Handler(http.MethodGet, "/docs", http.RedirectHandler("/docs/index.html", http.StatusMovedPermanently))

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
