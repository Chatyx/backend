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
	inVld "github.com/Mort4lis/scht-backend/internal/validator"
	"github.com/Mort4lis/scht-backend/pkg/auth"
	password "github.com/Mort4lis/scht-backend/pkg/hasher"
	"github.com/Mort4lis/scht-backend/pkg/logging"
	pkgVld "github.com/Mort4lis/scht-backend/pkg/validator"
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
	app := &App{
		cfg:    cfg,
		logger: logging.GetLogger(),
	}

	app.initPG(cfg.Postgres)
	app.initRedis(cfg.Redis)

	hasher := password.BCryptPasswordHasher{}

	tokenManager, err := auth.NewJWTTokenManager(cfg.Auth.SignKey)
	if err != nil {
		app.logger.WithError(err).Fatal("Failed to create new token manager")
	}

	validate, err := inVld.New()
	if err != nil {
		app.logger.WithError(err).Fatal("Failed to init validator")
	}

	pkgVld.SetValidate(validate)

	msgPubSub := repository.NewMessagePubSub(app.redisClient)
	userRepo := repository.NewUserPostgresRepository(app.dbPool)
	chatRepo := repository.NewChatCacheRepositoryDecorator(
		repository.NewChatPostgresRepository(app.dbPool),
		app.redisClient,
	)
	chatMemberRepo := repository.NewChatMemberCacheRepository(
		repository.NewChatMemberPostgresRepository(app.dbPool),
		app.redisClient,
	)
	sessionRepo := repository.NewSessionRedisRepository(app.redisClient)
	msgRepo := repository.NewMessageRedisRepository(app.redisClient)

	userService := service.NewUserService(userRepo, sessionRepo, hasher)
	chatService := service.NewChatService(chatRepo)
	chatMemberService := service.NewChatMemberService(service.ChatMemberConfig{
		UserService:    userService,
		ChatMemberRepo: chatMemberRepo,
		MessageRepo:    msgRepo,
		MessagePubSub:  msgPubSub,
	})
	messageService := service.NewMessageService(chatMemberRepo, msgRepo, msgPubSub)
	authService := service.NewAuthService(service.AuthServiceConfig{
		UserService:     userService,
		SessionRepo:     sessionRepo,
		Hasher:          hasher,
		TokenManager:    tokenManager,
		AccessTokenTTL:  cfg.Auth.AccessTokenTTL,
		RefreshTokenTTL: cfg.Auth.RefreshTokenTTL,
	})

	container := service.ServiceContainer{
		User:       userService,
		Chat:       chatService,
		ChatMember: chatMemberService,
		Message:    messageService,
		Auth:       authService,
	}

	apiListenCfg := cfg.Listen.API
	app.apiServer = server.NewHttpServerWrapper(
		server.Config{
			ServerName: "API Server",
			ListenType: apiListenCfg.Type,
			BindIP:     apiListenCfg.BindIP,
			BindPort:   apiListenCfg.BindPort,
		},
		&http.Server{
			Handler:      apiHandlers.Init(container, cfg),
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
		},
	)

	chatListenCfg := cfg.Listen.Chat
	app.chatServer = server.NewHttpServerWrapper(
		server.Config{
			ServerName: "Chat Server",
			ListenType: chatListenCfg.Type,
			BindIP:     chatListenCfg.BindIP,
			BindPort:   chatListenCfg.BindPort,
		},
		&http.Server{
			Handler:      chatHandlers.Init(container, cfg),
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		},
	)

	return app
}

func (app *App) initPG(cfg config.PostgresConfig) {
	var (
		err  error
		pool *pgxpool.Pool
	)

	dsn := fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s",
		cfg.Username, cfg.Password,
		cfg.Host, cfg.Port, cfg.Database,
	)

	err = DoWithAttempts(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		pool, err = pgxpool.Connect(ctx, dsn)
		if err != nil {
			app.logger.WithError(err).Warning("Failed to connect to postgres... Going to do the next attempt")

			return err
		}

		return nil
	}, cfg.MaxConnectionAttempts, cfg.FailedConnectionDelay)

	if err != nil {
		app.logger.WithError(err).Fatal("All attempts are exceeded. Unable to connect to postgres")
	}

	app.dbPool = pool
}

func (app *App) initRedis(cfg config.RedisConfig) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Username: cfg.Username,
		Password: cfg.Password,
		DB:       0, // use default database
	})

	err := DoWithAttempts(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := rdb.Ping(ctx).Err(); err != nil {
			app.logger.WithError(err).Warning("Failed to connect to redis... Going to do the next attempt")

			return err
		}

		return nil
	}, cfg.MaxConnectionAttempts, cfg.FailedConnectionDelay)
	if err != nil {
		app.logger.WithError(err).Fatal("All attempts are exceeded. Unable to connect to redis")
	}

	app.redisClient = rdb
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

func DoWithAttempts(fn func() error, maxAttempts int, delay time.Duration) error {
	var err error

	for maxAttempts > 0 {
		if err = fn(); err != nil {
			time.Sleep(delay)
			maxAttempts--

			continue
		}

		return nil
	}

	return err
}
