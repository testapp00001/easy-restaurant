package handlers

import (
	"restaurant-api/internal/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// --- Table Management ---

// CreateTable adds a new restaurant table
func CreateTable(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var table models.RestaurantTable
		if err := c.BodyParser(&table); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
		}
		db.Create(&table)
		return c.Status(fiber.StatusCreated).JSON(table)
	}
}

// GetTables lists all tables
func GetTables(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var tables []models.RestaurantTable
		db.Find(&tables)
		return c.JSON(tables)
	}
}

// UpdateTableStatus manually updates a table's status
func UpdateTableStatus(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		var req struct {
			Status models.TableStatus `json:"status"`
		}
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
		}

		db.Model(&models.RestaurantTable{}).Where("id = ?", id).Update("status", req.Status)
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success"})
	}
}

// --- Menu Item Management ---

// CreateMenuItem adds a new menu item
func CreateMenuItem(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var item models.MenuItem
		if err := c.BodyParser(&item); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
		}
		db.Create(&item)
		return c.Status(fiber.StatusCreated).JSON(item)
	}
}

// GetMenuItems lists all menu items
func GetMenuItems(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var items []models.MenuItem
		db.Find(&items)
		return c.JSON(items)
	}
}

// UpdateMenuItem updates a menu item's details (price, availability, etc.)
func UpdateMenuItem(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		var item models.MenuItem
		if err := c.BodyParser(&item); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
		}
		db.Model(&models.MenuItem{}).Where("id = ?", id).Updates(item)
		return c.Status(fiber.StatusOK).JSON(item)
	}
}
