package models

import "gorm.io/gorm"

type StaffRole string

const (
	AdminRole          StaffRole = "Admin"
	KitchenManagerRole StaffRole = "KitchenManager"
	ChefRole           StaffRole = "Chef"
	WaitStaffRole      StaffRole = "WaitStaff"
)

type StaffUser struct {
	gorm.Model
	Username     string    `gorm:"unique;not null"`
	PasswordHash string    `gorm:"not null"`
	Role         StaffRole `gorm:"type:varchar(20);not null"`
}
