package main

import (
	"log"
	"scht-backend/internal/app"
)

func main() {
	application := app.NewApp()
	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
