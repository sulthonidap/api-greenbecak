package handlers

import (
	"fmt"
	"net/http"

	"greenbecak-backend/database"
	"greenbecak-backend/models"
	"greenbecak-backend/utils"

	"github.com/gin-gonic/gin"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
	Phone    string `json:"phone"`
	Address  string `json:"address"`
}

type AuthResponse struct {
	Token   string      `json:"token"`
	User    interface{} `json:"user"`
	Message string      `json:"message"`
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()

	// Check if user exists
	var user models.User
	if err := db.Where("username = ? OR email = ?", req.Username, req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check password
	if !utils.CheckPasswordHash(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check if user is active
	if !user.IsActive {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account is deactivated"})
		return
	}

	// Generate token
	token, err := utils.GenerateToken(user.ID, user.Username, string(user.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// If user is driver, get driver information
	var responseData interface{} = user
	if user.Role == models.RoleDriver {
		var driver models.Driver
		if err := db.Where("user_id = ?", user.ID).First(&driver).Error; err == nil {
			// Create response with both user and driver info
			responseData = gin.H{
				"id":            user.ID,
				"username":      user.Username,
				"email":         user.Email,
				"role":          user.Role,
				"name":          user.Name,
				"phone":         user.Phone,
				"address":       user.Address,
				"is_active":     user.IsActive,
				"created_at":    user.CreatedAt,
				"updated_at":    user.UpdatedAt,
				"driver_id":     driver.ID,
				"driver_code":   driver.DriverCode,
				"vehicle_type":  driver.VehicleType,
				"driver_status": driver.Status,
			}
		}
	}

	response := AuthResponse{
		Token:   token,
		User:    responseData,
		Message: "Login successful",
	}

	c.JSON(http.StatusOK, response)
}

func Register(c *gin.Context) {
	var req RegisterRequest
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
		Role:     models.RoleCustomer, // Default role
	}

	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Generate token
	token, err := utils.GenerateToken(user.ID, user.Username, string(user.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	response := AuthResponse{
		Token:   token,
		User:    user,
		Message: "Registration successful",
	}

	c.JSON(http.StatusCreated, response)
}

func GetProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")
	fmt.Printf("GetProfile called for user_id: %v\n", userID)

	db := database.GetDB()
	var user models.User

	if err := db.First(&user, userID).Error; err != nil {
		fmt.Printf("User not found: %v\n", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	fmt.Printf("User found: ID=%d, Role=%s\n", user.ID, user.Role)

	// If user is driver, get driver information
	var responseData interface{} = user
	if user.Role == models.RoleDriver {
		fmt.Printf("User is driver, looking for driver with user_id: %d\n", user.ID)
		var driver models.Driver
		if err := db.Where("user_id = ?", user.ID).First(&driver).Error; err == nil {
			fmt.Printf("Driver found: ID=%d, Code=%s\n", driver.ID, driver.DriverCode)
			// Create response with both user and driver info
			responseData = gin.H{
				"id":           user.ID,
				"username":     user.Username,
				"email":        user.Email,
				"role":         user.Role,
				"name":         user.Name,
				"phone":        user.Phone,
				"address":      user.Address,
				"is_active":    user.IsActive,
				"created_at":   user.CreatedAt,
				"updated_at":   user.UpdatedAt,
				"driver_id":    driver.ID,
				"driver_code":  driver.DriverCode,
				"vehicle_type": driver.VehicleType,
				"status":       driver.Status,
				"total_trips":  driver.TotalTrips,
				"rating":       driver.Rating,
			}
		} else {
			fmt.Printf("Driver not found for user_id %d: %v\n", user.ID, err)
		}
	} else {
		fmt.Printf("User is not driver, role: %s\n", user.Role)
	}

	fmt.Printf("Final response data: %+v\n", responseData)
	c.JSON(http.StatusOK, gin.H{"user": responseData})
}

func CreateAdminPublic(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()

	// Check if username or email already exists
	var existingUser models.User
	if err := db.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username atau email sudah terdaftar"})
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengenkripsi password"})
		return
	}

	// Create admin user
	user := models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
		Name:     req.Name,
		Phone:    req.Phone,
		Address:  req.Address,
		Role:     models.RoleAdmin,
		IsActive: true,
	}

	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat admin"})
		return
	}

	// Generate token
	token, err := utils.GenerateToken(user.ID, user.Username, string(user.Role))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat token"})
		return
	}

	response := AuthResponse{
		Token:   token,
		User:    user,
		Message: "Registrasi admin berhasil",
	}

	c.JSON(http.StatusCreated, response)
}
