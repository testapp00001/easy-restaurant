package database

import (
	"log"
	"restaurant-api/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB is the global database connection
var DB *gorm.DB

// ConnectDB initializes the database connection
func ConnectDB(cfg *config.Config) {
	var err error
	DB, err = gorm.Open(postgres.Open(cfg.PostgresDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connection successfully opened")
}
