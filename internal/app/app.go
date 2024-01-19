package app

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/Chatyx/backend/internal/config"
	"github.com/Chatyx/backend/internal/infrastructure/repository/postgres"
	"github.com/Chatyx/backend/internal/service"
	inhttp "github.com/Chatyx/backend/internal/transport/http"
	v1 "github.com/Chatyx/backend/internal/transport/http/v1"
	"github.com/Chatyx/backend/pkg/auth"
	"github.com/Chatyx/backend/pkg/auth/storage/redis"
	auhttp "github.com/Chatyx/backend/pkg/auth/transport/http"
	"github.com/Chatyx/backend/pkg/httputil/middleware"
	"github.com/Chatyx/backend/pkg/log"
	"github.com/Chatyx/backend/pkg/validator"

	"github.com/ilyakaznacheev/cleanenv"
)

type CloserAdapter func()

func (c CloserAdapter) Close() error {
	c()
	return nil
}

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

	pgPool, err := postgres.NewConnPool(conf.Postgres)
	if err != nil {
		log.WithError(err).Fatal("Failed to init postgres connection pool")
	}
	closers = append(closers, CloserAdapter(pgPool.Close))

	userRepo := postgres.NewUserRepository(pgPool)

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

	userService := service.NewUser(service.UserConfig{
		UserRepository:    userRepo,
		SessionRepository: authStorage,
	})
	authService := auth.NewService(
		authStorage,
		auth.WithIssuer(conf.Auth.Issuer),
		auth.WithSignedKey([]byte(conf.Auth.SignKey)),
		auth.WithAccessTokenTTL(conf.Auth.AccessTokenTTL),
		auth.WithRefreshTokenTTL(conf.Auth.RefreshTokenTTL),
		auth.WithLogger(log.With("service", "auth")),
		auth.WithCheckPassword(userService.CheckPassword),
	)

	authorizeMiddleware := middleware.Authorize([]byte(conf.Auth.SignKey))

	vld := validator.NewValidator()
	userController := v1.NewUserController(v1.UserControllerConfig{
		Service:   userService,
		Authorize: authorizeMiddleware,
		Validator: vld,
	})
	authController := auhttp.NewController(
		authService, vld,
		auhttp.WithPrefixPath("/api/v1"),
		auhttp.WithRTCookieDomain(conf.Domain),
		auhttp.WithRTCookieTTL(conf.Auth.RefreshTokenTTL),
	)

	apiServer := inhttp.NewServer(
		inhttp.Config{
			Server: conf.API,
			Debug:  conf.Debug,
			Cors:   conf.Cors,
		},
		authController,
		userController,
	)
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
