package main

import (
	"flag"
)

var confPath string

func init() {
	flag.StringVar(&confPath, "config", "./configs/config.yaml", "config file path")
}

func main() {
	flag.Parse()
}
