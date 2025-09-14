package handlers

import (
	"restaurant-api/internal/auth"
	"restaurant-api/internal/config"
	"restaurant-api/internal/models"

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
