package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Phone    string `json:"phone"`
	Role     string `json:"role"`
	DriverCode string `json:"driver_code"`
}

type CreateTariffRequest struct {
	Name         string  `json:"name"`
	MinDistance  float64 `json:"min_distance"`
	MaxDistance  float64 `json:"max_distance"`
	Price        float64 `json:"price"`
	Destinations string  `json:"destinations"`
}

func main() {
	baseURL := "http://localhost:8080"
	
	fmt.Println("🌱 Starting to seed database via API...")

	// Wait for server to be ready
	fmt.Println("⏳ Waiting for server to be ready...")
	time.Sleep(2 * time.Second)

	// Seed Admin Users
	admins := []CreateUserRequest{
		{
			Username: "admin_utama",
			Email:    "admin@greenbecak.com",
			Password: "password",
			Name:     "Admin Utama",
			Phone:    "081234567890",
			Role:     "admin",
		},
		{
			Username: "admin_malioboro",
			Email:    "admin.malioboro@greenbecak.com",
			Password: "password",
			Name:     "Admin Malioboro",
			Phone:    "081234567891",
			Role:     "admin",
		},
		{
			Username: "super_admin",
			Email:    "superadmin@greenbecak.com",
			Password: "password",
			Name:     "Super Admin",
			Phone:    "081234567892",
			Role:     "admin",
		},
	}

	for _, admin := range admins {
		createUser(baseURL, admin)
	}

	// Seed Customer Users
	customers := []CreateUserRequest{
		{
			Username: "budi_santoso",
			Email:    "budi.santoso@gmail.com",
			Password: "password",
			Name:     "Budi Santoso",
			Phone:    "081234567903",
			Role:     "customer",
		},
		{
			Username: "siti_nurhaliza",
			Email:    "siti.nurhaliza@gmail.com",
			Password: "password",
			Name:     "Siti Nurhaliza",
			Phone:    "081234567904",
			Role:     "customer",
		},
		{
			Username: "ahmad_rizki",
			Email:    "ahmad.rizki@gmail.com",
			Password: "password",
			Name:     "Ahmad Rizki",
			Phone:    "081234567905",
			Role:     "customer",
		},
		{
			Username: "dewi_sartika",
			Email:    "dewi.sartika@gmail.com",
			Password: "password",
			Name:     "Dewi Sartika",
			Phone:    "081234567906",
			Role:     "customer",
		},
		{
			Username: "rizki_pratama",
			Email:    "rizki.pratama@gmail.com",
			Password: "password",
			Name:     "Rizki Pratama",
			Phone:    "081234567907",
			Role:     "customer",
		},
	}

	for _, customer := range customers {
		createUser(baseURL, customer)
	}

	// Seed Driver Users
	drivers := []CreateUserRequest{
		{
			Username: "driver_seno",
			Email:    "driver.seno@greenbecak.com",
			Password: "password",
			Name:     "Pak Seno",
			Phone:    "08123456789",
			Role:     "driver",
			DriverCode: "BEC001",
		},
		{
			Username: "driver_joko",
			Email:    "driver.joko@greenbecak.com",
			Password: "password",
			Name:     "Pak Joko",
			Phone:    "08123456790",
			Role:     "driver",
			DriverCode: "BEC002",
		},
		{
			Username: "driver_sari",
			Email:    "driver.sari@greenbecak.com",
			Password: "password",
			Name:     "Pak Sari",
			Phone:    "08123456791",
			Role:     "driver",
			DriverCode: "BEC003",
		},
		{
			Username: "driver_rudi",
			Email:    "driver.rudi@greenbecak.com",
			Password: "password",
			Name:     "Pak Rudi",
			Phone:    "08123456792",
			Role:     "driver",
			DriverCode: "BEC004",
		},
		{
			Username: "driver_bambang",
			Email:    "driver.bambang@greenbecak.com",
			Password: "password",
			Name:     "Pak Bambang",
			Phone:    "08123456793",
			Role:     "driver",
			DriverCode: "BEC005",
		},
	}

	for _, driver := range drivers {
		createUser(baseURL, driver)
	}

	// Seed Tariffs
	tariffs := []CreateTariffRequest{
		{
			Name:         "Dekat",
			MinDistance:  0,
			MaxDistance:  3,
			Price:        10000,
			Destinations: "Benteng Vredeburg, Bank Indonesia, Malioboro Mall",
		},
		{
			Name:         "Sedang",
			MinDistance:  3,
			MaxDistance:  7,
			Price:        20000,
			Destinations: "Taman Sari, Alun-Alun Selatan, Keraton Yogyakarta",
		},
		{
			Name:         "Jauh",
			MinDistance:  7,
			MaxDistance:  15,
			Price:        30000,
			Destinations: "Tugu Jogja, Stasiun Lempuyangan, Bandara Adisucipto",
		},
		{
			Name:         "Sangat Jauh",
			MinDistance:  15,
			MaxDistance:  25,
			Price:        40000,
			Destinations: "Candi Prambanan, Candi Borobudur, Gunung Merapi",
		},
		{
			Name:         "Tarif Malam",
			MinDistance:  0,
			MaxDistance:  10,
			Price:        25000,
			Destinations: "Semua destinasi (22:00-06:00)",
		},
		{
			Name:         "Tarif Hujan",
			MinDistance:  0,
			MaxDistance:  10,
			Price:        20000,
			Destinations: "Semua destinasi saat hujan",
		},
		{
			Name:         "Tarif Promo",
			MinDistance:  0,
			MaxDistance:  5,
			Price:        8000,
			Destinations: "Destinasi terbatas untuk pelanggan baru",
		},
		{
			Name:         "Tarif VIP",
			MinDistance:  0,
			MaxDistance:  20,
			Price:        50000,
			Destinations: "Semua destinasi dengan pelayanan premium",
		},
	}

	for _, tariff := range tariffs {
		createTariff(baseURL, tariff)
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

func createUser(baseURL string, user CreateUserRequest) {
	jsonData, _ := json.Marshal(user)
	resp, err := http.Post(baseURL+"/api/admin/users", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ Failed to create user %s: %v\n", user.Name, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 201 {
		fmt.Printf("✅ Created user: %s\n", user.Name)
	} else {
		fmt.Printf("⏭️  User already exists: %s\n", user.Name)
	}
}

func createTariff(baseURL string, tariff CreateTariffRequest) {
	jsonData, _ := json.Marshal(tariff)
	resp, err := http.Post(baseURL+"/api/admin/tariffs", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ Failed to create tariff %s: %v\n", tariff.Name, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 201 {
		fmt.Printf("✅ Created tariff: %s\n", tariff.Name)
	} else {
		fmt.Printf("⏭️  Tariff already exists: %s\n", tariff.Name)
	}
}
