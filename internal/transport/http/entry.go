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
