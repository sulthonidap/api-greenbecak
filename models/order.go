package models

import (
	"time"

	"gorm.io/gorm"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusAccepted  OrderStatus = "accepted"
	OrderStatusPickedUp  OrderStatus = "picked_up" // booking: penumpang sudah dijemput
	OrderStatusCompleted OrderStatus = "completed"
	OrderStatusCancelled OrderStatus = "cancelled"
)

// Tipe order
const (
	OrderTypeLangsung = "langsung" // penumpang scan QR di becak (alur lama)
	OrderTypeBooking  = "booking"  // penumpang pesan dari lokasinya, sistem cari driver
)

// Sumber order yang umum. Nilai lain (mis. "instagram") tetap diterima apa adanya.
const (
	OrderSourceQRBecak  = "qr_becak"
	OrderSourceQRHotel  = "qr_hotel"
	OrderSourceQRStatis = "qr_statis"
	OrderSourceWeb      = "web"
)

type Order struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	OrderNumber    string         `json:"order_number" gorm:"unique;not null"`
	CustomerID     *uint          `json:"customer_id"`
	DriverID       *uint          `json:"driver_id"`
	BecakCode      string         `json:"becak_code" gorm:"not null"` // Kode dari sticker barcode
	TariffID       *uint          `json:"tariff_id"`                  // nil untuk order booking (harga per km)
	PickupLocation string         `json:"pickup_location"`            // Nama/alamat titik jemput
	DropLocation   string         `json:"drop_location"`              // Nama/alamat tujuan
	Distance       float64        `json:"distance" gorm:"not null"`
	Price          float64        `json:"price" gorm:"not null"`
	ETA            int            `json:"eta" gorm:"-"` // Estimated Time of Arrival in minutes (calculated field)
	Status         OrderStatus    `json:"status" gorm:"type:enum('pending','accepted','picked_up','completed','cancelled');default:'pending';index:idx_orders_type_status,priority:2"`
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

	// Tipe & asal order
	OrderType string `json:"order_type" gorm:"type:varchar(20);not null;default:'langsung';index:idx_orders_type_status,priority:1"`
	Source    string `json:"source" gorm:"type:varchar(50);index"`  // qr_becak, qr_hotel, qr_statis, web, instagram, ...
	SourceRef string `json:"source_ref" gorm:"type:varchar(100)"`  // kode becak / slug hotel / kampanye
	HotelID   *uint  `json:"hotel_id" gorm:"index"`

	// Koordinat
	PickupLat   *float64 `json:"pickup_lat"`
	PickupLng   *float64 `json:"pickup_lng"`
	DropLat     *float64 `json:"drop_lat"`
	DropLng     *float64 `json:"drop_lng"`
	CustomerLat *float64 `json:"customer_lat"` // posisi GPS penumpang saat memesan (opsional)
	CustomerLng *float64 `json:"customer_lng"`

	// Khusus booking
	VehicleType    string     `json:"vehicle_type" gorm:"type:varchar(30)"`
	DistanceSource string     `json:"distance_source" gorm:"type:varchar(20)"` // "route" (rute jalan) atau "estimate" (garis lurus)
	TrackingToken  *string    `json:"-" gorm:"type:varchar(64);unique"`   // token rahasia untuk halaman status penumpang
	PickedUpAt     *time.Time `json:"picked_up_at"`
	ExpiresAt      *time.Time `json:"expires_at"`
	CancelReason   string     `json:"cancel_reason"`

	// Relationships
	Customer User     `json:"customer,omitempty" gorm:"foreignKey:CustomerID;references:ID"`
	Driver   Driver   `json:"driver,omitempty" gorm:"foreignKey:DriverID;references:ID"`
	Tariff   Tariff   `json:"tariff,omitempty" gorm:"foreignKey:TariffID;references:ID"`
	Payment  *Payment `json:"payment,omitempty" gorm:"foreignKey:OrderID;references:ID"`
	Hotel    *Hotel   `json:"hotel,omitempty" gorm:"foreignKey:HotelID;references:ID"`
}

func (o *Order) TableName() string {
	return "orders"
}
