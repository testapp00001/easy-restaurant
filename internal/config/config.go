package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	ServerPort  string
	PostgresDSN string
	JWTSecret   string
}

// LoadConfig loads configuration from .env file
func LoadConfig() (*Config, error) {
	// Attempt to load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg := &Config{
		ServerPort:  os.Getenv("SERVER_PORT"),
		PostgresDSN: os.Getenv("POSTGRES_DSN"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
	}

	return cfg, nil
}
