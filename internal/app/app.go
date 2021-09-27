package app

import (
	"context"
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

	"github.com/Mort4lis/scht-backend/internal/config"
	handlers "github.com/Mort4lis/scht-backend/internal/delivery/http"
	"github.com/Mort4lis/scht-backend/internal/repository"
	"github.com/Mort4lis/scht-backend/internal/service"
	"github.com/Mort4lis/scht-backend/pkg/auth"
	password "github.com/Mort4lis/scht-backend/pkg/hasher"
	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/Mort4lis/scht-backend/pkg/validator"
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v4/pgxpool"
)

type App struct {
	cfg         *config.Config
	server      *http.Server
	dbPool      *pgxpool.Pool
	redisClient *redis.Client

	logger logging.Logger
}

func NewApp(cfg *config.Config) *App {
	logger := logging.GetLogger()

	dbPool, err := initPG(cfg.Postgres)
	if err != nil {
		logger.WithError(err).Fatal("Unable to connect to postgres")
	}

	redisClient, err := initRedis(cfg.Redis)
	if err != nil {
		logger.WithError(err).Fatal("Unable to connect to redis")
	}

	hasher := password.BCryptPasswordHasher{}

	tokenManager, err := auth.NewTokenManager(cfg.Auth.SignKey)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create new token manager")
	}

	validate, err := validator.New()
	if err != nil {
		logger.WithError(err).Fatal("Failed to init validator")
	}

	userRepo := repository.NewUserPostgresRepository(dbPool)
	sessionRepo := repository.NewSessionRedisRepository(redisClient)
	userService := service.NewUserService(userRepo, hasher)
	authService := service.NewAuthService(service.AuthServiceConfig{
		UserService:     userService,
		SessionRepo:     sessionRepo,
		Hasher:          hasher,
		TokenManager:    tokenManager,
		AccessTokenTTL:  cfg.Auth.AccessTokenTTL,
		RefreshTokenTTL: cfg.Auth.RefreshTokenTTL,
	})
	container := service.ServiceContainer{
		User: userService,
		Auth: authService,
	}

	return &App{
		cfg:         cfg,
		logger:      logger,
		dbPool:      dbPool,
		redisClient: redisClient,
		server: &http.Server{
			Handler:      handlers.Init(container, cfg, validate),
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
		},
	}
}

func initPG(cfg config.PostgresConfig) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	dsn := fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s",
		cfg.Username, cfg.Password,
		cfg.Host, cfg.Port, cfg.Database,
	)

	pool, err := pgxpool.Connect(ctx, dsn)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

func initRedis(cfg config.RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       0, // use default database
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return rdb, nil
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

	app.dbPool.Close()

	return app.server.Close()
}
