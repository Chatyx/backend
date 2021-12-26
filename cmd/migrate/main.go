package main

import (
	"flag"
	"log"

	"github.com/Mort4lis/scht-backend/internal/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

var (
	cfgPath        string
	migrationsPath string
)

func init() {
	flag.StringVar(&cfgPath, "config", "./configs/main.yml", "config file path")
	flag.StringVar(&migrationsPath, "migrations", "./internal/db/migrations", "migrations file path")
}

func main() {
	cfg := config.GetConfig(cfgPath)

	dbMigration, err := migrate.New("file://"+migrationsPath, cfg.DBConnectionURL())
	if err != nil {
		log.Fatalf("failed to create migrate instance: %v", err)
	}

	if err = dbMigration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("failed to migrate up: %v", err)
	}
}
