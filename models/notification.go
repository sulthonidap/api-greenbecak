package models

import (
	"time"

	"gorm.io/gorm"
)

type NotificationPriority string
type NotificationType string

const (
	NotificationPriorityLow    NotificationPriority = "low"
	NotificationPriorityNormal NotificationPriority = "normal"
	NotificationPriorityHigh   NotificationPriority = "high"
	NotificationPriorityUrgent NotificationPriority = "urgent"
	
	NotificationTypeOrder    NotificationType = "order"
	NotificationTypePayment  NotificationType = "payment"
	NotificationTypeSystem   NotificationType = "system"
	NotificationTypePromo    NotificationType = "promo"
	NotificationTypeDriver   NotificationType = "driver"
)

type Notification struct {
	ID        uint                `json:"id" gorm:"primaryKey"`
	UserID    uint                `json:"user_id" gorm:"not null;index"`
	Title     string              `json:"title" gorm:"not null"`
	Message   string              `json:"message" gorm:"not null"`
	Type      NotificationType    `json:"type" gorm:"type:enum('order','payment','system','promo','driver');default:'system'"`
	Priority  NotificationPriority `json:"priority" gorm:"type:enum('low','normal','high','urgent');default:'normal'"`
	IsRead    bool                `json:"is_read" gorm:"default:false"`
	Data      string              `json:"data" gorm:"type:text"` // JSON string untuk data tambahan
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
	DeletedAt gorm.DeletedAt      `json:"-" gorm:"index"`

	// Relationships
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

func (n *Notification) TableName() string {
	return "notifications"
}
