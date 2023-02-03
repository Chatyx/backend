package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	_ "github.com/Mort4lis/scht-backend/docs"
	"github.com/Mort4lis/scht-backend/internal/config"
	"github.com/Mort4lis/scht-backend/internal/encoding"
	"github.com/Mort4lis/scht-backend/internal/service"
	pkgErrors "github.com/Mort4lis/scht-backend/pkg/errors"
	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/Mort4lis/scht-backend/pkg/validator"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
	httpSwagger "github.com/swaggo/http-swagger"
)

type baseHandler struct {
	logger logging.Logger
}

func (h *baseHandler) decodeBody(body io.ReadCloser, unmarshaler encoding.Unmarshaler) error {
	defer body.Close()

	payload, err := ioutil.ReadAll(body)
	if err != nil {
		return fmt.Errorf("an error occurred while reading request body: %w", err)
	}

	if err = unmarshaler.Unmarshal(payload); err != nil {
		return errInvalidDecodeBody.Wrap(fmt.Errorf("failed to unmarshal request body: %w", err))
	}

	return nil
}

func (h *baseHandler) validate(vld validator.Validator) error {
	if err := vld.Validate(); err != nil {
		if vldErr, ok := err.(validator.ValidationError); ok {
			return errValidationError.Wrap(err, WithFields(ErrorFields(vldErr.Fields)))
		}

		return errInternalServer.Wrap(err)
	}

	return nil
}

func extractTokenFromHeader(header string) (string, error) {
	if header == "" {
		return "", errInvalidAuthorizationHeader.Wrap(fmt.Errorf("authorization header is empty"))
	}

	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 {
		return "", errInvalidAuthorizationHeader.Wrap(fmt.Errorf("authorization header must contains with two parts"))
	}

	if headerParts[0] != "Bearer" {
		return "", errInvalidAuthorizationHeader.Wrap(fmt.Errorf("authorization header doesn't begin with Bearer"))
	}

	if headerParts[1] == "" {
		return "", errInvalidAuthorizationHeader.Wrap(fmt.Errorf("authorization header value is empty"))
	}

	return headerParts[1], nil
}

func respondSuccess(ctx context.Context, statusCode int, w http.ResponseWriter, marshaler encoding.Marshaler) {
	if marshaler == nil {
		w.WriteHeader(statusCode)
		return
	}

	respBody, err := marshaler.Marshal()
	if err != nil {
		respondError(ctx, w, fmt.Errorf("an error occurred while marshaling response structure: %v", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if _, err = w.Write(respBody); err != nil {
		respondError(ctx, w, fmt.Errorf("an error occurred while writing response body: %v", err))
		return
	}
}

func respondError(ctx context.Context, w http.ResponseWriter, err error) {
	respErr, ok := err.(ResponseError)
	if !ok {
		respondError(ctx, w, errInternalServer.Wrap(err))
		return
	}

	ctxFields := pkgErrors.UnwrapContextFields(err)
	logger := logging.GetLoggerFromContext(ctx).WithFields(logging.Fields(ctxFields))

	switch respErr.StatusCode {
	case http.StatusInternalServerError:
		logger.WithError(err).Error("response error")
	default:
		logger.WithError(err).Debug("response error")
	}

	respBody, err := json.Marshal(respErr)
	if err != nil {
		logger.WithError(err).Error("An error occurred while marshaling application error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(respErr.StatusCode)

	if _, err = w.Write(respBody); err != nil {
		logger.WithError(err).Error("An error occurred while writing response body")
	}
}

func Init(container service.ServiceContainer, cfg *config.Config) http.Handler {
	router := httprouter.New()
	authMid := AuthorizationMiddlewareFactory(container.Auth)

	newUserHandler(container.User).register(router, authMid)
	newChatHandler(container.Chat).register(router, authMid)
	newChatMemberHandler(container.ChatMember).register(router, authMid)
	newMessageHandler(container.Message).register(router, authMid)
	newAuthHandler(container.Auth, cfg.Domain, cfg.Auth.RefreshTokenTTL).register(router)

	router.HandlerFunc(http.MethodGet, "/docs/:any", httpSwagger.WrapHandler)
	router.Handler(http.MethodGet, "/", http.RedirectHandler("/docs/index.html", http.StatusMovedPermanently))
	router.Handler(http.MethodGet, "/docs", http.RedirectHandler("/docs/index.html", http.StatusMovedPermanently))

	router.PanicHandler = func(w http.ResponseWriter, req *http.Request, i interface{}) {
		logging.GetLogger().Errorf("There was a panic: %v", i)
		respondError(req.Context(), w, errInternalServer)
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
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Fingerprint"},
		MaxAge:           cfg.Cors.MaxAge,
		AllowCredentials: true,
		Debug:            cfg.IsDebug,
	})

	return loggingMiddleware(corsHandler.Handler(router))
}
