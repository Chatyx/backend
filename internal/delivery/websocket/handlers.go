package websocket

import (
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/Mort4lis/scht-backend/internal/config"
	httpHandlers "github.com/Mort4lis/scht-backend/internal/delivery/http"
	"github.com/Mort4lis/scht-backend/internal/service"
	"github.com/rs/cors"
)

func Init(container service.ServiceContainer, cfg *config.Config, validate *validator.Validate) http.Handler {
	mux := http.NewServeMux()
	authMid := httpHandlers.AuthorizationMiddlewareFactory(container.Auth)

	mux.Handle("/", authMid(newChatSessionHandler(container.Message, validate)))

	corsHandler := cors.New(cors.Options{
		AllowedOrigins: cfg.Cors.AllowedOrigins,
		MaxAge:         cfg.Cors.MaxAge,
		Debug:          cfg.IsDebug,
	})

	return corsHandler.Handler(mux)
}
