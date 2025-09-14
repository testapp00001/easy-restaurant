package models

import "gorm.io/gorm"

// BuffetBundle defines a buffet package
type BuffetBundle struct {
	gorm.Model
	Name       string
	AdultPrice float64
	ChildPrice float64
	IsActive   bool       `gorm:"default:true"`
	MenuItems  []MenuItem `gorm:"many2many:buffet_bundle_menu_items;"`
}

// Category for organizing menu items
type Category struct {
	gorm.Model
	Name           string
	KitchenGroupID uint
	KitchenGroup   KitchenGroup
}

// MenuItem represents an individual dish
type MenuItem struct {
	gorm.Model
	Name        string
	Description string
	Price       float64 // For à la carte
	CategoryID  uint
	Category    Category
	IsAvailable bool `gorm:"default:true"`
}

// KitchenGroup is a specialized group of chefs
type KitchenGroup struct {
	gorm.Model
	Name string
}
