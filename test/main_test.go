// +build integration

package test

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/Mort4lis/scht-backend/internal/app"
	"github.com/Mort4lis/scht-backend/internal/config"
	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/go-redis/redis/v8"
	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	ws "github.com/gorilla/websocket"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/stretchr/testify/suite"
)

const (
	configPath     = "../configs/test.yml"
	migrationsPath = "../internal/db/migrations"
	fixturesPath   = "./fixtures"
)

type AppTestSuite struct {
	suite.Suite
	app          *app.App
	cfg          *config.Config
	dbMigration  *migrate.Migrate
	dbConn       *sql.DB
	fixtures     *testfixtures.Loader
	redisClient  *redis.Client
	httpClient   *http.Client
	wsDialer     *ws.Dialer
	apiURLPrefix string
	chatURL      string
}

func TestAppTestSuite(t *testing.T) {
	suite.Run(t, new(AppTestSuite))
}

func (s *AppTestSuite) SetupSuite() {
	cfg := config.GetConfig(configPath)
	apiListenCfg := cfg.Listen.API
	chatListenCfg := cfg.Listen.Chat

	if apiListenCfg.Type != "port" {
		s.T().Fatalf("[API Server]: can't run integration tests with listen type = %q", apiListenCfg.Type)
	}

	if chatListenCfg.Type != "port" {
		s.T().Fatalf("[Chat Server]: can't run integration tests with listen type = %q", chatListenCfg.Type)
	}

	logging.InitLogger(logging.LogConfig{
		LogLevel:    cfg.Logging.Level,
		LogFilePath: cfg.Logging.FilePath,
		NeedRotate:  cfg.Logging.Rotate,
		MaxSize:     cfg.Logging.MaxSize,
		MaxBackups:  cfg.Logging.MaxBackups,
	})

	dbConn, err := sql.Open("pgx", cfg.DBConnectionURL())
	s.Require().NoError(err, "Failed to create *sql.DB instance")
	s.Require().NoError(dbConn.Ping(), "Failed to connect to postgres")

	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Username: cfg.Redis.Username,
		Password: cfg.Redis.Password,
		DB:       0, // use default database
	})

	s.Require().NoError(redisClient.Ping(context.Background()).Err(), "Failed to connect to redis")

	dbMigration, err := migrate.New("file://"+migrationsPath, cfg.DBConnectionURL())
	s.Require().NoError(err, "Failed to create migrate instance")

	if err = dbMigration.Up(); err != nil && err != migrate.ErrNoChange {
		s.Require().NoError(err, "Failed to migrate up")
	}

	fixtures, err := testfixtures.New(
		testfixtures.Database(dbConn),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory(fixturesPath),
	)
	s.Require().NoError(err, "Failed to create fixtures loader instance")

	s.cfg = cfg
	s.app = app.NewApp(cfg)
	s.dbMigration = dbMigration
	s.dbConn = dbConn
	s.fixtures = fixtures
	s.redisClient = redisClient
	s.httpClient = &http.Client{Timeout: 5 * time.Second}
	s.wsDialer = &ws.Dialer{}
	s.apiURLPrefix = fmt.Sprintf("http://localhost:%d/api", apiListenCfg.BindPort)
	s.chatURL = fmt.Sprintf("ws://localhost:%d", chatListenCfg.BindPort)

	go func() {
		s.Require().NoError(s.app.Run(), "An error occurred while running the application")
	}()
}

func (s *AppTestSuite) TearDownSuite() {
	process, err := os.FindProcess(syscall.Getpid())
	s.Require().NoError(err, "Failed to find process", syscall.Getpid())

	s.Require().NoError(process.Signal(syscall.SIGTERM), "Failed to send SIGTERM to process", process.Pid)
	s.Require().NoError(s.dbMigration.Down(), "Failed to migrate down")
}

func (s *AppTestSuite) SetupTest() {
	s.NoError(s.fixtures.Load(), "Failed to populate database")
}

func (s *AppTestSuite) TearDownTest() {
	s.NoError(
		s.redisClient.FlushAll(context.Background()).Err(),
		"Failed to remove all keys in the redis",
	)
}

func (s *AppTestSuite) buildURL(uri string) string {
	return s.apiURLPrefix + uri
}

func (s *AppTestSuite) newWebsocketConnection(username, password, fingerprint string) *ws.Conn {
	tokenPair := s.authenticate(username, password, fingerprint)

	reqHeaders := http.Header{}
	reqHeaders.Add("Authorization", "Bearer "+tokenPair.AccessToken)

	conn, _, err := s.wsDialer.Dial(s.chatURL, reqHeaders)
	s.Require().NoError(err, "Failed to create websocket connection")

	time.Sleep(50 * time.Millisecond)

	return conn
}
