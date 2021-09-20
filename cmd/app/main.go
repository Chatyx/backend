package main

import (
	"flag"
	"log"

	"github.com/Mort4lis/scht-backend/internal/app"
	"github.com/Mort4lis/scht-backend/internal/config"
)

var cfgPath string

func init() {
	flag.StringVar(&cfgPath, "config", "./configs/main.yml", "config file path")
}

func main() {
	flag.Parse()

	cfg, err := config.GetConfig(cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	application := app.NewApp(cfg)
	if err = application.Run(); err != nil {
		log.Fatal(err)
	}
}
