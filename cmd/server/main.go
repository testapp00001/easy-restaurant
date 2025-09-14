package main

import (
	"log"
	"restaurant-api/internal/config"
	"restaurant-api/internal/database"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Could not load configuration: %v", err)
	}

	// 2. Connect to Database
	database.ConnectDB(cfg)

	// 3. Initialize Fiber App
	app := fiber.New()

	// Add a logger middleware for better request logging
	app.Use(logger.New())

	// 4. Setup a simple health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "success",
			"message": "Welcome to the Restaurant API!",
		})
	})

	// 5. Start the server
	log.Printf("Server starting on port %s", cfg.ServerPort)
	err = app.Listen(cfg.ServerPort)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
