package models

import "gorm.io/gorm"

type TableStatus string

const (
	AvailableStatus TableStatus = "Available"
	OccupiedStatus  TableStatus = "Occupied"
	CleaningStatus  TableStatus = "Cleaning"
)

type RestaurantTable struct {
	gorm.Model
	TableNumber string      `gorm:"unique;not null"`
	Status      TableStatus `gorm:"type:varchar(20);default:'Available'"`
}
