package handlers

import (
	"net/http"
	"time"

	"greenbecak-backend/database"

	"github.com/gin-gonic/gin"
)

type HealthStatus struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Uptime    string            `json:"uptime"`
	Version   string            `json:"version"`
	Services  map[string]string `json:"services"`
}

var startTime = time.Now()

func HealthCheck(c *gin.Context) {
	// Simple health check for container orchestration
	// Always return healthy if the service is running
	health := HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Uptime:    time.Since(startTime).String(),
		Version:   "1.0.0",
		Services: map[string]string{
			"api": "healthy",
		},
	}

	c.JSON(http.StatusOK, health)
}

func ReadinessCheck(c *gin.Context) {
	// Check if application is ready to serve requests
	db := database.GetDB()
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
