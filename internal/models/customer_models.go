package models

import (
	"time"

	"gorm.io/gorm"
)

const (
	CustomerRole string = "Customer"
)

// UserAccount represents a persistent customer account
type UserAccount struct {
	gorm.Model
	PhoneNumber  string `gorm:"unique;not null"`
	PasswordHash string `gorm:"not null"` // For the 6-digit PIN
	Name         string
}

// UserSession represents a single customer visit
type UserSession struct {
	gorm.Model
	UserAccountID   uint
	UserAccount     UserAccount
	BuffetBundleID  uint
	BuffetBundle    BuffetBundle
	CurrentTableID  *uint           // Pointer to allow null
	RestaurantTable RestaurantTable `gorm:"foreignKey:CurrentTableID"`
	Status          string          `gorm:"type:varchar(20);default:'Active'"` // Active, Paid, Ended
	ExpiresAt       time.Time       `gorm:"not null"`
}
