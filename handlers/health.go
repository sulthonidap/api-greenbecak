package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"greenbecak-backend/database"
	"greenbecak-backend/monitoring"
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
	// Run comprehensive health check
	status := monitoring.RunHealthCheck()
	
	// Determine overall health
	overallStatus := "healthy"
	httpStatus := http.StatusOK
	
	if !status.Database || !status.API || !status.Memory || !status.Disk {
		overallStatus = "unhealthy"
		httpStatus = http.StatusServiceUnavailable
	}

	health := HealthStatus{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Uptime:    status.Uptime,
		Version:   "1.0.0",
		Services: map[string]string{
			"database": map[bool]string{true: "healthy", false: "unhealthy"}[status.Database],
			"api":      map[bool]string{true: "healthy", false: "unhealthy"}[status.API],
			"memory":   map[bool]string{true: "healthy", false: "unhealthy"}[status.Memory],
			"disk":     map[bool]string{true: "healthy", false: "unhealthy"}[status.Disk],
		},
	}

	c.JSON(httpStatus, health)
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
