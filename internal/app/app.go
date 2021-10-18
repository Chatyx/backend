package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Mort4lis/scht-backend/internal/config"
	apiHandlers "github.com/Mort4lis/scht-backend/internal/delivery/http"
	chatHandlers "github.com/Mort4lis/scht-backend/internal/delivery/websocket"
	"github.com/Mort4lis/scht-backend/internal/repository"
	"github.com/Mort4lis/scht-backend/internal/server"
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
	dbPool      *pgxpool.Pool
	redisClient *redis.Client

	apiServer  server.Server
	chatServer server.Server

	logger logging.Logger
}

// @title Scht REST API
// @version 1.0
// @description REST API for Scht Backend application

// @host 127.0.0.1
// @BasePath /api

// @contact.name Pavel Korchagin
// @contact.email mortalis94@gmail.com

// @securityDefinitions.apikey JWTTokenAuth
// @in header
// @name Authorization

// NewApp creates new application.
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

	tokenManager, err := auth.NewJWTTokenManager(cfg.Auth.SignKey)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create new token manager")
	}

	validate, err := validator.New()
	if err != nil {
		logger.WithError(err).Fatal("Failed to init validator")
	}

	msgPubSub := repository.NewMessagePubSub(redisClient)
	msgRepo := repository.NewMessageRedisRepository(redisClient)
	userRepo := repository.NewUserPostgresRepository(dbPool)
	chatRepo := repository.NewChatPostgresRepository(dbPool)
	sessionRepo := repository.NewSessionRedisRepository(redisClient)

	userService := service.NewUserService(userRepo, sessionRepo, hasher)
	chatService := service.NewChatService(chatRepo)
	messageService := service.NewMessageService(chatService, msgRepo, msgPubSub)
	authService := service.NewAuthService(service.AuthServiceConfig{
		UserService:     userService,
		SessionRepo:     sessionRepo,
		Hasher:          hasher,
		TokenManager:    tokenManager,
		AccessTokenTTL:  cfg.Auth.AccessTokenTTL,
		RefreshTokenTTL: cfg.Auth.RefreshTokenTTL,
	})

	container := service.ServiceContainer{
		User:    userService,
		Chat:    chatService,
		Message: messageService,
		Auth:    authService,
	}

	apiListenCfg := cfg.Listen.API
	apiServer := server.NewHttpServerWrapper(
		server.Config{
			ServerName: "API Server",
			ListenType: apiListenCfg.Type,
			BindIP:     apiListenCfg.BindIP,
			BindPort:   apiListenCfg.BindPort,
		},
		&http.Server{
			Handler:      apiHandlers.Init(container, cfg, validate),
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
		},
	)

	chatListenCfg := cfg.Listen.Chat
	chatServer := server.NewHttpServerWrapper(
		server.Config{
			ServerName: "Chat Server",
			ListenType: chatListenCfg.Type,
			BindIP:     chatListenCfg.BindIP,
			BindPort:   chatListenCfg.BindPort,
		},
		&http.Server{
			Handler:      chatHandlers.Init(container, cfg, validate),
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		},
	)

	return &App{
		cfg:         cfg,
		logger:      logger,
		dbPool:      dbPool,
		redisClient: redisClient,
		apiServer:   apiServer,
		chatServer:  chatServer,
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
	if err := app.apiServer.Run(); err != nil {
		return err
	}

	if err := app.chatServer.Run(); err != nil {
		return err
	}

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

	if err := app.chatServer.Stop(); err != nil {
		return err
	}

	return app.apiServer.Stop()
}
