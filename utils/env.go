package utils

import (
	"fmt"
	"os"
)

// ValidateEnvironment checks if all required environment variables are set
func ValidateEnvironment() error {
	// Set default values for required environment variables if not set
	setDefaultEnv("DB_HOST", "localhost")
	setDefaultEnv("DB_PORT", "3306")
	setDefaultEnv("DB_USER", "root")
	setDefaultEnv("DB_PASSWORD", "password")
	setDefaultEnv("DB_NAME", "greenbecak_db")
	setDefaultEnv("JWT_SECRET", "your-super-secret-jwt-key-here-change-this-in-production")
	setDefaultEnv("SERVER_PORT", "8080")
	setDefaultEnv("SERVER_MODE", "debug")

	// Validate JWT secret length
	jwtSecret := os.Getenv("JWT_SECRET")
	if len(jwtSecret) < 16 {
		return fmt.Errorf("JWT_SECRET must be at least 16 characters long")
	}

	return nil
}

// setDefaultEnv sets environment variable to default value if not already set
func setDefaultEnv(key, defaultValue string) {
	if os.Getenv(key) == "" {
		os.Setenv(key, defaultValue)
	}
}

// GetEnvWithDefault gets environment variable with default value
func GetEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// IsDevelopment checks if the application is running in development mode
func IsDevelopment() bool {
	return GetEnvWithDefault("SERVER_MODE", "debug") == "debug"
}

// IsProduction checks if the application is running in production mode
func IsProduction() bool {
	return GetEnvWithDefault("SERVER_MODE", "debug") == "release"
}
