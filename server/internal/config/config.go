package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	ServerPort   string
	PostgresDSN  string
	JWTSecret    string
	AllowOrigins string
}

// LoadConfig loads configuration from .env file
func LoadConfig() (*Config, error) {
	// Attempt to load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	POSTGRES_DSN := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_DB"),
		os.Getenv("DB_PORT"),
	)
	if os.Getenv("POSTGRES_DSN") != "" {
		POSTGRES_DSN = os.Getenv("POSTGRES_DSN")
	}

	fmt.Println(POSTGRES_DSN)

	cfg := &Config{
		ServerPort:   os.Getenv("SERVER_PORT"),
		PostgresDSN:  POSTGRES_DSN,
		JWTSecret:    os.Getenv("JWT_SECRET"),
		AllowOrigins: os.Getenv("ALLOW_ORIGINS"),
	}

	return cfg, nil
}
