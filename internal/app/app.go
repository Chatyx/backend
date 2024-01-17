package app

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/Chatyx/backend/internal/config"
	inhttp "github.com/Chatyx/backend/internal/transport/http"
	"github.com/Chatyx/backend/pkg/auth"
	"github.com/Chatyx/backend/pkg/auth/storage/redis"
	auhttp "github.com/Chatyx/backend/pkg/auth/transport/http"
	"github.com/Chatyx/backend/pkg/httputil/middleware"
	"github.com/Chatyx/backend/pkg/log"
	"github.com/Chatyx/backend/pkg/validator"

	"github.com/ilyakaznacheev/cleanenv"
)

type Runner interface {
	Run()
}

type App struct {
	runners []Runner
	closers []io.Closer
}

func NewApp(confPath string) *App {
	var (
		conf    config.Config
		runners []Runner
		closers []io.Closer
	)

	if err := cleanenv.ReadConfig(confPath, &conf); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read config: %v", err)
		os.Exit(1)
	}

	if err := log.Configure(log.Config{
		Level:          conf.Log.Level,
		ProductionMode: !conf.Debug,
	}); err != nil {
		log.WithError(err).Fatal("Failed to configure logger")
	}

	vld := validator.NewValidator()

	authStorageDBNum, _ := strconv.Atoi(conf.Redis.Database)
	authStorage, err := redis.NewStorage(redis.Config{
		Host:        conf.Redis.Host,
		Port:        conf.Redis.Port,
		Username:    conf.Redis.User,
		Password:    conf.Redis.Password,
		DB:          authStorageDBNum,
		ConnTimeout: conf.Redis.Timeout,
	})
	if err != nil {
		log.WithError(err).Fatal("Failed to establish redis connection")
	}
	closers = append(closers, authStorage)

	authService := auth.NewService(
		authStorage,
		auth.WithIssuer(conf.Auth.Issuer),
		auth.WithSignedKey([]byte(conf.Auth.SignKey)),
		auth.WithAccessTokenTTL(conf.Auth.AccessTokenTTL),
		auth.WithRefreshTokenTTL(conf.Auth.RefreshTokenTTL),
		auth.WithLogger(log.With("service", "auth")),
		// auth.WithCheckPassword(), TODO
	)

	authorizeMiddleware := middleware.Authorize([]byte(conf.Auth.SignKey))
	_ = authorizeMiddleware
	authController := auhttp.NewController(
		authService, vld,
		auhttp.WithPrefixPath("/api/v1"),
		auhttp.WithRTCookieDomain(conf.Domain),
		auhttp.WithRTCookieTTL(conf.Auth.RefreshTokenTTL),
	)

	apiServer := inhttp.NewServer(conf.API, authController)
	runners = append(runners, apiServer)
	closers = append(closers, apiServer)

	return &App{
		runners: runners,
		closers: closers,
	}
}

func (a *App) Run() {
	for _, runner := range a.runners {
		runner.Run()
	}

	a.gracefulShutdown()
}

func (a *App) gracefulShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, os.Interrupt)

	log.Infof("Caught signal %s. Shutting down...", <-quit)

	for i := len(a.closers) - 1; i >= 0; i-- {
		if err := a.closers[i].Close(); err != nil {
			log.WithError(err).Error("Failed to close")
		}
	}
}
