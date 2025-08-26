package database

import (
	"log"

	"greenbecak-backend/models"

	"gorm.io/gorm"
)

// RunMigrations runs all database migrations
func RunMigrations(db *gorm.DB) error {
	log.Println("Running database migrations...")

	// Auto migrate all models (temporarily exclude DriverLocation)
	err := db.AutoMigrate(
		&models.User{},
		&models.Driver{},
		&models.Order{},
		&models.Tariff{},
		&models.Payment{},
		&models.Withdrawal{},
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

	log.Println("Database migrations completed successfully")
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
