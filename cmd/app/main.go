package main

import (
	"flag"

	"github.com/Mort4lis/scht-backend/internal/app"
	"github.com/Mort4lis/scht-backend/internal/config"
	"github.com/Mort4lis/scht-backend/pkg/logging"
)

var cfgPath string

func init() {
	flag.StringVar(&cfgPath, "config", "./configs/main.yml", "config file path")
}

func main() {
	flag.Parse()

	logging.InitLogger(logging.LogConfig{
		LogLevel:    "debug",
		LogFilePath: "./logs/all.log",
		NeedRotate:  true,
		MaxSize:     100,
		MaxBackups:  5,
	})

	cfg := config.GetConfig(cfgPath)
	application := app.NewApp(cfg)
	logger := logging.GetLogger()

	if err := application.Run(); err != nil {
		logger.WithError(err).Fatal("Error occurred while running the application")
	}
}
