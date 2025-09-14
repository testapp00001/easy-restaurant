// file: internal/handlers/customer_handler.go
package handlers

import (
	"log"
	"math/rand"
	"restaurant-api/internal/config"
	"restaurant-api/internal/models"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type CreateAccountRequest struct {
	PhoneNumber string `json:"phone_number"`
	Name        string `json:"name"`
}

// CreateCustomerAccount handles creation of a new customer account by staff
func CreateCustomerAccount(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req CreateAccountRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
		}

		pin := strconv.Itoa(100000 + rand.Intn(900000)) // Generate 6-digit PIN
		hashedPin, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create account"})
		}

		account := models.UserAccount{
			PhoneNumber:  req.PhoneNumber,
			PasswordHash: string(hashedPin),
			Name:         req.Name,
		}
		if result := db.Create(&account); result.Error != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Phone number may already exist"})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"phone_number": account.PhoneNumber, "pin": pin})
	}
}

type CreateSessionRequest struct {
	PhoneNumber    string `json:"phone_number"`
	BuffetBundleID uint   `json:"buffet_bundle_id"`
}

// CreateCustomerSession starts a new dining session for a customer
func CreateCustomerSession(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req CreateSessionRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
		}

		var account models.UserAccount
		if db.Where("phone_number = ?", req.PhoneNumber).First(&account).Error != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Customer account not found"})
		}

		session := models.UserSession{
			UserAccountID:  account.ID,
			BuffetBundleID: req.BuffetBundleID,
			ExpiresAt:      time.Now().Add(3 * time.Hour),
		}
		if result := db.Create(&session); result.Error != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create session"})
		}

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{"user_session_id": session.ID})
	}
}

type CustomerLoginRequest struct {
	PhoneNumber string `json:"phone_number"`
	Pin         string `json:"pin"`
}

// LoginCustomer authenticates a customer and returns a JWT for their active session
func LoginCustomer(db *gorm.DB, cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req CustomerLoginRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"error": "Cannot parse JSON"})
		}

		var account models.UserAccount
		if db.Where("phone_number = ?", req.PhoneNumber).First(&account).Error != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
		}

		if err := bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(req.Pin)); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid credentials"})
		}

		var session models.UserSession
		// Find the most recent active session for this user
		if db.Where("user_account_id = ? AND status = 'Active'", account.ID).Last(&session).Error != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "No active session found. Please ask staff to start one."})
		}

		// We'll create a simplified JWT for customers
		claims := jwt.MapClaims{
			"session_id": session.ID,
			"role":       models.CustomerRole,
			"exp":        session.ExpiresAt.Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(cfg.JWTSecret))
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not generate token"})
		}

		return c.JSON(fiber.Map{"token": tokenString, "user_session_id": session.ID})
	}
}

type AttachTableRequest struct {
	TableID uint `json:"table_id"`
}

// AttachTable allows a customer to associate their session with a physical table
func AttachTable(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessionID := c.Locals("session_id").(uint)

		var req AttachTableRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
		}

		// Update the session with the new table ID
		result := db.Model(&models.UserSession{}).Where("id = ?", sessionID).Update("current_table_id", req.TableID)
		if result.Error != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not attach to table"})
		}
		if result.RowsAffected == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Active session not found"})
		}

		// In a real system, you'd also manage the old and new table statuses here
		db.Model(&models.RestaurantTable{}).Where("id = ?", req.TableID).Update("status", models.OccupiedStatus)

		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "message": "Successfully attached to table"})
	}
}

// GetMenu retrieves the personalized menu for the customer's session
func GetMenu(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessionID := c.Locals("session_id").(uint)

		var session models.UserSession
		if err := db.Preload("BuffetBundle.MenuItems.Category").First(&session, sessionID).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Session not found"})
		}

		var allAlaCarteItems []models.MenuItem
		db.Find(&allAlaCarteItems) // In a real app, you might filter this more

		return c.JSON(fiber.Map{
			"buffet_name":      session.BuffetBundle.Name,
			"buffet_items":     session.BuffetBundle.MenuItems,
			"a_la_carte_items": allAlaCarteItems,
		})
	}
}

type CreateOrderRequest struct {
	Items []struct {
		MenuItemID uint   `json:"menu_item_id"`
		Quantity   int    `json:"quantity"`
		Notes      string `json:"notes"`
	} `json:"items"`
}

// CreateOrder allows a customer to place a new order
func CreateOrder(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessionID := c.Locals("session_id").(uint)

		var req CreateOrderRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
		}

		var session models.UserSession
		if db.First(&session, sessionID).Error != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Session not found"})
		}
		if session.CurrentTableID == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Please attach to a table before ordering"})
		}

		var orderItems []models.OrderItem
		for _, itemReq := range req.Items {
			// In a real app, you'd add logic here to check if the item is part of the buffet
			// and set the UnitPrice to 0. For now, we'll assume à la carte for simplicity.
			var menuItem models.MenuItem
			db.First(&menuItem, itemReq.MenuItemID)

			orderItems = append(orderItems, models.OrderItem{
				MenuItemID:         itemReq.MenuItemID,
				Quantity:           itemReq.Quantity,
				CustomizationNotes: itemReq.Notes,
				UnitPrice:          menuItem.Price, // This would be 0 for buffet items
			})
		}

		order := models.Order{
			UserSessionID:        sessionID,
			TableIDAtTimeOfOrder: *session.CurrentTableID,
			OrderItems:           orderItems,
		}

		if result := db.Create(&order); result.Error != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create order"})
		}

		// This is where you would trigger the WebSocket event to the kitchen.
		// We will implement that in the next milestone.
		log.Printf("New order created: %d, for table ID: %d", order.ID, order.TableIDAtTimeOfOrder)

		return c.Status(fiber.StatusCreated).JSON(order)
	}
}
