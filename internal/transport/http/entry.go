package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
)

const defaultShutdownTimeout = 15 * time.Second

type Controller interface {
	Register(mux *httprouter.Router)
}

type Server struct {
	srv *http.Server
}

func NewServer() *Server {
	mux := httprouter.New()

	return &Server{
		srv: &http.Server{
			Addr:                         ":8080",
			Handler:                      mux,
			DisableGeneralOptionsHandler: false,
			TLSConfig:                    nil,
			ReadTimeout:                  0,
			ReadHeaderTimeout:            0,
			WriteTimeout:                 0,
			IdleTimeout:                  0,
			MaxHeaderBytes:               0,
			TLSNextProto:                 nil,
			ConnState:                    nil,
			ErrorLog:                     nil,
			BaseContext:                  nil,
			ConnContext:                  nil,
		},
	}
}

func (s *Server) Run() {

}

func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
	defer cancel()

	if err := s.srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	return nil
}
