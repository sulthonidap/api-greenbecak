package handlers

import (
	"net/http"
	"time"

	"greenbecak-backend/database"
	"greenbecak-backend/utils"

	"github.com/gin-gonic/gin"
)

type HealthStatus struct {
	Status    string               `json:"status"`
	Timestamp time.Time            `json:"timestamp"`
	Uptime    string               `json:"uptime"`
	Version   string               `json:"version"`
	Services  map[string]string    `json:"services"`
	Database  utils.DatabaseStatus `json:"database"`
}

var startTime = time.Now()

func HealthCheck(c *gin.Context) {
	// Simple health check - always return healthy if service is running
	health := gin.H{
		"status":    "healthy",
		"timestamp": time.Now(),
		"uptime":    time.Since(startTime).String(),
		"version":   "1.0.0",
		"message":   "Service is running",
	}

	c.JSON(http.StatusOK, health)
}

func ReadinessCheck(c *gin.Context) {
	// Check if application is ready to serve requests
	db := database.GetDB()
	if db == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "not ready",
			"message": "Database not initialized",
		})
		return
	}

	if err := db.Raw("SELECT 1").Error; err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "not ready",
			"message": "Database connection failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "ready",
		"message": "Application is ready to serve requests",
	})
}

func LivenessCheck(c *gin.Context) {
	// Simple liveness check
	c.JSON(http.StatusOK, gin.H{
		"status":  "alive",
		"message": "Application is running",
	})
}

func DatabaseStatusCheck(c *gin.Context) {
	// Check database status
	dbStatus := utils.GetDatabaseStatus()

	httpStatus := http.StatusOK
	if !dbStatus.Connected {
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, gin.H{
		"status":    dbStatus,
		"timestamp": time.Now(),
		"message":   map[bool]string{true: "Database is connected", false: "Database is not connected"}[dbStatus.Connected],
	})
}
