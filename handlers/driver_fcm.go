package handlers

import (
	"greenbecak-backend/database"
	"greenbecak-backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Update FCM Token Request
type UpdateFCMTokenRequest struct {
	Token string `json:"token" binding:"required"`
}

// UpdateFCMToken - Update FCM token untuk driver
func UpdateFCMToken(c *gin.Context) {
	db := database.GetDB()
	userID, _ := c.Get("user_id")

	// Find driver by user_id
	var driver models.Driver
	if err := db.Where("user_id = ?", userID).First(&driver).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Driver not found"})
		return
	}

	// Parse request
	var req UpdateFCMTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Update FCM token
	driver.FCMToken = req.Token
	if err := db.Save(&driver).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update FCM token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "FCM token updated successfully",
		"driver_id": driver.ID,
	})
}

// GetFCMToken - Get FCM token untuk driver
func GetFCMToken(c *gin.Context) {
	db := database.GetDB()
	userID, _ := c.Get("user_id")

	// Find driver by user_id
	var driver models.Driver
	if err := db.Where("user_id = ?", userID).First(&driver).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Driver not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"fcm_token": driver.FCMToken,
		"driver_id": driver.ID,
	})
}

// DeleteFCMToken - Delete FCM token untuk driver
func DeleteFCMToken(c *gin.Context) {
	db := database.GetDB()
	userID, _ := c.Get("user_id")

	// Find driver by user_id
	var driver models.Driver
	if err := db.Where("user_id = ?", userID).First(&driver).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Driver not found"})
		return
	}

	// Clear FCM token
	driver.FCMToken = ""
	if err := db.Save(&driver).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete FCM token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "FCM token deleted successfully",
		"driver_id": driver.ID,
	})
}
