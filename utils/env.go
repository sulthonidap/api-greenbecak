package utils

import (
	"fmt"
	"os"
)

// ValidateEnvironment checks if all required environment variables are set
func ValidateEnvironment() error {
	required := []string{
		"DB_HOST",
		"DB_USER",
		"DB_PASSWORD",
		"DB_NAME",
		"JWT_SECRET",
	}

	for _, env := range required {
		if os.Getenv(env) == "" {
			return fmt.Errorf("required environment variable %s is not set", env)
		}
	}

	// Validate JWT secret length
	jwtSecret := os.Getenv("JWT_SECRET")
	if len(jwtSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters long")
	}

	return nil
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
