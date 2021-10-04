package test

import (
	"os"
	"syscall"
	"testing"

	"github.com/Mort4lis/scht-backend/internal/app"
	"github.com/Mort4lis/scht-backend/internal/config"
	"github.com/Mort4lis/scht-backend/pkg/logging"
	"github.com/stretchr/testify/suite"
)

const configPath = "../configs/test.yml"

type AppTestSuite struct {
	suite.Suite

	logger logging.Logger
}

func TestAppTestSuite(t *testing.T) {
	suite.Run(t, new(AppTestSuite))
}

func (s *AppTestSuite) SetupSuite() {
	cfg := config.GetConfig(configPath)

	logging.InitLogger(logging.LogConfig{
		LogLevel:    cfg.Logging.Level,
		LogFilePath: cfg.Logging.FilePath,
		NeedRotate:  cfg.Logging.Rotate,
		MaxSize:     cfg.Logging.MaxSize,
		MaxBackups:  cfg.Logging.MaxBackups,
	})

	s.logger = logging.GetLogger()

	s.logger.Info("Start to setup test suite...")

	application := app.NewApp(cfg)

	go func() {
		if err := application.Run(); err != nil {
			s.logger.WithError(err).Fatal("An error occurred while running the application")
		}
	}()
}

func (s *AppTestSuite) TearDownSuite() {
	process, err := os.FindProcess(syscall.Getpid())
	if err != nil {
		s.logger.WithError(err).Fatalf("Failed to find process %d", syscall.Getpid())
	}

	if err = process.Signal(syscall.SIGTERM); err != nil {
		s.logger.WithError(err).Errorf("Failed to send SIGTERM to process %d", process.Pid)
	}

	s.logger.Info("Test suite is successfully finished!")
}

func (s *AppTestSuite) SetupTest() {
	s.logger.Info("Setup test 'HelloWorld'")
}

func (s *AppTestSuite) TearDownTest() {
	s.logger.Info("Teardown test 'HelloWorld'")
}

func (s *AppTestSuite) TestHelloWorld() {
	s.T().Error("Hello world from tests")
}
