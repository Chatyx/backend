package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Chatyx/backend/internal/config"
	"github.com/Chatyx/backend/pkg/httputil/middleware"
	"github.com/Chatyx/backend/pkg/log"

	"github.com/julienschmidt/httprouter"
)

const defaultShutdownTimeout = 15 * time.Second

type Controller interface {
	Register(mux *httprouter.Router)
}

type Server struct {
	srv *http.Server
}

// NewServer creates a new http server
//
//	@title						Chatyx REST API
//	@version					1.0
//	@description				REST API for Chatyx backend application
//
//	@contact.name				Pavel Korchagin
//	@contact.email				mortalis94@gmail.com
//
//	@license.name				MIT
//	@license.url				https://opensource.org/license/mit/
//
//	@host						localhost:8080
//	@BasePath					/api/v1
//
//	@securityDefinitions.apikey	JWTAuth
//	@in							header
//	@name						Authorization
func NewServer(conf config.Server, cs ...Controller) *Server {
	mux := httprouter.New()
	for _, c := range cs {
		c.Register(mux)
	}

	return &Server{
		srv: &http.Server{
			Addr: conf.Listen,
			Handler: middleware.Chain(
				mux,
				middleware.RequestID,
				middleware.Log,
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
