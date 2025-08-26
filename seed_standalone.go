package main

import (
	"fmt"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"greenbecak-backend/models"
)

func main() {
	// Try to connect to database
	var db *gorm.DB
	var err error

	// Use SQLite for development
	db, err = gorm.Open(sqlite.Open("greenbecak.db"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Failed to connect to SQLite database:", err)
	}
	fmt.Println("✅ Connected to SQLite database")

	// Auto migrate tables
	err = db.AutoMigrate(
		&models.User{},
		&models.Driver{},
		&models.Order{},
		&models.Payment{},
		&models.Tariff{},
		&models.Notification{},
		&models.Withdrawal{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	fmt.Println("🌱 Starting to seed database...")

	// Seed Admin Users
	admins := []models.User{
		{
			Username: "admin_utama",
			Name:     "Admin Utama",
			Email:    "admin@greenbecak.com",
			Phone:    "081234567890",
			Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // password
			Role:     "admin",
			IsActive: true,
		},
		{
			Username: "admin_malioboro",
			Name:     "Admin Malioboro",
			Email:    "admin.malioboro@greenbecak.com",
			Phone:    "081234567891",
			Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi",
			Role:     "admin",
			IsActive: true,
		},
	}

	for _, admin := range admins {
		var existingUser models.User
		if err := db.Where("email = ?", admin.Email).First(&existingUser).Error; err != nil {
			if err.Error() == "record not found" {
				db.Create(&admin)
				fmt.Printf("✅ Created admin: %s\n", admin.Name)
			}
		} else {
			fmt.Printf("⏭️  Admin already exists: %s\n", admin.Name)
		}
	}

	// Seed Customer Users
	customers := []models.User{
		{
			Username: "budi_santoso",
			Name:     "Budi Santoso",
			Email:    "budi.santoso@gmail.com",
			Phone:    "081234567903",
			Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi",
			Role:     "customer",
			IsActive: true,
		},
		{
			Username: "siti_nurhaliza",
			Name:     "Siti Nurhaliza",
			Email:    "siti.nurhaliza@gmail.com",
			Phone:    "081234567904",
			Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi",
			Role:     "customer",
			IsActive: true,
		},
	}

	for _, customer := range customers {
		var existingUser models.User
		if err := db.Where("email = ?", customer.Email).First(&existingUser).Error; err != nil {
			if err.Error() == "record not found" {
				db.Create(&customer)
				fmt.Printf("✅ Created customer: %s\n", customer.Name)
			}
		} else {
			fmt.Printf("⏭️  Customer already exists: %s\n", customer.Name)
		}
	}

	// Seed Driver Users and Drivers
	driverUsers := []models.User{
		{
			Username: "pak_suwarno",
			Name:     "Pak Suwarno",
			Email:    "driver001@drivers.local",
			Phone:    "081234567901",
			Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi",
			Role:     "driver",
			IsActive: true,
		},
		{
			Username: "pak_suparman",
			Name:     "Pak Suparman",
			Email:    "driver002@drivers.local",
			Phone:    "081234567902",
			Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi",
			Role:     "driver",
			IsActive: true,
		},
		{
			Username: "pak_sutrisno",
			Name:     "Pak Sutrisno",
			Email:    "driver003@drivers.local",
			Phone:    "081234567906",
			Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi",
			Role:     "driver",
			IsActive: true,
		},
	}

	for _, driverUser := range driverUsers {
		var existingUser models.User
		if err := db.Where("email = ?", driverUser.Email).First(&existingUser).Error; err != nil {
			if err.Error() == "record not found" {
				db.Create(&driverUser)
				fmt.Printf("✅ Created driver user: %s\n", driverUser.Name)
			}
		} else {
			fmt.Printf("⏭️  Driver user already exists: %s\n", driverUser.Name)
			driverUser = existingUser
		}

		// Create corresponding driver record
		var existingDriver models.Driver
		if err := db.Where("user_id = ?", driverUser.ID).First(&existingDriver).Error; err != nil {
			if err.Error() == "record not found" {
				driverCode := fmt.Sprintf("BEC%03d", driverUser.ID)
				driver := models.Driver{
					UserID:       &driverUser.ID,
					DriverCode:   driverCode,
					Name:         driverUser.Name,
					Phone:        driverUser.Phone,
					IDCard:       fmt.Sprintf("123456789%03d", driverUser.ID),
					VehicleNumber: fmt.Sprintf("AB %d CD", 1000+driverUser.ID),
					VehicleType:  "becak_manual",
					IsActive:     true,
				}
				db.Create(&driver)
				fmt.Printf("✅ Created driver: %s (Code: %s)\n", driver.Name, driver.DriverCode)
			}
		} else {
			fmt.Printf("⏭️  Driver already exists: %s\n", existingDriver.Name)
		}
	}

	// Seed Tariffs (Flat Pricing)
	tariffs := []models.Tariff{
		{
			Name:         "Jarak Dekat (0-2 km)",
			MinDistance:  0,
			MaxDistance:  2,
			Price:        10000,
			Destinations: "Malioboro, Tugu, Stasiun Tugu",
		},
		{
			Name:         "Jarak Menengah (2-5 km)",
			MinDistance:  2,
			MaxDistance:  5,
			Price:        15000,
			Destinations: "Kraton, Tamansari, Pasar Beringharjo",
		},
		{
			Name:         "Jarak Jauh (5-10 km)",
			MinDistance:  5,
			MaxDistance:  10,
			Price:        25000,
			Destinations: "Bandara Adisucipto, Universitas Gadjah Mada, Malioboro Mall",
		},
		{
			Name:         "Jarak Sangat Jauh (10+ km)",
			MinDistance:  10,
			MaxDistance:  50,
			Price:        35000,
			Destinations: "Candi Prambanan, Candi Borobudur, Gunung Merapi",
		},
	}

	for _, tariff := range tariffs {
		var existingTariff models.Tariff
		if err := db.Where("name = ?", tariff.Name).First(&existingTariff).Error; err != nil {
			if err.Error() == "record not found" {
				db.Create(&tariff)
				fmt.Printf("✅ Created tariff: %s (Rp %d)\n", tariff.Name, tariff.Price)
			}
		} else {
			fmt.Printf("⏭️  Tariff already exists: %s\n", tariff.Name)
		}
	}

	fmt.Println("🎉 Database seeding completed!")
}
