package models

import (
	"time"

	"gorm.io/gorm"
)

// Hotel mitra. QR hotel mengarah ke /booking?hotel=<slug> sehingga titik jemput
// terisi otomatis dan order tercatat berasal dari hotel ini.
type Hotel struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"not null"`
	Slug      string         `json:"slug" gorm:"type:varchar(100);unique;not null"`
	Address   string         `json:"address"`
	Latitude  float64        `json:"latitude" gorm:"not null"`
	Longitude float64        `json:"longitude" gorm:"not null"`
	Phone     string         `json:"phone"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (h *Hotel) TableName() string {
	return "hotels"
}

// Destination adalah destinasi populer yang muncul paling atas saat penumpang mencari tujuan.
type Destination struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"not null"`
	Category  string         `json:"category"` // wisata, belanja, kuliner, transportasi, ...
	Address   string         `json:"address"`
	Latitude  float64        `json:"latitude" gorm:"not null"`
	Longitude float64        `json:"longitude" gorm:"not null"`
	SortOrder int            `json:"sort_order" gorm:"default:0"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (d *Destination) TableName() string {
	return "destinations"
}

// BookingRate adalah tarif booking per km untuk satu jenis kendaraan.
// Harga = max(MinFare, BaseFare + km × PerKm), dibulatkan ke atas ke kelipatan Rounding.
type BookingRate struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	VehicleType   string    `json:"vehicle_type" gorm:"type:varchar(30);unique;not null"` // sama dengan drivers.vehicle_type
	Label         string    `json:"label" gorm:"not null"`
	BaseFare      float64   `json:"base_fare" gorm:"not null"`
	PerKm         float64   `json:"per_km" gorm:"not null"`
	MinFare       float64   `json:"min_fare" gorm:"not null"`
	MaxDistanceKm float64   `json:"max_distance_km" gorm:"not null;default:10"`
	Rounding      float64   `json:"rounding" gorm:"not null;default:1000"`
	IsActive      bool      `json:"is_active" gorm:"default:true"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (r *BookingRate) TableName() string {
	return "booking_rates"
}
