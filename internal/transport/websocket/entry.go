package websocket

import (
	"net/http"

	"github.com/Chatyx/backend/internal/config"
	"github.com/Chatyx/backend/pkg/httputil"
	"github.com/Chatyx/backend/pkg/httputil/middleware"

	"github.com/rs/cors"
)

type Config struct {
	config.Server
	Debug bool
	Cors  config.Cors
}

func NewServer(conf Config, h http.Handler) *httputil.Server {
	mux := http.NewServeMux()
	mux.Handle("/", h)

	corsObj := cors.New(cors.Options{
		AllowedOrigins: conf.Cors.AllowedOrigins,
		AllowedHeaders: []string{"Authorization"},
		MaxAge:         int(conf.Cors.MaxAge.Seconds()),
		Debug:          conf.Debug,
	})

	return httputil.NewServer(conf.Server,
		middleware.Chain(
			mux,
			corsObj.Handler,
		),
	)
}
