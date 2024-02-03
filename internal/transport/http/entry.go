package http

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/Chatyx/backend/internal/config"
	"github.com/Chatyx/backend/pkg/httputil"
	"github.com/Chatyx/backend/pkg/httputil/middleware"
	"github.com/Chatyx/backend/pkg/log"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
	httpSwagger "github.com/swaggo/http-swagger"

	// Register swagger.
	_ "github.com/Chatyx/backend/api"
)

type Controller interface {
	Register(mux *httprouter.Router)
}

type Config struct {
	config.Server
	Debug bool
	Cors  config.Cors
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
func NewServer(conf Config, cs ...Controller) *httputil.Server {
	mux := httprouter.New()
	mux.PanicHandler = func(w http.ResponseWriter, req *http.Request, i interface{}) {
		log.FromContext(req.Context()).Debug(string(debug.Stack()))

		err := fmt.Errorf("panic caught: %v", i)
		httputil.RespondError(req.Context(), w, httputil.ErrInternalServer.Wrap(err))
	}
	mux.GlobalOPTIONS = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	mux.HandlerFunc(http.MethodGet, "/swagger/:any", httpSwagger.Handler())
	mux.Handler(http.MethodGet, "/", http.RedirectHandler("/swagger/index.html", http.StatusMovedPermanently))
	mux.Handler(http.MethodGet, "/swagger", http.RedirectHandler("/swagger/index.html", http.StatusMovedPermanently))

	corsObj := cors.New(cors.Options{
		AllowedOrigins: conf.Cors.AllowedOrigins,
		AllowedMethods: []string{
			http.MethodHead, http.MethodGet, http.MethodPost,
			http.MethodPut, http.MethodPatch, http.MethodDelete,
		},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Fingerprint"},
		MaxAge:           int(conf.Cors.MaxAge.Seconds()),
		AllowCredentials: true,
		Debug:            conf.Debug,
	})

	for _, c := range cs {
		c.Register(mux)
	}

	return httputil.NewServer(conf.Server,
		middleware.Chain(
			mux,
			middleware.RequestID,
			middleware.Log,
			corsObj.Handler,
		),
	)
}
