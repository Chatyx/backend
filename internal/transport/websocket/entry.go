package websocket

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Chatyx/backend/internal/config"
	"github.com/Chatyx/backend/pkg/httputil/middleware"
	"github.com/Chatyx/backend/pkg/log"

	"github.com/rs/cors"
)

const defaultShutdownTimeout = 15 * time.Second

type Config struct {
	config.Server
	Debug bool
	Cors  config.Cors
}

type Server struct {
	srv *http.Server
}

func NewServer(conf Config, h http.Handler) *Server {
	mux := http.NewServeMux()
	mux.Handle("/", h)

	corsObj := cors.New(cors.Options{
		AllowedOrigins: conf.Cors.AllowedOrigins,
		AllowedHeaders: []string{"Authorization"},
		MaxAge:         int(conf.Cors.MaxAge.Seconds()),
		Debug:          conf.Debug,
	})

	return &Server{
		srv: &http.Server{
			Addr: conf.Listen,
			Handler: middleware.Chain(
				mux,
				corsObj.Handler,
			),
			ReadTimeout:  conf.ReadTimeout,
			WriteTimeout: conf.WriteTimeout,
		},
	}
}

func (s *Server) Run() {
	go func() {
		if err := s.srv.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.WithError(err).Fatalf("Failed listen and serve %s", s.srv.Addr)
			}
		}
	}()

	log.Infof("Server successfully started! Listen %s", s.srv.Addr)
}

func (s *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
	defer cancel()

	if err := s.srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	return nil
}
