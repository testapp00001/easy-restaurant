package auth

import (
	"fmt"
	"restaurant-api/internal/config"
	"restaurant-api/internal/models"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// GenerateJWT creates a new JWT token for a staff user
func GenerateJWT(user *models.StaffUser, cfg *config.Config) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 72).Unix(), // Token expires in 3 days
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecret))
}

// AuthMiddleware creates a Fiber middleware for JWT authentication.
func AuthMiddleware(cfg *config.Config, requiredRole models.StaffRole) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing or malformed JWT"})
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing or malformed JWT"})
		}
		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired JWT"})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid JWT claims"})
		}

		// Check role
		role := models.StaffRole(claims["role"].(string))
		if role != requiredRole && role != models.AdminRole { // Admins can do anything
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Insufficient permissions"})
		}

		c.Locals("user_id", claims["user_id"])
		c.Locals("role", role)
		return c.Next()
	}
}

// CustomerAuthMiddleware creates a Fiber middleware for customer JWT authentication.
func CustomerAuthMiddleware(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing or malformed JWT"})
		}
		tokenString := parts[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired JWT"})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || claims["role"] != models.CustomerRole {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
		}

		// Extract session_id and convert to uint
		sessionIDFloat, ok := claims["session_id"].(float64)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid session ID in token"})
		}
		sessionID := uint(sessionIDFloat)

		c.Locals("session_id", sessionID)
		return c.Next()
	}
}
