package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"greenbecak-backend/monitoring"
)

func GetAlerts(c *gin.Context) {
	alertManager := monitoring.GetAlertManager()
	alerts := alertManager.GetAlerts()
	
	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
	})
}

func GetActiveAlerts(c *gin.Context) {
	alertManager := monitoring.GetAlertManager()
	alerts := alertManager.GetActiveAlerts()
	
	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
	})
}

func AcknowledgeAlert(c *gin.Context) {
	alertID := c.Param("id")
	alertManager := monitoring.GetAlertManager()
	
	if alertManager.AcknowledgeAlert(alertID) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Alert acknowledged successfully",
		})
	} else {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Alert not found",
		})
	}
}

func CreateAlert(c *gin.Context) {
	var req struct {
		Level   string `json:"level" binding:"required"`
		Message string `json:"message" binding:"required"`
		Service string `json:"service" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	alertManager := monitoring.GetAlertManager()
	alert := alertManager.NewAlert(
		monitoring.AlertLevel(req.Level),
		req.Message,
		req.Service,
	)
	
	c.JSON(http.StatusCreated, gin.H{
		"message": "Alert created successfully",
		"alert":   alert,
	})
}

func ClearOldAlerts(c *gin.Context) {
	alertManager := monitoring.GetAlertManager()
	alertManager.ClearOldAlerts(24 * 60 * 60 * time.Second) // 24 hours
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Old alerts cleared successfully",
	})
}
