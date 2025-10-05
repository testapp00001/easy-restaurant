package handlers

import (
	"restaurant-api/internal/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type CreateTableRequest struct {
	TableNumber string `json:"table_number"`
}

type UpdateTableStatusRequest struct {
	Status models.TableStatus `json:"status"`
}

type CreateMenuItemRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	CategoryID  uint    `json:"category_id"`
	IsAvailable bool    `json:"is_available"`
}

type UpdateMenuItemRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	CategoryID  uint    `json:"category_id"`
	IsAvailable bool    `json:"is_available"`
}

// --- Table Management ---

// CreateTable adds a new restaurant table
func CreateTable(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req CreateTableRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
		}

		table := models.RestaurantTable{
			TableNumber: req.TableNumber,
		}

		if err := db.Create(&table).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create table"})
		}

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

		var req UpdateTableStatusRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
		}

		result := db.Model(&models.RestaurantTable{}).Where("id = ?", id).Update("status", req.Status)
		if result.Error != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update status"})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success"})
	}
}

// --- Menu Item Management ---

// CreateMenuItem adds a new menu item
func CreateMenuItem(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req CreateMenuItemRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
		}

		item := models.MenuItem{
			Name:        req.Name,
			Description: req.Description,
			Price:       req.Price,
			CategoryID:  req.CategoryID,
			IsAvailable: req.IsAvailable,
		}

		if err := db.Create(&item).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create menu item"})
		}

		return c.Status(fiber.StatusCreated).JSON(item)
	}
}

// GetMenuItems lists all menu items
func GetMenuItems(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var items []models.MenuItem
		db.Preload("Category").Find(&items)
		return c.JSON(items)
	}
}

// UpdateMenuItem updates a menu item's details (price, availability, etc.)
func UpdateMenuItem(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")

		var req UpdateMenuItemRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
		}

		updateData := map[string]interface{}{
			"name":         req.Name,
			"description":  req.Description,
			"price":        req.Price,
			"category_id":  req.CategoryID,
			"is_available": req.IsAvailable,
		}

		result := db.Model(&models.MenuItem{}).Where("id = ?", id).Updates(updateData)
		if result.Error != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update menu item"})
		}
		if result.RowsAffected == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Menu item not found"})
		}

		var updatedItem models.MenuItem
		db.First(&updatedItem, id)
		return c.Status(fiber.StatusOK).JSON(updatedItem)
	}
}
