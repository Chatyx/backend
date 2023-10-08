package main

import (
	"flag"
	"fmt"
)

var confPath string

func init() {
	flag.StringVar(&confPath, "config", "./configs/config.yaml", "config file path")
}

func main() {
	flag.Parse()

	fmt.Println("Hello, world")
	//time.Parse("20060102")
}
