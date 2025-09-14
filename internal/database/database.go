package database

import (
	"log"
	"restaurant-api/internal/config"
	"restaurant-api/internal/models"

	"golang.org/x/crypto/bcrypt"
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

	err = DB.AutoMigrate(
		&models.StaffUser{},
		&models.RestaurantTable{},
		&models.UserAccount{},
		&models.UserSession{},
		&models.KitchenGroup{},
		&models.Category{},
		&models.MenuItem{},
		&models.BuffetBundle{},
		&models.Order{},
		&models.OrderItem{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	log.Println("Database migrated")
}

// SeedDatabase creates an initial admin user if one doesn't exist
func SeedDatabase() {
	var userCount int64
	DB.Model(&models.StaffUser{}).Count(&userCount)

	if userCount == 0 {
		log.Println("No staff users found, seeding admin user...")
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123@"), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("Failed to hash password: %v", err)
		}

		adminUser := models.StaffUser{
			Username:     "admin",
			PasswordHash: string(hashedPassword),
			Role:         models.AdminRole,
		}
		if result := DB.Create(&adminUser); result.Error != nil {
			log.Fatalf("Failed to seed admin user: %v", result.Error)
		}
		log.Println("Admin user seeded successfully")
	}

	var staffCount int64
	DB.Model(&models.StaffUser{}).Count(&staffCount)
	if staffCount == 0 {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		adminUser := models.StaffUser{Username: "admin", PasswordHash: string(hashedPassword), Role: models.AdminRole}
		DB.Create(&adminUser)
	}

	// Seed Kitchen Groups, Categories, Menu Items, and Buffet Bundles
	var kitchenGroupCount int64
	DB.Model(&models.KitchenGroup{}).Count(&kitchenGroupCount)
	if kitchenGroupCount == 0 {
		log.Println("Seeding menu data...")

		// Create Kitchen Groups
		grillGroup := models.KitchenGroup{Name: "Grill Station"}
		sushiGroup := models.KitchenGroup{Name: "Sushi Bar"}
		kitchenGroups := []models.KitchenGroup{grillGroup, sushiGroup}
		DB.Create(&kitchenGroups)

		// Create Categories
		steakCategory := models.Category{Name: "Steaks", KitchenGroupID: kitchenGroups[0].ID}
		sushiCategory := models.Category{Name: "Sushi Rolls", KitchenGroupID: kitchenGroups[1].ID}
		categories := []models.Category{steakCategory, sushiCategory}
		DB.Create(&categories)

		// Create Menu Items
		ribeye := models.MenuItem{Name: "Ribeye Steak", Price: 35.50, CategoryID: categories[0].ID}
		californiaRoll := models.MenuItem{Name: "California Roll", Price: 8.00, CategoryID: categories[1].ID}
		tunaRoll := models.MenuItem{Name: "Tuna Roll", Price: 9.50, CategoryID: categories[1].ID}
		menuItems := []models.MenuItem{ribeye, californiaRoll, tunaRoll}
		DB.Create(&menuItems)

		// Create Buffet Bundles
		basicBuffet := models.BuffetBundle{Name: "Basic Buffet", AdultPrice: 25.00, ChildPrice: 12.00}
		premiumBuffet := models.BuffetBundle{Name: "Premium Buffet", AdultPrice: 45.00, ChildPrice: 22.00}
		buffetBundles := []models.BuffetBundle{basicBuffet, premiumBuffet}
		DB.Create(&buffetBundles)

		// Associate Menu Items with Bundles
		DB.Model(&buffetBundles[0]).Association("MenuItems").Append(&[]models.MenuItem{menuItems[1]})
		DB.Model(&buffetBundles[1]).Association("MenuItems").Append(&[]models.MenuItem{menuItems[0], menuItems[1], menuItems[2]})

		DB.Create(&models.RestaurantTable{TableNumber: "table-1", Status: models.AvailableStatus})

		log.Println("Menu data seeded.")
	}
}
