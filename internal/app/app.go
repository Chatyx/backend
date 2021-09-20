package app

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"
	"time"

	"github.com/Mort4lis/scht-backend/pkg/logging"

	"github.com/Mort4lis/scht-backend/internal/config"
)

type App struct {
	cfg    *config.Config
	server *http.Server

	logger logging.Logger
}

func NewApp(cfg *config.Config) *App {
	logger := logging.GetLogger()
	mux := http.NewServeMux()

	return &App{
		cfg:    cfg,
		logger: logger,
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

	logger := app.logger

	lisType := app.cfg.Listen.Type
	switch lisType {
	case "sock":
		appDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			logger.WithError(err).Error("Failed to get the application directory")

			return fmt.Errorf("failed to get the application directory")
		}

		sockPath := path.Join(appDir, "scht.sock")

		logger.Infof("Bind application to unix socket %s", sockPath)
		lis, lisErr = net.Listen("unix", sockPath)
	case "port":
		ip, port := app.cfg.Listen.BindIP, app.cfg.Listen.BindPort
		addr := fmt.Sprintf("%s:%d", ip, port)

		logger.Infof("Bind application to tcp %s", addr)
		lis, lisErr = net.Listen("tcp", addr)
	default:
		return fmt.Errorf("unsupport listen type %q", lisType)
	}

	if lisErr != nil {
		logger.WithError(lisErr).Errorf("Failed to listen %s", lisType)

		return fmt.Errorf("failed to listen %s", lisType)
	}

	go func() {
		if err := app.server.Serve(lis); err != nil {
			switch {
			case errors.Is(err, http.ErrServerClosed):
				logger.Info("Server shutdown")
			default:
				logger.Fatal(err)
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
	app.logger.Infof("Caught signal %s. Shutting down...", sig)

	return app.server.Close()
}
