package websocket

import (
	"net/http"

	"github.com/Mort4lis/scht-backend/internal/config"
	"github.com/rs/cors"
)

func Init(cfg *config.Config) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/", &chatSessionHandler{})

	corsHandler := cors.New(cors.Options{
		AllowedOrigins: cfg.Cors.AllowedOrigins,
		MaxAge:         cfg.Cors.MaxAge,
		Debug:          cfg.IsDebug,
	})

	return corsHandler.Handler(mux)
}
