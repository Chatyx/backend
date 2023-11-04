package http

import (
	"context"
	"net/http"
	"time"

	"github.com/Chatyx/backend/pkg/usersrv/transport/http/v1"

	"github.com/julienschmidt/httprouter"
)

type Controller interface {
	Register(mux *httprouter.Router)
}

type Server struct {
	srv *http.Server
}

func NewServer() *Server {
	mux := httprouter.New()
	(&v1.UserController{}).Register(mux)

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
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	return s.srv.Shutdown(ctx)
}
