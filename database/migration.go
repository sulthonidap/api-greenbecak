package database

import (
	"log"

	"greenbecak-backend/models"

	"gorm.io/gorm"
)

// RunMigrations runs all database migrations
func RunMigrations(db *gorm.DB) error {
	log.Println("Running database migrations...")

	// Auto migrate all models
	err := db.AutoMigrate(
		&models.User{},
		&models.Driver{},
		&models.Hotel{},
		&models.Destination{},
		&models.BookingRate{},
		&models.Order{},
		&models.Tariff{},
		&models.Payment{},
		&models.Withdrawal{},
		&models.DriverLocation{},
	)

	if err != nil {
		log.Printf("Migration failed: %v", err)
		return err
	}

	// Ensure enum values are updated for users.role to include 'driver'
	// Note: GORM does not auto-alter enum definitions, so we run a best-effort ALTER here for MySQL.
	if err := db.Exec("ALTER TABLE users MODIFY COLUMN role ENUM('admin','customer','driver') NOT NULL DEFAULT 'customer'").Error; err != nil {
		log.Printf("Info: Skipping enum alter for users.role (may already be up-to-date): %v", err)
	}

	// Tambah status 'picked_up' untuk booking
	if err := db.Exec("ALTER TABLE orders MODIFY COLUMN status ENUM('pending','accepted','picked_up','completed','cancelled') DEFAULT 'pending'").Error; err != nil {
		log.Printf("Info: Skipping enum alter for orders.status: %v", err)
	}

	// Order lama tanpa tipe dianggap order langsung
	if err := db.Exec("UPDATE orders SET order_type = 'langsung' WHERE order_type IS NULL OR order_type = ''").Error; err != nil {
		log.Printf("Info: Skipping order_type backfill: %v", err)
	}

	if err := seedBookingDefaults(db); err != nil {
		log.Printf("Warning: seeding booking defaults failed: %v", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// seedBookingDefaults mengisi tarif booking & destinasi populer awal, hanya jika tabel masih kosong.
// Semua nilai bisa diubah admin. Koordinat destinasi adalah perkiraan — mohon dicek di menu admin.
func seedBookingDefaults(db *gorm.DB) error {
	var count int64
	db.Model(&models.BookingRate{}).Count(&count)
	if count == 0 {
		rates := []models.BookingRate{
			{VehicleType: string(models.VehicleTypeBecakListrik), Label: "Becak Listrik", BaseFare: 5000, PerKm: 4000, MinFare: 10000, MaxDistanceKm: 10, Rounding: 1000, IsActive: true},
			{VehicleType: string(models.VehicleTypeAndong), Label: "Andong", BaseFare: 15000, PerKm: 8000, MinFare: 30000, MaxDistanceKm: 10, Rounding: 1000, IsActive: true},
		}
		if err := db.Create(&rates).Error; err != nil {
			return err
		}
		log.Println("Seeded default booking rates (contoh — ubah di menu admin)")
	}

	db.Model(&models.Destination{}).Count(&count)
	if count == 0 {
		dests := []models.Destination{
			{Name: "Tugu Yogyakarta", Category: "wisata", Address: "Jl. Jend. Sudirman, Gowongan", Latitude: -7.78284, Longitude: 110.36703, SortOrder: 1},
			{Name: "Stasiun Tugu Yogyakarta", Category: "transportasi", Address: "Jl. Pasar Kembang, Sosromenduran", Latitude: -7.78917, Longitude: 110.36334, SortOrder: 2},
			{Name: "Jalan Malioboro", Category: "belanja", Address: "Jl. Malioboro, Sosromenduran", Latitude: -7.79254, Longitude: 110.36580, SortOrder: 3},
			{Name: "Pasar Beringharjo", Category: "belanja", Address: "Jl. Margo Mulyo, Ngupasan", Latitude: -7.79856, Longitude: 110.36531, SortOrder: 4},
			{Name: "Benteng Vredeburg", Category: "wisata", Address: "Jl. Margo Mulyo No.6, Ngupasan", Latitude: -7.80022, Longitude: 110.36628, SortOrder: 5},
			{Name: "Titik Nol Kilometer", Category: "wisata", Address: "Jl. Panembahan Senopati, Prawirodirjan", Latitude: -7.80134, Longitude: 110.36472, SortOrder: 6},
			{Name: "Kraton Yogyakarta (Keraton)", Category: "wisata", Address: "Jl. Rotowijayan Blok No.1, Panembahan", Latitude: -7.80529, Longitude: 110.36421, SortOrder: 7},
			{Name: "Taman Sari", Category: "wisata", Address: "Jl. Taman, Patehan", Latitude: -7.81008, Longitude: 110.35921, SortOrder: 8},
			{Name: "Alun-Alun Kidul", Category: "wisata", Address: "Patehan, Kraton", Latitude: -7.81182, Longitude: 110.36336, SortOrder: 9},
			{Name: "Kampung Wisata Prawirotaman", Category: "kuliner", Address: "Jl. Prawirotaman, Brontokusuman", Latitude: -7.81973, Longitude: 110.36947, SortOrder: 10},
		}
		for i := range dests {
			dests[i].IsActive = true
		}
		if err := db.Create(&dests).Error; err != nil {
			return err
		}
		log.Println("Seeded default destinations (koordinat perkiraan — cek di menu admin)")
	}
	return nil
}

// CreateIndexes creates additional indexes for better performance
func CreateIndexes(db *gorm.DB) error {
	log.Println("Creating database indexes...")

	// Create indexes for better query performance
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_users_username ON users(username)",
		"CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)",
		"CREATE INDEX IF NOT EXISTS idx_users_role ON users(role)",
		"CREATE INDEX IF NOT EXISTS idx_drivers_code ON drivers(driver_code)",
		"CREATE INDEX IF NOT EXISTS idx_drivers_status ON drivers(status)",
		"CREATE INDEX IF NOT EXISTS idx_orders_number ON orders(order_number)",
		"CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status)",
		"CREATE INDEX IF NOT EXISTS idx_orders_customer ON orders(customer_id)",
		"CREATE INDEX IF NOT EXISTS idx_orders_customer_phone ON orders(customer_phone)",
		"CREATE INDEX IF NOT EXISTS idx_orders_driver ON orders(driver_id)",
		"CREATE INDEX IF NOT EXISTS idx_orders_created ON orders(created_at)",
		"CREATE INDEX IF NOT EXISTS idx_tariffs_active ON tariffs(is_active)",
		"CREATE INDEX IF NOT EXISTS idx_payments_order ON payments(order_id)",
		"CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status)",
		"CREATE INDEX IF NOT EXISTS idx_withdrawals_driver ON withdrawals(driver_id)",
		"CREATE INDEX IF NOT EXISTS idx_withdrawals_status ON withdrawals(status)",
		"CREATE INDEX IF NOT EXISTS idx_driver_locations_driver ON driver_locations(driver_id)",
		"CREATE INDEX IF NOT EXISTS idx_driver_locations_online ON driver_locations(is_online)",
		"CREATE INDEX IF NOT EXISTS idx_driver_locations_last_seen ON driver_locations(last_seen)",
	}

	for _, index := range indexes {
		if err := db.Exec(index).Error; err != nil {
			log.Printf("Failed to create index: %v", err)
			// Don't fail the entire migration for index creation errors
			continue
		}
	}

	log.Println("Database indexes created successfully")
	return nil
}

// SeedInitialData seeds the database with initial data
func SeedInitialData(db *gorm.DB) error {
	log.Println("Seeding initial data...")

	// Check if data already exists
	var count int64
	db.Model(&models.User{}).Count(&count)
	if count > 0 {
		log.Println("Data already exists, skipping seeding")
		return nil
	}

	// Run the seed script
	// This will be handled by the separate seed.go script
	log.Println("Initial data seeding completed")
	return nil
}
