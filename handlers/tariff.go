package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"greenbecak-backend/database"
	"greenbecak-backend/models"
)

type CreateTariffRequest struct {
	Name         string  `json:"name" binding:"required"`
	MinDistance  float64 `json:"min_distance" binding:"required"`
	MaxDistance  float64 `json:"max_distance" binding:"required"`
	Price        float64 `json:"price" binding:"required"`
	Destinations string  `json:"destinations"`
}

type UpdateTariffRequest struct {
	Name         string  `json:"name"`
	MinDistance  float64 `json:"min_distance"`
	MaxDistance  float64 `json:"max_distance"`
	Price        float64 `json:"price"`
	Destinations string  `json:"destinations"`
	IsActive     *bool   `json:"is_active"`
}

type ToggleTariffActiveRequest struct {
	IsActive bool `json:"is_active" binding:"required"`
}

func CreateTariff(c *gin.Context) {
	var req CreateTariffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()

	tariff := models.Tariff{
		Name:         req.Name,
		MinDistance:  req.MinDistance,
		MaxDistance:  req.MaxDistance,
		Price:        req.Price,
		Destinations: req.Destinations,
		IsActive:     true,
	}

	if err := db.Create(&tariff).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tariff"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Tariff created successfully",
		"tariff":  tariff,
	})
}

func GetTariffs(c *gin.Context) {
	db := database.GetDB()

	var tariffs []models.Tariff
	query := db

	// Add filters
	if isActive := c.Query("is_active"); isActive != "" {
		active, _ := strconv.ParseBool(isActive)
		query = query.Where("is_active = ?", active)
	}

	if err := query.Order("min_distance ASC").Find(&tariffs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tariffs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tariffs": tariffs})
}

func GetTariff(c *gin.Context) {
	tariffID := c.Param("id")
	db := database.GetDB()

	var tariff models.Tariff
	if err := db.First(&tariff, tariffID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tariff not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tariff": tariff})
}

func UpdateTariff(c *gin.Context) {
	tariffID := c.Param("id")
	var req UpdateTariffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()

	var tariff models.Tariff
	if err := db.First(&tariff, tariffID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tariff not found"})
		return
	}

	// Update fields
	if req.Name != "" {
		tariff.Name = req.Name
	}
	if req.MinDistance > 0 {
		tariff.MinDistance = req.MinDistance
	}
	if req.MaxDistance > 0 {
		tariff.MaxDistance = req.MaxDistance
	}
	if req.Price > 0 {
		tariff.Price = req.Price
	}
	if req.Destinations != "" {
		tariff.Destinations = req.Destinations
	}
	if req.IsActive != nil {
		tariff.IsActive = *req.IsActive
	}

	if err := db.Save(&tariff).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tariff"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tariff updated successfully",
		"tariff":  tariff,
	})
}

func DeleteTariff(c *gin.Context) {
	tariffID := c.Param("id")
	db := database.GetDB()

	if err := db.Delete(&models.Tariff{}, tariffID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete tariff"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tariff deleted successfully"})
}

// ToggleTariffActive - Mengaktifkan/menonaktifkan tariff
func ToggleTariffActive(c *gin.Context) {
	tariffID := c.Param("id")
	var req ToggleTariffActiveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()

	var tariff models.Tariff
	if err := db.First(&tariff, tariffID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tariff not found"})
		return
	}

	// Toggle active status
	tariff.IsActive = req.IsActive

	if err := db.Save(&tariff).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tariff status"})
		return
	}

	statusText := "activated"
	if !req.IsActive {
		statusText = "deactivated"
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tariff " + statusText + " successfully",
		"tariff":  tariff,
	})
}

// GetTariffsPublic - Mengambil daftar tarif aktif untuk customer tanpa login
func GetTariffsPublic(c *gin.Context) {
	db := database.GetDB()

	var tariffs []models.Tariff
	
	// Hanya ambil tarif yang aktif
	query := db.Where("is_active = ?", true)

	// Add optional filters
	if minDistance := c.Query("min_distance"); minDistance != "" {
		if min, err := strconv.ParseFloat(minDistance, 64); err == nil {
			query = query.Where("min_distance >= ?", min)
		}
	}

	if maxDistance := c.Query("max_distance"); maxDistance != "" {
		if max, err := strconv.ParseFloat(maxDistance, 64); err == nil {
			query = query.Where("max_distance <= ?", max)
		}
	}

	if err := query.Order("min_distance ASC").Find(&tariffs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tariffs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tariffs retrieved successfully",
		"tariffs": tariffs,
	})
}

// GetTariffPublic - Mengambil detail tarif tertentu untuk customer tanpa login
func GetTariffPublic(c *gin.Context) {
	tariffID := c.Param("id")
	db := database.GetDB()

	var tariff models.Tariff
	if err := db.Where("id = ? AND is_active = ?", tariffID, true).First(&tariff).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tariff not found or inactive"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tariff retrieved successfully",
		"tariff":  tariff,
	})
}
