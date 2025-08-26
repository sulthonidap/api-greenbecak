package main

import (
	"fmt"
	"log"

	"greenbecak-backend/database"
	"greenbecak-backend/models"
)

func main() {
	// Connect to database using the same connection as main app
	db := database.GetDB()

	// Auto migrate tables
	err := db.AutoMigrate(
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
		{
			Username: "super_admin",
			Name:     "Super Admin",
			Email:    "superadmin@greenbecak.com",
			Phone:    "081234567892",
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
		{
			Username: "ahmad_rizki",
			Name:     "Ahmad Rizki",
			Email:    "ahmad.rizki@gmail.com",
			Phone:    "081234567905",
			Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi",
			Role:     "customer",
			IsActive: true,
		},
		{
			Username: "dewi_sartika",
			Name:     "Dewi Sartika",
			Email:    "dewi.sartika@gmail.com",
			Phone:    "081234567906",
			Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi",
			Role:     "customer",
			IsActive: true,
		},
		{
			Username: "rizki_pratama",
			Name:     "Rizki Pratama",
			Email:    "rizki.pratama@gmail.com",
			Phone:    "081234567907",
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

	// Seed Driver Users (akan terhubung dengan tabel drivers)
	driverUsers := []models.User{
		{
			Username: "driver_seno",
			Name:     "Pak Seno",
			Email:    "driver.seno@greenbecak.com",
			Phone:    "08123456789",
			Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi",
			Role:     "driver",
			IsActive: true,
		},
		{
			Username: "driver_joko",
			Name:     "Pak Joko",
			Email:    "driver.joko@greenbecak.com",
			Phone:    "08123456790",
			Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi",
			Role:     "driver",
			IsActive: true,
		},
		{
			Username: "driver_sari",
			Name:     "Pak Sari",
			Email:    "driver.sari@greenbecak.com",
			Phone:    "08123456791",
			Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi",
			Role:     "driver",
			IsActive: false,
		},
		{
			Username: "driver_rudi",
			Name:     "Pak Rudi",
			Email:    "driver.rudi@greenbecak.com",
			Phone:    "08123456792",
			Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi",
			Role:     "driver",
			IsActive: true,
		},
		{
			Username: "driver_bambang",
			Name:     "Pak Bambang",
			Email:    "driver.bambang@greenbecak.com",
			Phone:    "08123456793",
			Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi",
			Role:     "driver",
			IsActive: true,
		},
	}

	driverUserIDs := make([]uint, 0)
	for _, driverUser := range driverUsers {
		var existingUser models.User
		if err := db.Where("email = ?", driverUser.Email).First(&existingUser).Error; err != nil {
			if err.Error() == "record not found" {
				db.Create(&driverUser)
				driverUserIDs = append(driverUserIDs, driverUser.ID)
				fmt.Printf("✅ Created driver user: %s\n", driverUser.Name)
			}
		} else {
			driverUserIDs = append(driverUserIDs, existingUser.ID)
			fmt.Printf("⏭️  Driver user already exists: %s\n", driverUser.Name)
		}
	}

	// Seed Drivers (terhubung dengan users)
	drivers := []models.Driver{
		{
			UserID:        &driverUserIDs[0],
			DriverCode:    "DRV-001",
			Name:          "Pak Seno",
			Phone:         "08123456789",
			Email:         "driver.seno@greenbecak.com",
			Address:       "Jl. Malioboro No. 10",
			IDCard:        "1234567890123456",
			VehicleNumber: "AB 1234 XX",
			VehicleType:   models.VehicleTypeBecakManual,
			Status:        models.DriverStatusActive,
			IsActive:      true,
			Rating:        4.5,
			TotalTrips:    150,
			TotalEarnings: 2500000,
		},
		{
			UserID:        &driverUserIDs[1],
			DriverCode:    "DRV-002",
			Name:          "Pak Joko",
			Phone:         "08123456790",
			Email:         "driver.joko@greenbecak.com",
			Address:       "Jl. Malioboro No. 11",
			IDCard:        "1234567890123457",
			VehicleNumber: "AB 1235 XX",
			VehicleType:   models.VehicleTypeBecakMotor,
			Status:        models.DriverStatusActive,
			IsActive:      true,
			Rating:        4.8,
			TotalTrips:    200,
			TotalEarnings: 3000000,
		},
		{
			UserID:        &driverUserIDs[2],
			DriverCode:    "DRV-003",
			Name:          "Pak Sari",
			Phone:         "08123456791",
			Email:         "driver.sari@greenbecak.com",
			Address:       "Jl. Malioboro No. 12",
			IDCard:        "1234567890123458",
			VehicleNumber: "AB 1236 XX",
			VehicleType:   models.VehicleTypeBecakListrik,
			Status:        models.DriverStatusInactive,
			IsActive:      false,
			Rating:        4.2,
			TotalTrips:    100,
			TotalEarnings: 1500000,
		},
		{
			UserID:        &driverUserIDs[3],
			DriverCode:    "DRV-004",
			Name:          "Pak Rudi",
			Phone:         "08123456792",
			Email:         "driver.rudi@greenbecak.com",
			Address:       "Jl. Malioboro No. 13",
			IDCard:        "1234567890123459",
			VehicleNumber: "AB 1237 XX",
			VehicleType:   models.VehicleTypeAndong,
			Status:        models.DriverStatusOnTrip,
			IsActive:      true,
			Rating:        4.7,
			TotalTrips:    180,
			TotalEarnings: 2800000,
		},
		{
			UserID:        &driverUserIDs[4],
			DriverCode:    "DRV-005",
			Name:          "Pak Bambang",
			Phone:         "08123456793",
			Email:         "driver.bambang@greenbecak.com",
			Address:       "Jl. Malioboro No. 14",
			IDCard:        "1234567890123460",
			VehicleNumber: "AB 1238 XX",
			VehicleType:   models.VehicleTypeBecakMotor,
			Status:        models.DriverStatusActive,
			IsActive:      true,
			Rating:        4.6,
			TotalTrips:    220,
			TotalEarnings: 3200000,
		},
	}

	for _, driver := range drivers {
		var existingDriver models.Driver
		if err := db.Where("driver_code = ?", driver.DriverCode).First(&existingDriver).Error; err != nil {
			if err.Error() == "record not found" {
				db.Create(&driver)
				fmt.Printf("✅ Created driver: %s\n", driver.Name)
			}
		} else {
			fmt.Printf("⏭️  Driver already exists: %s\n", driver.Name)
		}
	}

	// Seed Tariffs (flat pricing)
	tariffs := []models.Tariff{
		{
			Name:         "Dekat",
			MinDistance:  0,
			MaxDistance:  3,
			Price:        10000,
			Destinations: "Benteng Vredeburg, Bank Indonesia, Malioboro Mall",
			IsActive:     true,
		},
		{
			Name:         "Sedang",
			MinDistance:  3,
			MaxDistance:  7,
			Price:        20000,
			Destinations: "Taman Sari, Alun-Alun Selatan, Keraton Yogyakarta",
			IsActive:     true,
		},
		{
			Name:         "Jauh",
			MinDistance:  7,
			MaxDistance:  15,
			Price:        30000,
			Destinations: "Tugu Jogja, Stasiun Lempuyangan, Bandara Adisucipto",
			IsActive:     true,
		},
		{
			Name:         "Sangat Jauh",
			MinDistance:  15,
			MaxDistance:  25,
			Price:        40000,
			Destinations: "Candi Prambanan, Candi Borobudur, Gunung Merapi",
			IsActive:     true,
		},
		{
			Name:         "Tarif Malam",
			MinDistance:  0,
			MaxDistance:  10,
			Price:        25000,
			Destinations: "Semua destinasi (22:00-06:00)",
			IsActive:     true,
		},
		{
			Name:         "Tarif Hujan",
			MinDistance:  0,
			MaxDistance:  10,
			Price:        20000,
			Destinations: "Semua destinasi saat hujan",
			IsActive:     true,
		},
		{
			Name:         "Tarif Promo",
			MinDistance:  0,
			MaxDistance:  5,
			Price:        8000,
			Destinations: "Destinasi terbatas untuk pelanggan baru",
			IsActive:     true,
		},
		{
			Name:         "Tarif VIP",
			MinDistance:  0,
			MaxDistance:  20,
			Price:        50000,
			Destinations: "Semua destinasi dengan pelayanan premium",
			IsActive:     true,
		},
	}

	for _, tariff := range tariffs {
		var existingTariff models.Tariff
		if err := db.Where("name = ?", tariff.Name).First(&existingTariff).Error; err != nil {
			if err.Error() == "record not found" {
				db.Create(&tariff)
				fmt.Printf("✅ Created tariff: %s\n", tariff.Name)
			}
		} else {
			fmt.Printf("⏭️  Tariff already exists: %s\n", tariff.Name)
		}
	}

	fmt.Println("\n🎉 Database seeding completed successfully!")
	fmt.Println("\n📋 Login Credentials:")
	fmt.Println("Admin: admin@greenbecak.com / password")
	fmt.Println("Driver: driver.seno@greenbecak.com / password")
	fmt.Println("Customer: budi.santoso@gmail.com / password")
	fmt.Println("\n💰 Tariff System (Flat Pricing):")
	fmt.Println("- Dekat (0-3 km): Rp 10.000")
	fmt.Println("- Sedang (3-7 km): Rp 20.000")
	fmt.Println("- Jauh (7-15 km): Rp 30.000")
	fmt.Println("- Sangat Jauh (15-25 km): Rp 40.000")
	fmt.Println("- Tarif Malam (0-10 km): Rp 25.000")
	fmt.Println("- Tarif Hujan (0-10 km): Rp 20.000")
	fmt.Println("- Tarif Promo (0-5 km): Rp 8.000")
	fmt.Println("- Tarif VIP (0-20 km): Rp 50.000")
	fmt.Println("\n🌐 Test the API at: http://localhost:8080/swagger")
}
