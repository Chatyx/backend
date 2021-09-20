package app

import (
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type App struct {
	server *http.Server
}

func NewApp() *App {
	mux := http.NewServeMux()

	return &App{
		server: &http.Server{
			Handler:      mux,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
		},
	}
}

func (app *App) Run() error {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		return err
	}

	go func() {
		if err = app.server.Serve(listener); err != nil {
			switch {
			case errors.Is(err, http.ErrServerClosed):
				log.Println("Server shutdown")
			default:
				log.Fatal(err)
			}
		}
	}()

	return app.gracefulShutdown()
}

func (app *App) gracefulShutdown() error {
	quit := make(chan os.Signal, 1)
	signal.Notify(
		quit,
		[]os.Signal{syscall.SIGABRT, syscall.SIGQUIT, syscall.SIGHUP, syscall.SIGTERM, os.Interrupt}...,
	)

	sig := <-quit
	log.Printf("Caught signal %s. Shutting down...", sig)

	return app.server.Close()
}
