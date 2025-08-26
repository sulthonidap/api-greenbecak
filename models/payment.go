package models

import (
	"time"

	"gorm.io/gorm"
)

type PaymentStatus string
type PaymentMethod string

const (
	PaymentStatusPending  PaymentStatus = "pending"
	PaymentStatusPaid     PaymentStatus = "paid"
	PaymentStatusFailed   PaymentStatus = "failed"
	PaymentStatusRefunded PaymentStatus = "refunded"

	PaymentMethodCash     PaymentMethod = "cash"
	PaymentMethodTransfer PaymentMethod = "transfer"
	PaymentMethodQR       PaymentMethod = "qr"
)

type Payment struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	OrderID   uint           `json:"order_id" gorm:"unique"`
	Amount    float64        `json:"amount" gorm:"not null"`
	Method    PaymentMethod  `json:"method" gorm:"type:enum('cash','transfer','qr');default:'cash'"`
	Status    PaymentStatus  `json:"status" gorm:"type:enum('pending','paid','failed','refunded');default:'pending'"`
	Reference string         `json:"reference"`
	PaidAt    *time.Time     `json:"paid_at"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Order Order `json:"order,omitempty" gorm:"foreignKey:OrderID;references:ID"`
}

func (p *Payment) TableName() string {
	return "payments"
}
