package models

import (
	"time"

	"gorm.io/gorm"
)

type DriverStatus string

const (
	DriverStatusActive   DriverStatus = "active"
	DriverStatusInactive DriverStatus = "inactive"
	DriverStatusOnTrip   DriverStatus = "on_trip"
)

type VehicleType string

const (
	VehicleTypeBecakManual  VehicleType = "becak_manual"
	VehicleTypeBecakMotor   VehicleType = "becak_motor"
	VehicleTypeBecakListrik VehicleType = "becak_listrik"
	VehicleTypeAndong       VehicleType = "andong"
)

type Driver struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	UserID        *uint          `json:"user_id" gorm:"unique"`
	DriverCode    string         `json:"driver_code" gorm:"unique;not null"`
	Name          string         `json:"name" gorm:"not null"`
	Phone         string         `json:"phone" gorm:"not null"`
	Email         string         `json:"email"`
	Address       string         `json:"address"`
	IDCard        string         `json:"id_card" gorm:"unique"`
	VehicleNumber string         `json:"vehicle_number"`
	VehicleType   VehicleType    `json:"vehicle_type" gorm:"type:enum('becak_manual','becak_motor','becak_listrik','andong');default:'becak_manual'"`
	Status        DriverStatus   `json:"status" gorm:"type:enum('active','inactive','on_trip');default:'active'"`
	IsActive      bool           `json:"is_active" gorm:"default:true"`
	Rating        float64        `json:"rating" gorm:"default:0"`
	TotalTrips    int            `json:"total_trips" gorm:"default:0"`
	TotalEarnings float64        `json:"total_earnings" gorm:"default:0"`
	FCMToken      string         `json:"fcm_token" gorm:"column:fcm_token"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	// Calculated fields (not stored in database)
	AvailableBalance     float64 `json:"available_balance,omitempty" gorm:"-"`
	CompletedWithdrawals float64 `json:"completed_withdrawals,omitempty" gorm:"-"`

	// Relationships
	User        User         `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID"`
	Orders      []Order      `json:"orders,omitempty" gorm:"foreignKey:DriverID;references:ID"`
	Withdrawals []Withdrawal `json:"withdrawals,omitempty" gorm:"foreignKey:DriverID;references:ID"`
}

func (d *Driver) TableName() string {
	return "drivers"
}
