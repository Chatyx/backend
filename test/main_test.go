package test

import (
	"database/sql"
	"os"
	"syscall"
	"testing"

	"github.com/Mort4lis/scht-backend/internal/app"
	"github.com/Mort4lis/scht-backend/internal/config"
	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/go-testfixtures/testfixtures/v3"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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
	app         *app.App
	cfg         *config.Config
	dbMigration *migrate.Migrate
	fixtures    *testfixtures.Loader
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
	s.Require().NoError(err, "Failed to create migrate instance")
	s.Require().NoError(dbMigration.Up(), "Failed to migrate up")

	db, err := sql.Open("pgx", s.cfg.DBConnectionURL())
	s.Require().NoError(err, "Failed to create *sql.DB instance")

	fixtures, err := testfixtures.New(
		testfixtures.Database(db),
		testfixtures.Dialect("postgres"),
		testfixtures.Directory(fixturesPath),
	)
	s.Require().NoError(err, "Failed to create fixtures loader instance")

	s.dbMigration = dbMigration
	s.fixtures = fixtures
	s.app = app.NewApp(s.cfg)

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
	s.Require().NoError(s.fixtures.Load())
}
