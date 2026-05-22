package main

import (
	"log"

	"pleco-api/internal/appsetup"
	"pleco-api/internal/config"
)

func main() {
	config.LoadEnv()
	appConfig := config.LoadAppConfig()
	if err := appsetup.RunMigrationsForDriver(appConfig.DatabaseURL, appConfig.DatabaseDriver); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Migration success")
}
