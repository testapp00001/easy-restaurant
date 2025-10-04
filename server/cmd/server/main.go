package main

import (
	"log"
	"os"
	"restaurant-api/internal/auth"
	"restaurant-api/internal/config"
	"restaurant-api/internal/database"
	"restaurant-api/internal/handlers"
	"restaurant-api/internal/hub"
	"restaurant-api/internal/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/websocket/v2"
	"gorm.io/gorm"
)

// Add this function somewhere in your main.go
func runSessionCleanup(db *gorm.DB) {
	ticker := time.NewTicker(5 * time.Minute) // Check every 5 minutes
	defer ticker.Stop()

	for range ticker.C {
		log.Println("Running session cleanup...")
		now := time.Now()
		result := db.Model(&models.UserSession{}).
			Where("status = ? AND expires_at < ?", "Active", now).
			Update("status", "Ended")

		if result.Error != nil {
			log.Printf("Error during session cleanup: %v", result.Error)
		} else if result.RowsAffected > 0 {
			log.Printf("Cleaned up %d expired sessions.", result.RowsAffected)
		}
	}
}

func main() {
	// 1. Load Configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Could not load configuration: %v", err)
	}

	// 2. Connect to Database
	database.ConnectDB(cfg)
	database.SeedDatabase()

	// Start the session cleanup goroutine
	go runSessionCleanup(database.DB)

	// 3. Initialize Fiber App
	app := fiber.New()

	// Add CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: cfg.AllowOrigins, // CHANGE THIS
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Add a logger middleware for better request logging
	app.Use(logger.New(logger.Config{
		Format: `{"ip":"${ip}","timestamp":"${time}","status":${status},"latency":"${latency}","method":"${method}","path":"${path}","error":"${error}"}` + "\n",
		Output: os.Stdout,
	}))

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

	// --- Admin Routes (Protected by Admin only) ---
	adminGroup := api.Group("/admin")
	adminGroup.Use(auth.AuthMiddleware(cfg, models.AdminRole))
	// Table management
	adminGroup.Post("/tables", handlers.CreateTable(database.DB))
	adminGroup.Get("/tables", handlers.GetTables(database.DB))
	adminGroup.Put("/tables/:id/status", handlers.UpdateTableStatus(database.DB))
	// Menu management
	adminGroup.Post("/menu-items", handlers.CreateMenuItem(database.DB))
	adminGroup.Get("/menu-items", handlers.GetMenuItems(database.DB))
	adminGroup.Put("/menu-items/:id", handlers.UpdateMenuItem(database.DB))

	// --- Staff Routes (Protected) ---
	staffGroup := api.Group("/staff")
	// Apply middleware for WaitStaff and above
	staffGroup.Use(auth.AuthMiddleware(cfg, models.WaitStaffRole))
	staffGroup.Post("/customer-accounts", handlers.CreateCustomerAccount(database.DB))
	staffGroup.Post("/customer-sessions", handlers.CreateCustomerSession(database.DB))
	// Billing routes
	staffGroup.Get("/sessions/:id/bill", handlers.GetSessionBill(database.DB))
	staffGroup.Post("/sessions/:id/payments", handlers.ProcessSessionPayment(database.DB))

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
