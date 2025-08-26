package models

import (
	"time"

	"gorm.io/gorm"
)

type Tariff struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	Name         string         `json:"name" gorm:"not null"`
	MinDistance  float64        `json:"min_distance" gorm:"not null"`
	MaxDistance  float64        `json:"max_distance" gorm:"not null"`
	Price        float64        `json:"price" gorm:"not null"`
	Destinations string         `json:"destinations"` // Contoh destinasi (opsional)
	IsActive     bool           `json:"is_active" gorm:"default:true"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Orders []Order `json:"orders,omitempty" gorm:"foreignKey:TariffID;references:ID"`
}

func (t *Tariff) TableName() string {
	return "tariffs"
}
