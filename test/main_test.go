package test

import (
	"os"
	"syscall"
	"testing"

	"github.com/Mort4lis/scht-backend/internal/app"
	"github.com/Mort4lis/scht-backend/internal/config"
	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/suite"
)

const (
	configPath     = "../configs/test.yml"
	migrationsPath = "../internal/db/migrations"
)

type AppTestSuite struct {
	suite.Suite
	app         *app.App
	cfg         *config.Config
	dbMigration *migrate.Migrate
}

func TestAppTestSuite(t *testing.T) {
	suite.Run(t, new(AppTestSuite))
}

func (s *AppTestSuite) SetupSuite() {
	s.cfg = config.GetConfig(configPath)

	logging.InitLogger(logging.LogConfig{
		LogLevel:    s.cfg.Logging.Level,
		LogFilePath: s.cfg.Logging.FilePath,
		NeedRotate:  s.cfg.Logging.Rotate,
		MaxSize:     s.cfg.Logging.MaxSize,
		MaxBackups:  s.cfg.Logging.MaxBackups,
	})

	dbMigration, err := migrate.New("file://"+migrationsPath, s.cfg.DBConnectionURL())
	s.Require().NoError(err, "An error occurred while creating migrate instance")

	s.dbMigration = dbMigration
	s.app = app.NewApp(s.cfg)

	go func() {
		s.Require().NoError(s.app.Run(), "An error occurred while running the application")
	}()
}

func (s *AppTestSuite) TearDownSuite() {
	process, err := os.FindProcess(syscall.Getpid())
	s.Require().NoError(err, "Failed to find process", syscall.Getpid())

	s.Require().NoError(process.Signal(syscall.SIGTERM), "Failed to send SIGTERM to process", process.Pid)
}

func (s *AppTestSuite) SetupTest() {
	s.Require().NoError(s.dbMigration.Up())
}

func (s *AppTestSuite) TearDownTest() {
	s.Require().NoError(s.dbMigration.Down())
}
