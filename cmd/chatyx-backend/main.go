package main

import (
	"flag"

	"github.com/Chatyx/backend/internal/app"
)

var confPath string

func init() {
	flag.StringVar(&confPath, "config", "./configs/config.yaml", "config file path")
}

func main() {
	flag.Parse()
	app.NewApp(confPath).Run()
}
