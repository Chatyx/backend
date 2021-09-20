package app

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Mort4lis/scht-backend/internal/config"
)

type App struct {
	cfg    *config.Config
	server *http.Server
}

func NewApp(cfg *config.Config) *App {
	mux := http.NewServeMux()

	return &App{
		cfg: cfg,
		server: &http.Server{
			Handler:      mux,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
		},
	}
}

func (app *App) Run() error {
	var (
		lis    net.Listener
		lisErr error
	)

	lisType := app.cfg.Listen.Type
	switch lisType {
	case "sock":
		appDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			return fmt.Errorf("failed to get the application directory due %v", err)
		}

		sockPath := path.Join(appDir, "scht.sock")

		log.Printf("bind application to unix socket %s", sockPath)
		lis, lisErr = net.Listen("unix", sockPath)
	case "port":
		ip, port := app.cfg.Listen.BindIP, app.cfg.Listen.BindPort
		addr := fmt.Sprintf("%s:%d", ip, port)

		log.Printf("bind application to tcp %s", addr)
		lis, lisErr = net.Listen("tcp", addr)
	default:
		return fmt.Errorf("unsupport listen type %q", lisType)
	}

	if lisErr != nil {
		return fmt.Errorf("failed to listen %s due %v", lisType, lisErr)
	}

	go func() {
		if err := app.server.Serve(lis); err != nil {
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
