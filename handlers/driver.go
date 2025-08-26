package handlers

import (
	"net/http"
	"sort"
	"strconv"
	"time"

	"greenbecak-backend/database"
	"greenbecak-backend/models"
	"greenbecak-backend/utils"

	"github.com/gin-gonic/gin"
)

type CreateDriverRequest struct {
	DriverCode    string `json:"driver_code" binding:"required"`
	Name          string `json:"name" binding:"required"`
	Phone         string `json:"phone" binding:"required"`
	Address       string `json:"address"`
	IDCard        string `json:"id_card"`
	VehicleNumber string `json:"vehicle_number"`
	VehicleType   string `json:"vehicle_type"`
}

type UpdateDriverRequest struct {
	Name          string `json:"name"`
	Phone         string `json:"phone"`
	Address       string `json:"address"`
	IDCard        string `json:"id_card"`
	VehicleNumber string `json:"vehicle_number"`
	VehicleType   string `json:"vehicle_type"`
	Status        string `json:"status"`
	IsActive      *bool  `json:"is_active"`
}

func CreateDriver(c *gin.Context) {
	var req CreateDriverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()

	// Check if driver code already exists
	var existingDriver models.Driver
	if err := db.Where("driver_code = ?", req.DriverCode).First(&existingDriver).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Driver code already exists"})
		return
	}

	// Create user account for driver
	username := req.DriverCode
	email := req.DriverCode + "@drivers.local" // Auto-generate email untuk driver
	hashedPwd, _ := utils.HashPassword(req.DriverCode)
	user := models.User{
		Username: username,
		Email:    email,
		Password: hashedPwd,
		Role:     models.RoleDriver,
		Name:     req.Name,
		Phone:    req.Phone,
		Address:  req.Address,
		IsActive: true,
	}
	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create driver user account"})
		return
	}

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
		Email:         email, // Use auto-generated email
		Address:       req.Address,
		IDCard:        req.IDCard,
		VehicleNumber: req.VehicleNumber,
		VehicleType:   vehicleType,
		Status:        models.DriverStatusActive,
		IsActive:      true,
	}

	if err := db.Create(&driver).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create driver"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Driver created successfully",
		"driver":  driver,
	})
}

func GetDrivers(c *gin.Context) {
	db := database.GetDB()

	var drivers []models.Driver
	query := db

	// Add filters
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if isActive := c.Query("is_active"); isActive != "" {
		active, _ := strconv.ParseBool(isActive)
		query = query.Where("is_active = ?", active)
	}

	// Filter out soft-deleted records (GORM automatically handles this with gorm.DeletedAt)
	// But we'll be explicit about it for clarity
	if err := query.Preload("User").Unscoped().Where("deleted_at IS NULL").Find(&drivers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch drivers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"drivers": drivers})
}

func GetDriver(c *gin.Context) {
	driverID := c.Param("id")
	db := database.GetDB()

	var driver models.Driver
	if err := db.Preload("Orders").Preload("Withdrawals").Unscoped().Where("deleted_at IS NULL").First(&driver, driverID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Driver not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"driver": driver})
}

func UpdateDriver(c *gin.Context) {
	driverID := c.Param("id")
	var req UpdateDriverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()

	var driver models.Driver
	if err := db.Unscoped().Where("deleted_at IS NULL").First(&driver, driverID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Driver not found"})
		return
	}

	// Update fields
	if req.Name != "" {
		driver.Name = req.Name
	}
	if req.Phone != "" {
		driver.Phone = req.Phone
	}
	if req.Address != "" {
		driver.Address = req.Address
	}
	if req.IDCard != "" {
		driver.IDCard = req.IDCard
	}
	if req.VehicleNumber != "" {
		driver.VehicleNumber = req.VehicleNumber
	}
	if req.VehicleType != "" {
		driver.VehicleType = models.VehicleType(req.VehicleType)
	}
	if req.Status != "" {
		driver.Status = models.DriverStatus(req.Status)
	}
	if req.IsActive != nil {
		driver.IsActive = *req.IsActive
	}

	if err := db.Save(&driver).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update driver"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Driver updated successfully",
		"driver":  driver,
	})
}

func DeleteDriver(c *gin.Context) {
	driverID := c.Param("id")
	db := database.GetDB()

	if err := db.Delete(&models.Driver{}, driverID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete driver"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Driver deleted successfully"})
}

func GetDriverPerformance(c *gin.Context) {
	driverID := c.Param("id")
	db := database.GetDB()

	var driver models.Driver
	if err := db.Unscoped().Where("deleted_at IS NULL").First(&driver, driverID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Driver not found"})
		return
	}

	// Get completed orders count
	var completedOrders int64
	db.Model(&models.Order{}).Where("driver_id = ? AND status = ?", driverID, models.OrderStatusCompleted).Count(&completedOrders)

	// Get total earnings
	var totalEarnings float64
	db.Model(&models.Order{}).Where("driver_id = ? AND status = ?", driverID, models.OrderStatusCompleted).Select("SUM(price)").Scan(&totalEarnings)

	// Get average rating
	var avgRating float64
	db.Model(&models.Order{}).Where("driver_id = ? AND status = ?", driverID, models.OrderStatusCompleted).Select("AVG(rating)").Scan(&avgRating)

	performance := gin.H{
		"driver_id":        driver.ID,
		"driver_name":      driver.Name,
		"completed_orders": completedOrders,
		"total_earnings":   totalEarnings,
		"average_rating":   avgRating,
		"total_trips":      driver.TotalTrips,
		"rating":           driver.Rating,
	}

	c.JSON(http.StatusOK, gin.H{"performance": performance})
}

func GetDriverFinancialData(c *gin.Context) {
	db := database.GetDB()

	// Get all drivers with their financial data (excluding soft-deleted)
	var drivers []models.Driver
	if err := db.Preload("User").Unscoped().Where("deleted_at IS NULL").Find(&drivers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch drivers"})
		return
	}

	var driverFinancialData []gin.H

	for _, driver := range drivers {
		// Get completed orders count and total earnings
		var completedOrders int64
		var totalEarnings float64
		db.Model(&models.Order{}).Where("driver_id = ? AND status = ?", driver.ID, models.OrderStatusCompleted).Count(&completedOrders)
		db.Model(&models.Order{}).Where("driver_id = ? AND status = ?", driver.ID, models.OrderStatusCompleted).Select("SUM(price)").Scan(&totalEarnings)

		// Get monthly earnings
		currentMonth := time.Now().Format("2006-01")
		var monthlyEarnings float64
		db.Model(&models.Order{}).
			Where("driver_id = ? AND status = ? AND DATE_FORMAT(completed_at, '%Y-%m') = ?",
				driver.ID, models.OrderStatusCompleted, currentMonth).
			Select("SUM(price)").
			Scan(&monthlyEarnings)

		// Get last month earnings
		lastMonth := time.Now().AddDate(0, -1, 0).Format("2006-01")
		var lastMonthEarnings float64
		db.Model(&models.Order{}).
			Where("driver_id = ? AND status = ? AND DATE_FORMAT(completed_at, '%Y-%m') = ?",
				driver.ID, models.OrderStatusCompleted, lastMonth).
			Select("SUM(price)").
			Scan(&lastMonthEarnings)

		// Get withdrawal data
		var pendingWithdrawals float64
		var completedWithdrawals float64
		db.Model(&models.Withdrawal{}).Where("driver_id = ? AND status = ?", driver.ID, "pending").Select("SUM(amount)").Scan(&pendingWithdrawals)
		db.Model(&models.Withdrawal{}).Where("driver_id = ? AND status = ?", driver.ID, "completed").Select("SUM(amount)").Scan(&completedWithdrawals)

		// Get last withdrawal
		var lastWithdrawal models.Withdrawal
		err := db.Where("driver_id = ?", driver.ID).Order("created_at DESC").First(&lastWithdrawal).Error
		if err != nil {
			// If no withdrawal found, create empty withdrawal data
			lastWithdrawal = models.Withdrawal{
				Amount: 0,
				Status: "pending",
			}
		}

		// Calculate average per trip
		var averagePerTrip float64
		if completedOrders > 0 {
			averagePerTrip = totalEarnings / float64(completedOrders)
		}

		// Get monthly earnings for last 6 months
		var monthlyEarningsData []gin.H
		for i := 5; i >= 0; i-- {
			month := time.Now().AddDate(0, -i, 0).Format("2006-01")
			var monthEarnings float64
			var monthTrips int64
			db.Model(&models.Order{}).
				Where("driver_id = ? AND status = ? AND DATE_FORMAT(completed_at, '%Y-%m') = ?",
					driver.ID, models.OrderStatusCompleted, month).
				Select("SUM(price)").Scan(&monthEarnings)
			db.Model(&models.Order{}).
				Where("driver_id = ? AND status = ? AND DATE_FORMAT(completed_at, '%Y-%m') = ?",
					driver.ID, models.OrderStatusCompleted, month).
				Count(&monthTrips)

			monthlyEarningsData = append(monthlyEarningsData, gin.H{
				"month":    time.Now().AddDate(0, -i, 0).Format("Jan"),
				"earnings": monthEarnings,
				"trips":    monthTrips,
			})
		}

		// Get recent transactions (last 5 orders and withdrawals)
		var recentTransactions []gin.H

		// Recent orders
		var recentOrders []models.Order
		db.Where("driver_id = ? AND status = ?", driver.ID, models.OrderStatusCompleted).
			Order("completed_at DESC").Limit(3).Find(&recentOrders)

		for _, order := range recentOrders {
			// Handle nil CompletedAt
			var completedAt time.Time
			if order.CompletedAt != nil {
				completedAt = *order.CompletedAt
			} else {
				completedAt = time.Now() // Fallback to current time
			}

			recentTransactions = append(recentTransactions, gin.H{
				"type":        "trip",
				"amount":      order.Price,
				"date":        completedAt,
				"description": "Trip dari " + order.PickupLocation + " ke " + order.DropLocation,
			})
		}

		// Recent withdrawals
		var recentWithdrawals []models.Withdrawal
		db.Where("driver_id = ?", driver.ID).
			Order("created_at DESC").Limit(2).Find(&recentWithdrawals)

		for _, withdrawal := range recentWithdrawals {
			recentTransactions = append(recentTransactions, gin.H{
				"type":        "withdrawal",
				"amount":      -withdrawal.Amount,
				"date":        withdrawal.CreatedAt,
				"description": "Penarikan ke " + withdrawal.BankName,
			})
		}

		// Sort transactions by date
		sort.Slice(recentTransactions, func(i, j int) bool {
			// Handle both *time.Time and time.Time
			var dateI, dateJ time.Time

			if datePtr, ok := recentTransactions[i]["date"].(*time.Time); ok {
				dateI = *datePtr
			} else if dateVal, ok := recentTransactions[i]["date"].(time.Time); ok {
				dateI = dateVal
			} else {
				// Fallback to current time if type assertion fails
				dateI = time.Now()
			}

			if datePtr, ok := recentTransactions[j]["date"].(*time.Time); ok {
				dateJ = *datePtr
			} else if dateVal, ok := recentTransactions[j]["date"].(time.Time); ok {
				dateJ = dateVal
			} else {
				// Fallback to current time if type assertion fails
				dateJ = time.Now()
			}

			return dateI.After(dateJ)
		})

		// Limit to 5 transactions
		if len(recentTransactions) > 5 {
			recentTransactions = recentTransactions[:5]
		}

		driverFinancialData = append(driverFinancialData, gin.H{
			"id":                    driver.ID,
			"name":                  driver.Name,
			"current_balance":       totalEarnings - completedWithdrawals,
			"total_earnings":        totalEarnings,
			"this_month_earnings":   monthlyEarnings,
			"last_month_earnings":   lastMonthEarnings,
			"pending_withdrawals":   pendingWithdrawals,
			"completed_withdrawals": completedWithdrawals,
			"total_trips":           completedOrders,
			"average_per_trip":      averagePerTrip,
			"rating":                driver.Rating,
			"status":                driver.Status,
			"last_withdrawal": gin.H{
				"amount": lastWithdrawal.Amount,
				"date":   lastWithdrawal.CreatedAt,
				"status": lastWithdrawal.Status,
			},
			"monthly_earnings":    monthlyEarningsData,
			"recent_transactions": recentTransactions,
		})
	}

	c.JSON(http.StatusOK, gin.H{"drivers": driverFinancialData})
}
