package utils

import (
	"greenbecak-backend/database"
	"time"
)

// DatabaseStatus represents the current status of database connection
type DatabaseStatus struct {
	Connected bool      `json:"connected"`
	LastCheck time.Time `json:"last_check"`
	Error     string    `json:"error,omitempty"`
}

var (
	dbStatus      = DatabaseStatus{Connected: false, LastCheck: time.Now()}
	lastDbCheck   = time.Now()
	checkInterval = 30 * time.Second
)

// CheckDatabaseStatus checks if database is connected
func CheckDatabaseStatus() DatabaseStatus {
	now := time.Now()

	// Only check if enough time has passed since last check
	if now.Sub(lastDbCheck) < checkInterval {
		return dbStatus
	}

	lastDbCheck = now
	db := database.GetDB()

	if db == nil {
		dbStatus = DatabaseStatus{
			Connected: false,
			LastCheck: now,
			Error:     "Database not initialized",
		}
		return dbStatus
	}

	// Test database connection
	if err := db.Raw("SELECT 1").Error; err != nil {
		dbStatus = DatabaseStatus{
			Connected: false,
			LastCheck: now,
			Error:     err.Error(),
		}
		return dbStatus
	}

	dbStatus = DatabaseStatus{
		Connected: true,
		LastCheck: now,
	}
	return dbStatus
}

// IsDatabaseConnected returns true if database is connected
func IsDatabaseConnected() bool {
	status := CheckDatabaseStatus()
	return status.Connected
}

// GetDatabaseStatus returns current database status
func GetDatabaseStatus() DatabaseStatus {
	return CheckDatabaseStatus()
}
