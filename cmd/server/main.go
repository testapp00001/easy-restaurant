package main

import (
	"log"
	"restaurant-api/internal/auth"
	"restaurant-api/internal/config"
	"restaurant-api/internal/database"
	"restaurant-api/internal/handlers"
	"restaurant-api/internal/hub"
	"restaurant-api/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/websocket/v2"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Could not load configuration: %v", err)
	}

	// 2. Connect to Database
	database.ConnectDB(cfg)
	database.SeedDatabase()

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

	// Initialize the WebSocket Hub
	wsHub := hub.NewHub() // Create a local variable for the hub
	go wsHub.Run()        // Run it in a goroutine

	// WebSocket Upgrade Middleware and Endpoint
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// websocket route
	app.Get("/ws/:id", websocket.New(handlers.WebSocketHandler(wsHub)))

	// --- API V1 Routes ---
	api := app.Group("/api/v1")

	// --- Auth Routes (Public) ---
	authGroup := api.Group("/auth")
	authGroup.Post("/staff/login", handlers.LoginStaff(database.DB, cfg))
	authGroup.Post("/customer/login", handlers.LoginCustomer(database.DB, cfg))

	// --- Staff Routes (Protected) ---
	staffGroup := api.Group("/staff")
	// Apply middleware for WaitStaff and above
	staffGroup.Use(auth.AuthMiddleware(cfg, models.WaitStaffRole))
	staffGroup.Post("/customer-accounts", handlers.CreateCustomerAccount(database.DB))
	staffGroup.Post("/customer-sessions", handlers.CreateCustomerSession(database.DB))

	// --- Customer Routes (Protected by Customer Auth) ---
	customerGroup := api.Group("/customer")
	customerGroup.Use(auth.CustomerAuthMiddleware(cfg))
	customerGroup.Post("/session/attach-table", handlers.AttachTable(database.DB))
	customerGroup.Get("/menu", handlers.GetMenu(database.DB))
	customerGroup.Post("/orders", handlers.CreateOrder(database.DB))

	// --- Kitchen Routes (Protected by Staff Auth) ---
	kitchenGroup := api.Group("/kitchen")
	kitchenGroup.Use(auth.AuthMiddleware(cfg, models.ChefRole)) // Protects all kitchen routes

	// Manager-specific routes
	kitchenGroup.Get("/orders/pending", auth.AuthMiddleware(cfg, models.KitchenManagerRole), handlers.GetPendingOrders(database.DB))
	kitchenGroup.Put("/orders/review", auth.AuthMiddleware(cfg, models.KitchenManagerRole), handlers.ReviewOrder(database.DB, wsHub))

	// Chef routes
	kitchenGroup.Get("/group/:id/items", handlers.GetGroupItems(database.DB))
	kitchenGroup.Put("/order-items/:id/assign", handlers.AssignOrderItem(database.DB))
	kitchenGroup.Put("/order-items/:id/status", handlers.UpdateOrderItemStatus(database.DB))

	// 5. Start the server
	log.Printf("Server starting on port %s", cfg.ServerPort)
	err = app.Listen(cfg.ServerPort)
	if err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
