package httputil

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Chatyx/backend/internal/config"
	"github.com/Chatyx/backend/pkg/log"
)

const defaultShutdownTimeout = 15 * time.Second

type Server struct {
	srv *http.Server
}

func NewServer(conf config.Server, h http.Handler) *Server {
	return &Server{
		srv: &http.Server{
			Addr:         conf.Listen,
			Handler:      h,
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

	log.Infof("HTTP server successfully started! Listen %s", s.srv.Addr)
}

func (s *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
	defer cancel()

	if err := s.srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	return nil
}
