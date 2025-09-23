// file: internal/models/order_models.go
package models

import "gorm.io/gorm"

// Order is a single submission of one or more dishes
type Order struct {
	gorm.Model
	UserSessionID        uint
	UserSession          UserSession
	TableIDAtTimeOfOrder uint
	RestaurantTable      RestaurantTable `gorm:"foreignKey:TableIDAtTimeOfOrder"`
	Status               string          `gorm:"type:varchar(20);default:'PendingApproval'"`
	OrderItems           []OrderItem
}

// OrderItem is a specific dish within an order
type OrderItem struct {
	gorm.Model
	OrderID            uint
	Order              Order
	MenuItemID         uint
	MenuItem           MenuItem
	Quantity           int
	UnitPrice          float64 // 0 for buffet items
	CustomizationNotes string
	Status             string    `gorm:"type:varchar(20);default:'Pending'"`
	AssignedChefID     *uint     // Pointer to allow null
	StaffUser          StaffUser `gorm:"foreignKey:AssignedChefID"`
	RejectedReason     string
}
