package models

import "time"

// PricingSetting menyimpan pengaturan tarif per-km untuk order "Jemput".
// Hanya ada satu baris (ID = 1).
type PricingSetting struct {
	ID uint `json:"id" gorm:"primaryKey"`

	// Mode jemput aktif / tidak. Kalau false, customer hanya bisa pakai mode scan becak.
	Enabled bool `json:"enabled" gorm:"default:false"`

	BaseFare      float64 `json:"base_fare" gorm:"not null;default:5000"`    // Tarif dasar (buka pintu)
	PerKm         float64 `json:"per_km" gorm:"not null;default:4000"`       // Tarif per km
	MinFare       float64 `json:"min_fare" gorm:"not null;default:10000"`    // Tarif minimum
	MaxDistanceKm float64 `json:"max_distance_km" gorm:"not null;default:7"` // Jarak maksimal perjalanan
	RoundTo       float64 `json:"round_to" gorm:"not null;default:1000"`     // Pembulatan ke atas (0 = tanpa pembulatan)

	// Tambahan tarif malam (persen). 0 = nonaktif. Jam format "HH:MM" (WIB).
	NightSurchargePercent float64 `json:"night_surcharge_percent" gorm:"not null;default:0"`
	NightStart            string  `json:"night_start" gorm:"size:5;default:'22:00'"`
	NightEnd              string  `json:"night_end" gorm:"size:5;default:'05:00'"`

	// Area layanan: titik jemput harus berada dalam radius ini dari titik pusat. 0 = tanpa batas.
	ServiceCenterLat float64 `json:"service_center_lat" gorm:"not null;default:-7.7926"`
	ServiceCenterLng float64 `json:"service_center_lng" gorm:"not null;default:110.3658"`
	ServiceRadiusKm  float64 `json:"service_radius_km" gorm:"not null;default:15"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (p *PricingSetting) TableName() string {
	return "pricing_settings"
}

// DefaultPricingSetting mengembalikan nilai awal pengaturan tarif per-km.
func DefaultPricingSetting() PricingSetting {
	return PricingSetting{
		ID:                    1,
		Enabled:               false,
		BaseFare:              5000,
		PerKm:                 4000,
		MinFare:               10000,
		MaxDistanceKm:         7,
		RoundTo:               1000,
		NightSurchargePercent: 0,
		NightStart:            "22:00",
		NightEnd:              "05:00",
		ServiceCenterLat:      -7.7926,
		ServiceCenterLng:      110.3658,
		ServiceRadiusKm:       15,
	}
}
