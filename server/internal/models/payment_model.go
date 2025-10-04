package models

import "gorm.io/gorm"

// Payment records every financial transaction.
type Payment struct {
	gorm.Model
	UserSessionID      uint
	UserSession        UserSession
	Amount             float64
	PaymentType        string // "BuffetPrepayment", "AlaCarteSettlement"
	PaymentMethod      string // "Cash", "Card", "Online"
	ProcessedByStaffID uint
	StaffUser          StaffUser `gorm:"foreignKey:ProcessedByStaffID"`
}
