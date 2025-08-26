package models

import (
	"time"

	"gorm.io/gorm"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusAccepted  OrderStatus = "accepted"
	OrderStatusCompleted OrderStatus = "completed"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type Order struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	OrderNumber    string         `json:"order_number" gorm:"unique;not null"`
	CustomerID     *uint          `json:"customer_id"`
	DriverID       *uint          `json:"driver_id"`
	BecakCode      string         `json:"becak_code" gorm:"not null"` // Kode dari sticker barcode
	TariffID       uint           `json:"tariff_id"`
	PickupLocation string         `json:"pickup_location"` // Bisa null, diisi nanti oleh sistem
	DropLocation   string         `json:"drop_location"`   // Bisa null, diisi nanti oleh sistem
	Distance       float64        `json:"distance" gorm:"not null"`
	Price          float64        `json:"price" gorm:"not null"`
	ETA            int            `json:"eta" gorm:"-"` // Estimated Time of Arrival in minutes (calculated field)
	Status         OrderStatus    `json:"status" gorm:"type:enum('pending','accepted','completed','cancelled');default:'pending'"`
	PaymentStatus  string         `json:"payment_status" gorm:"default:'pending'"`
	CustomerPhone  string         `json:"customer_phone" gorm:"not null"`
	CustomerName   string         `json:"customer_name"`
	Notes          string         `json:"notes"`
	AcceptedAt     *time.Time     `json:"accepted_at"`
	CompletedAt    *time.Time     `json:"completed_at"`
	CancelledAt    *time.Time     `json:"cancelled_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Customer User     `json:"customer,omitempty" gorm:"foreignKey:CustomerID;references:ID"`
	Driver   Driver   `json:"driver,omitempty" gorm:"foreignKey:DriverID;references:ID"`
	Tariff   Tariff   `json:"tariff,omitempty" gorm:"foreignKey:TariffID;references:ID"`
	Payment  *Payment `json:"payment,omitempty" gorm:"foreignKey:OrderID;references:ID"`
}

func (o *Order) TableName() string {
	return "orders"
}
