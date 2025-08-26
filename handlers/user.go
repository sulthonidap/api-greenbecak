package handlers

import (
	"net/http"
	"strconv"

	"greenbecak-backend/database"
	"greenbecak-backend/models"
	"greenbecak-backend/utils"

	"github.com/gin-gonic/gin"
)

type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
	Phone    string `json:"phone"`
	Address  string `json:"address"`
	Role     string `json:"role" binding:"required,oneof=admin customer driver"`
	// Driver-specific fields (optional, required if role=driver)
	DriverCode    string `json:"driver_code"`
	IDCard        string `json:"id_card"`
	VehicleNumber string `json:"vehicle_number"`
	VehicleType   string `json:"vehicle_type"`
}

type UpdateUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Address  string `json:"address"`
	IsActive *bool  `json:"is_active"`
}

func CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()

	// Check if username or email already exists
	var existingUser models.User
	if err := db.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username or email already exists"})
		return
	}

	// Validate driver-specific fields if role is driver
	if req.Role == "driver" {
		if req.DriverCode == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Driver code is required for driver role"})
			return
		}
		// Check if driver code already exists
		var existingDriver models.Driver
		if err := db.Where("driver_code = ?", req.DriverCode).First(&existingDriver).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Driver code already exists"})
			return
		}
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create user
	user := models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
		Name:     req.Name,
		Phone:    req.Phone,
		Address:  req.Address,
		Role:     models.UserRole(req.Role),
		IsActive: true,
	}

	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// If role is driver, create driver record
	if req.Role == "driver" {
		// Set default vehicle type if not provided
		vehicleType := models.VehicleTypeBecakManual
		if req.VehicleType != "" {
			vehicleType = models.VehicleType(req.VehicleType)
		}

		driver := models.Driver{
			UserID:        &user.ID,
			DriverCode:    req.DriverCode,
			Name:          req.Name,
			Phone:         req.Phone,
			Email:         req.Email,
			Address:       req.Address,
			IDCard:        req.IDCard,
			VehicleNumber: req.VehicleNumber,
			VehicleType:   vehicleType,
			Status:        models.DriverStatusActive,
			IsActive:      true,
		}

		if err := db.Create(&driver).Error; err != nil {
			// Rollback user creation if driver creation fails
			db.Delete(&user)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create driver record"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "User and driver created successfully",
			"user":    user,
			"driver":  driver,
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user":    user,
	})
}

func GetUsers(c *gin.Context) {
	db := database.GetDB()

	var users []models.User
	query := db

	// Add filters
	if role := c.Query("role"); role != "" {
		query = query.Where("role = ?", role)
	}

	if isActive := c.Query("is_active"); isActive != "" {
		active, _ := strconv.ParseBool(isActive)
		query = query.Where("is_active = ?", active)
	}

	if err := query.Preload("Driver").Preload("Orders").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

func GetUser(c *gin.Context) {
	userID := c.Param("id")
	db := database.GetDB()

	var user models.User
	if err := db.Preload("Orders").Preload("Driver").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": user})
}

func UpdateUser(c *gin.Context) {
	userID := c.Param("id")
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()

	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Update fields
	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Address != "" {
		user.Address = req.Address
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if err := db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User updated successfully",
		"user":    user,
	})
}

func DeleteUser(c *gin.Context) {
	userID := c.Param("id")
	db := database.GetDB()

	if err := db.Delete(&models.User{}, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

func ResetUserPassword(c *gin.Context) {
	userID := c.Param("id")

	db := database.GetDB()

	// Check if user exists
	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Generate new password (default: password123)
	newPassword := "password123"

	// Hash the new password
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Update user password
	if err := db.Model(&user).Update("password", hashedPassword).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Password reset successfully",
		"newPassword": newPassword,
	})
}
