package monitoring

import (
	"log"
	"time"

	"greenbecak-backend/database"
)

type HealthStatus struct {
	Database  bool      `json:"database"`
	API       bool      `json:"api"`
	Memory    bool      `json:"memory"`
	Disk      bool      `json:"disk"`
	LastCheck time.Time `json:"last_check"`
	Uptime    string    `json:"uptime"`
}

var (
	healthStatus HealthStatus
	startTime    = time.Now()
)

// CheckDatabaseHealth checks database connectivity
func CheckDatabaseHealth() bool {
	db := database.GetDB()
	if db == nil {
		log.Printf("Database connection not initialized")
		return false
	}
	if err := db.Raw("SELECT 1").Error; err != nil {
		log.Printf("Database health check failed: %v", err)
		return false
	}
	return true
}

// CheckMemoryHealth checks memory usage
func CheckMemoryHealth() bool {
	// This is a simple check - in production you'd want more sophisticated monitoring
	// For now, we'll just return true
	return true
}

// CheckDiskHealth checks disk space
func CheckDiskHealth() bool {
	// This is a simple check - in production you'd want more sophisticated monitoring
	// For now, we'll just return true
	return true
}

// RunHealthCheck runs all health checks
func RunHealthCheck() HealthStatus {
	healthStatus.Database = CheckDatabaseHealth()
	healthStatus.API = true // API is running if this function is called
	healthStatus.Memory = CheckMemoryHealth()
	healthStatus.Disk = CheckDiskHealth()
	healthStatus.LastCheck = time.Now()
	healthStatus.Uptime = time.Since(startTime).String()

	return healthStatus
}

// IsHealthy checks if all systems are healthy
func IsHealthy() bool {
	status := RunHealthCheck()
	return status.Database && status.API && status.Memory && status.Disk
}

// GetHealthStatus returns current health status
func GetHealthStatus() HealthStatus {
	return healthStatus
}
