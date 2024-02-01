package test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"syscall"
	"testing"

	"github.com/Chatyx/backend/internal/app"

	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/golang-migrate/migrate/v4"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	configPath     = "../configs/test_config.yaml"
	migrationsPath = "../db/migrations"
	fixturesPath   = "./fixtures"
)

type AppTestSuite struct {
	suite.Suite
	db        *sql.DB
	redisCli  *redis.Client
	mig       *migrate.Migrate
	fixLoader *testfixtures.Loader
}

func TestAppTestSuite(t *testing.T) {
	suite.Run(t, new(AppTestSuite))
}

func (s *AppTestSuite) SetupSuite() {
	var err error

	application := app.NewApp(configPath)
	conf := application.Config()

	connString := fmt.Sprintf(
		"postgres://%s:%s@%s/%s?sslmode=disable",
		conf.Postgres.User, conf.Postgres.Password, net.JoinHostPort(conf.Postgres.Host, conf.Postgres.Port), conf.Postgres.Database,
	)
	s.db, err = sql.Open("pgx", connString)
	s.Require().NoError(err, "Failed to create *sql.DB instance")
	s.Require().NoError(s.db.Ping(), "Failed to connect to postgres")

	redisDBNum, _ := strconv.Atoi(conf.Redis.Database)
	s.redisCli = redis.NewClient(&redis.Options{
		Addr:     net.JoinHostPort(conf.Redis.Host, conf.Redis.Port),
		Username: conf.Redis.User,
		Password: conf.Redis.Password,
		DB:       redisDBNum,
	})
	s.Require().NoError(s.redisCli.Ping(context.Background()).Err(), "Failed to connect to redis")

	s.mig, err = migrate.New("file://"+migrationsPath, connString)
	s.Require().NoError(err, "Failed to create *migrate.Migrate instance")

	if err = s.mig.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		s.Require().NoError(err, "Failed to migrate up")
	}

	s.fixLoader, err = testfixtures.New(
		testfixtures.Database(s.db),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory(fixturesPath),
	)
	s.Require().NoError(err, "Failed to create *testfixtures.Loader instance")

	go application.Run()
}

func (s *AppTestSuite) TearDownSuite() {
	pid := syscall.Getpid()
	process, err := os.FindProcess(pid)
	s.Require().NoErrorf(err, "Failed to find process %d", pid)
	s.Require().NoErrorf(process.Signal(syscall.SIGTERM), "Failed to send SIGTERM to process with pid %d", process.Pid)

	s.NoError(s.mig.Down(), "Failed to migrate down")
	s.NoError(s.redisCli.FlushAll(context.Background()).Err(), "Failed to flush all redis keys")
}

func (s *AppTestSuite) SetupTest() {
	s.Require().NoError(s.fixLoader.Load(), "Failed to populate postgres")
	s.Require().NoError(s.redisCli.FlushAll(context.Background()).Err(), "Failed to flush all redis keys")
}
