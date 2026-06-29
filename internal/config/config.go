package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppPort string
}

func Load() *Config {
	err := godotenv.Load()

	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	appPort := os.Getenv("APP_PORT")

	if appPort == "" {
		appPort = "8080"
	}

	return &Config{
		AppPort: appPort,
	}
}
