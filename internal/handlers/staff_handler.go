package handlers

import (
	"restaurant-api/internal/auth"
	"restaurant-api/internal/config"
	"restaurant-api/internal/models"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginStaff authenticates a staff user and returns a JWT
func LoginStaff(db *gorm.DB, cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
		}

		var user models.StaffUser
		if result := db.Where("username = ?", req.Username).First(&user); result.Error != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid username or password"})
		}

		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid username or password"})
		}

		token, err := auth.GenerateJWT(&user, cfg)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Could not generate token"})
		}

		return c.JSON(fiber.Map{"token": token})
	}
}

// GetSessionBill calculates the outstanding balance for a customer's session.
func GetSessionBill(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessionID := c.Params("id")

		var total float64
		// This is the corrected, efficient, single-query approach.
		// It joins order_items with orders, filters by the session, and sums the cost.
		result := db.Model(&models.OrderItem{}).
			Joins("JOIN orders ON orders.id = order_items.order_id").
			Where("orders.user_session_id = ? AND orders.is_paid = ?", sessionID, false).
			Select("COALESCE(SUM(order_items.quantity * order_items.unit_price), 0)").
			Row()

		if err := result.Scan(&total); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to calculate bill"})
		}

		// It's also useful to return the itemized list for the receipt.
		var unpaidItems []models.OrderItem
		db.Preload("MenuItem").
			Joins("JOIN orders ON orders.id = order_items.order_id").
			Where("orders.user_session_id = ? AND orders.is_paid = ? AND order_items.unit_price > 0", sessionID, false).
			Find(&unpaidItems)

		return c.JSON(fiber.Map{
			"outstanding_amount": total,
			"unpaid_items":       unpaidItems,
		})
	}
}

type ProcessPaymentRequest struct {
	Amount        float64 `json:"amount"`
	PaymentMethod string  `json:"payment_method"`
	PaymentType   string  `json:"payment_type"` // "BuffetPrepayment" or "AlaCarteSettlement"
}

// ProcessSessionPayment records a payment for a session.
func ProcessSessionPayment(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessionID, _ := strconv.Atoi(c.Params("id"))
		staffID := c.Locals("user_id").(uint)

		var req ProcessPaymentRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
		}

		payment := models.Payment{
			UserSessionID:      uint(sessionID),
			Amount:             req.Amount,
			PaymentMethod:      req.PaymentMethod,
			PaymentType:        req.PaymentType,
			ProcessedByStaffID: staffID,
		}

		err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&payment).Error; err != nil {
				return err
			}

			switch req.PaymentType {
			case "BuffetPrepayment":
				if err := tx.Model(&models.UserSession{}).Where("id = ?", sessionID).Update("buffet_payment_status", "Paid").Error; err != nil {
					return err
				}
			case "AlaCarteSettlement":
				// Find all Order IDs that contain unpaid a la carte items for this session
				subQuery := tx.Model(&models.OrderItem{}).
					Select("order_id").
					Joins("JOIN orders ON orders.id = order_items.order_id").
					Where("orders.user_session_id = ? AND order_items.unit_price > 0 AND orders.is_paid = ?", sessionID, false)

				// Mark only those specific orders as paid
				if err := tx.Model(&models.Order{}).Where("id IN (?)", subQuery).Update("is_paid", true).Error; err != nil {
					return err
				}

				// Close the entire session
				if err := tx.Model(&models.UserSession{}).Where("id = ?", sessionID).Update("status", "Paid").Error; err != nil {
					return err
				}
			}

			return nil
		})

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Payment processing failed: " + err.Error()})
		}

		return c.Status(fiber.StatusCreated).JSON(payment)
	}
}
