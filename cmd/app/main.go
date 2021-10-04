package main

import (
	"flag"

	"github.com/Mort4lis/scht-backend/internal/app"
	"github.com/Mort4lis/scht-backend/internal/config"
	"github.com/Mort4lis/scht-backend/pkg/logging"
)

var cfgPath string

func init() {
	flag.StringVar(&cfgPath, "config", "./configs/dev.yml", "config file path")
}

func main() {
	flag.Parse()

	cfg := config.GetConfig(cfgPath)

	logging.InitLogger(logging.LogConfig{
		LogLevel:    cfg.Logging.Level,
		LogFilePath: cfg.Logging.FilePath,
		NeedRotate:  cfg.Logging.Rotate,
		MaxSize:     cfg.Logging.MaxSize,
		MaxBackups:  cfg.Logging.MaxBackups,
	})

	application := app.NewApp(cfg)
	logger := logging.GetLogger()

	if err := application.Run(); err != nil {
		logger.WithError(err).Fatal("Error occurred while running the application")
	}
}
