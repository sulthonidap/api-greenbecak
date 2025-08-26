package models

import (
	"time"

	"gorm.io/gorm"
)

type UserRole string

const (
	RoleAdmin    UserRole = "admin"
	RoleCustomer UserRole = "customer"
	RoleDriver   UserRole = "driver"
)

type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Username  string         `json:"username" gorm:"unique;not null"`
	Email     string         `json:"email" gorm:"unique;not null"`
	Password  string         `json:"-" gorm:"not null"`
	Role      UserRole       `json:"role" gorm:"type:enum('admin','customer','driver');default:'customer'"`
	Name      string         `json:"name" gorm:"not null"`
	Phone     string         `json:"phone"`
	Address   string         `json:"address"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Orders []Order `json:"orders,omitempty" gorm:"foreignKey:CustomerID;references:ID"`
	Driver *Driver `json:"driver,omitempty" gorm:"foreignKey:UserID;references:ID"`
}

func (u *User) TableName() string {
	return "users"
}
