package models

import (
	"time"

	"gorm.io/gorm"
)

type DriverLocation struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	DriverID  uint           `json:"driver_id" gorm:"not null;index"`
	Latitude  float64        `json:"latitude" gorm:"not null"`
	Longitude float64        `json:"longitude" gorm:"not null"`
	Accuracy  float64        `json:"accuracy"`
	Speed     float64        `json:"speed"`
	Heading   float64        `json:"heading"`
	IsOnline  bool           `json:"is_online" gorm:"default:false"`
	LastSeen  time.Time      `json:"last_seen"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Driver Driver `json:"driver,omitempty" gorm:"foreignKey:DriverID"`
}

func (dl *DriverLocation) TableName() string {
	return "driver_locations"
}
