package models

import (
	"time"

	"gorm.io/gorm"
)

type WithdrawalStatus string

const (
	WithdrawalStatusPending   WithdrawalStatus = "pending"
	WithdrawalStatusApproved  WithdrawalStatus = "approved"
	WithdrawalStatusRejected  WithdrawalStatus = "rejected"
	WithdrawalStatusCompleted WithdrawalStatus = "completed"
)

type Withdrawal struct {
	ID            uint             `json:"id" gorm:"primaryKey"`
	DriverID      uint             `json:"driver_id"`
	Amount        float64          `json:"amount" gorm:"not null"`
	Status        WithdrawalStatus `json:"status" gorm:"type:enum('pending','approved','rejected','completed');default:'pending'"`
	BankName      string           `json:"bank_name"`
	AccountNumber string           `json:"account_number"`
	AccountName   string           `json:"account_name"`
	Notes         string           `json:"notes"`
	ApprovedAt    *time.Time       `json:"approved_at"`
	ApprovedBy    *string          `json:"approved_by"`
	RejectedAt    *time.Time       `json:"rejected_at"`
	RejectedBy    *string          `json:"rejected_by"`
	CompletedAt   *time.Time       `json:"completed_at"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
	DeletedAt     gorm.DeletedAt   `json:"-" gorm:"index"`

	// Relationships
	Driver Driver `json:"driver,omitempty" gorm:"foreignKey:DriverID;references:ID"`
}

func (w *Withdrawal) TableName() string {
	return "withdrawals"
}
