package database

import (
	"fmt"
	"log"
	"time"

	"greenbecak-backend/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() *gorm.DB {
	cfg := config.LoadConfig()

	// Use MySQL only for now
	var err error
	maxRetries := 10
	retryDelay := 5 * time.Second

	// Connect to MySQL with retry mechanism
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&allowNativePasswords=true&allowOldPasswords=true&autocommit=true&sql_mode='STRICT_TRANS_TABLES,NO_ZERO_DATE,NO_ZERO_IN_DATE,ERROR_FOR_DIVISION_BY_ZERO'&tls=false",
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	for i := 0; i < maxRetries; i++ {
		DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})

		if err == nil {
			log.Println("Connected to MySQL database")
			break
		}

		log.Printf("Failed to connect to MySQL database (attempt %d/%d): %v", i+1, maxRetries, err)
		if i < maxRetries-1 {
			log.Printf("Retrying in %v...", retryDelay)
			time.Sleep(retryDelay)
		}
	}

	if err != nil {
		log.Printf("Failed to connect to database after %d attempts: %v", maxRetries, err)
		// Don't exit, let the application start without database
		return nil
	}

	// Configure connection pool
	sqlDB, err := DB.DB()
	if err != nil {
		log.Printf("Failed to get underlying sql.DB: %v", err)
		return nil
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Run migrations (non-blocking)
	go func() {
		if err := RunMigrations(DB); err != nil {
			log.Printf("Warning: Failed to run migrations: %v", err)
		} else {
			log.Println("Database migrations completed successfully")
		}

		// Create indexes (non-blocking)
		if err := CreateIndexes(DB); err != nil {
			log.Printf("Warning: Failed to create some indexes: %v", err)
		}
	}()

	log.Println("Database connected successfully")
	return DB
}

func CloseDB() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			log.Println("Error getting underlying sql.DB:", err)
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

func GetDB() *gorm.DB {
	return DB
}
