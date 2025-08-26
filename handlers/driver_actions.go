package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"greenbecak-backend/database"
	"greenbecak-backend/models"

	"github.com/gin-gonic/gin"
)

func GetDriverOrders(c *gin.Context) {
	db := database.GetDB()
	userID, _ := c.Get("user_id")

	// Find driver by user_id first
	var driver models.Driver
	if err := db.Where("user_id = ?", userID).First(&driver).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Driver not found"})
		return
	}

	var orders []models.Order
	query := db.Preload("Customer").Preload("Driver").Preload("Driver.User").Preload("Tariff").Preload("Payment").Where("driver_id = ?", driver.ID)

	// Add filters
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at DESC").Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"orders": orders})
}

// GetAvailableOrders - Get orders that are available for drivers to accept (pending orders without driver)
func GetAvailableOrders(c *gin.Context) {
	db := database.GetDB()

	var orders []models.Order
	query := db.Preload("Customer").Preload("Tariff").
		Where("status = ? AND driver_id IS NULL", models.OrderStatusPending)

	// Add distance filter if driver location is provided
	if lat := c.Query("lat"); lat != "" && c.Query("lng") != "" {
		// TODO: Add distance calculation and filtering
		// For now, just get all pending orders
	}

	if err := query.Order("created_at ASC").Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch available orders"})
		return
	}

	// Add distance and ETA calculation for each order
	for i := range orders {
		// TODO: Calculate distance from driver's current location
		// For now, use the order's distance
		if orders[i].Distance > 0 {
			// Estimate ETA: 1 minute per km + 2 minutes base
			orders[i].ETA = int(orders[i].Distance) + 2
		}
	}

	c.JSON(http.StatusOK, gin.H{"orders": orders})
}

// GetOrdersByDriverID - Get orders by specific driver ID (for admin or testing)
func GetOrdersByDriverID(c *gin.Context) {
	db := database.GetDB()
	driverID := c.Param("driver_id")

	// Validate driver exists
	var driver models.Driver
	if err := db.First(&driver, driverID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Driver not found", "driver_id": driverID})
		return
	}

	// Debug: Check all orders in database
	var allOrders []models.Order
	db.Find(&allOrders)
	fmt.Printf("Total orders in database: %d\n", len(allOrders))
	for _, order := range allOrders {
		fmt.Printf("Order ID: %d, DriverID: %v, Status: %s\n", order.ID, order.DriverID, order.Status)
	}

	// Build base query - show all orders for this driver (not just pending)
	query := db.Where("driver_id = ?", driverID)

	// Add status filter if provided
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// Get orders with proper preloading
	var orders []models.Order
	if err := query.
		Preload("Customer").
		Preload("Driver").
		Preload("Driver.User").
		Preload("Tariff").
		Preload("Payment").
		Order("created_at DESC").
		Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}

	fmt.Printf("Orders found for driver %s: %d\n", driverID, len(orders))

	// Note: Data sudah ter-load otomatis dengan Preload di atas

	c.JSON(http.StatusOK, gin.H{"orders": orders})
}

func AcceptOrder(c *gin.Context) {
	orderID := c.Param("id")
	db := database.GetDB()
	userID, _ := c.Get("user_id")

	var order models.Order
	if err := db.First(&order, orderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	// Check if order is pending
	if order.Status != models.OrderStatusPending {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order is not available for acceptance"})
		return
	}

	// Check if driver is available - find driver by user_id
	var driver models.Driver
	if err := db.Where("user_id = ?", userID).First(&driver).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Driver not found"})
		return
	}

	if driver.Status != models.DriverStatusActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Driver is not available"})
		return
	}

	// Update order
	now := time.Now()
	order.Status = models.OrderStatusAccepted
	order.DriverID = &driver.ID
	order.AcceptedAt = &now

	// Update driver status
	driver.Status = models.DriverStatusOnTrip

	if err := db.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to accept order"})
		return
	}

	if err := db.Save(&driver).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update driver status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order accepted successfully",
		"order":   order,
	})
}

func CompleteOrder(c *gin.Context) {
	orderID := c.Param("id")
	db := database.GetDB()
	userID, _ := c.Get("user_id")

	var order models.Order
	if err := db.First(&order, orderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	// Find driver by user_id
	var driver models.Driver
	if err := db.Where("user_id = ?", userID).First(&driver).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Driver not found"})
		return
	}

	// Check if order belongs to driver
	if order.DriverID == nil || *order.DriverID != driver.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Order does not belong to this driver"})
		return
	}

	// Check if order is accepted
	if order.Status != models.OrderStatusAccepted {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order is not in accepted status"})
		return
	}

	// Update order
	now := time.Now()
	order.Status = models.OrderStatusCompleted
	order.CompletedAt = &now

	// Update driver status and earnings
	driver.Status = models.DriverStatusActive
	driver.TotalTrips++
	driver.TotalEarnings += order.Price

	if err := db.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete order"})
		return
	}

	if err := db.Save(&driver).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update driver"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order completed successfully",
		"order":   order,
	})
}

func GetDriverEarnings(c *gin.Context) {
	db := database.GetDB()
	userID, _ := c.Get("user_id")

	fmt.Printf("GetDriverEarnings - userID: %v\n", userID)

	// Find driver by user_id
	var driver models.Driver
	if err := db.Where("user_id = ?", userID).First(&driver).Error; err != nil {
		fmt.Printf("Driver not found for userID: %v, error: %v\n", userID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Driver not found"})
		return
	}

	fmt.Printf("Found driver: ID=%d, Name=%s\n", driver.ID, driver.Name)

	// Get completed orders count
	var completedOrders int64
	db.Model(&models.Order{}).Where("driver_id = ? AND status = ?", driver.ID, models.OrderStatusCompleted).Count(&completedOrders)

	// Get total earnings
	var totalEarnings sql.NullFloat64
	db.Model(&models.Order{}).Where("driver_id = ? AND status = ?", driver.ID, models.OrderStatusCompleted).Select("SUM(price)").Scan(&totalEarnings)



	// Get today's earnings and trips
	today := time.Now().Format("2006-01-02")
	var todayEarnings sql.NullFloat64
	var todayTrips int64
	db.Model(&models.Order{}).
		Where("driver_id = ? AND status = ? AND DATE_FORMAT(completed_at, '%Y-%m-%d') = ?",
			driver.ID, models.OrderStatusCompleted, today).
		Select("SUM(price)").Scan(&todayEarnings)

	db.Model(&models.Order{}).
		Where("driver_id = ? AND status = ? AND DATE_FORMAT(completed_at, '%Y-%m-%d') = ?",
			driver.ID, models.OrderStatusCompleted, today).
		Count(&todayTrips)

	// Get monthly earnings and trips
	currentMonth := time.Now().Format("2006-01")
	var monthlyEarnings sql.NullFloat64
	var monthlyTrips int64
	db.Model(&models.Order{}).
		Where("driver_id = ? AND status = ? AND DATE_FORMAT(completed_at, '%Y-%m') = ?",
			driver.ID, models.OrderStatusCompleted, currentMonth).
		Select("SUM(price)").Scan(&monthlyEarnings)

	db.Model(&models.Order{}).
		Where("driver_id = ? AND status = ? AND DATE_FORMAT(completed_at, '%Y-%m') = ?",
			driver.ID, models.OrderStatusCompleted, currentMonth).
		Count(&monthlyTrips)

		earnings := gin.H{
		"driver_id":   driver.ID,
		"driver_name": driver.Name,
		"total_earnings": func() float64 {
			if totalEarnings.Valid {
				return totalEarnings.Float64
			} else {
				return 0
			}
		}(),
		"today_earnings": func() float64 {
			if todayEarnings.Valid {
				return todayEarnings.Float64
			} else {
				return 0
			}
		}(),
		"today_trips": todayTrips,
		"monthly_earnings": func() float64 {
			if monthlyEarnings.Valid {
				return monthlyEarnings.Float64
			} else {
				return 0
			}
		}(),
		"monthly_trips":    monthlyTrips,
		"completed_orders": completedOrders,
		"total_trips":      driver.TotalTrips,
		"rating":           driver.Rating,
	}

	fmt.Printf("Earnings data: %+v\n", earnings)
	fmt.Printf("Earnings JSON response: %+v\n", gin.H{"earnings": earnings})

	c.JSON(http.StatusOK, gin.H{"earnings": earnings})
}

// DebugAllOrders - Temporary function to see all orders in database
func DebugAllOrders(c *gin.Context) {
	db := database.GetDB()

	var orders []models.Order
	if err := db.Preload("Tariff").Preload("Payment").Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}

	// Count completed orders
	var completedCount int64
	db.Model(&models.Order{}).Where("status = ?", models.OrderStatusCompleted).Count(&completedCount)

	// Count orders by status
	var pendingCount, acceptedCount int64
	db.Model(&models.Order{}).Where("status = ?", "pending").Count(&pendingCount)
	db.Model(&models.Order{}).Where("status = ?", "accepted").Count(&acceptedCount)

	// Manually load driver and customer data
	for i := range orders {
		if orders[i].DriverID != nil {
			var driver models.Driver
			if err := db.Preload("User").First(&driver, *orders[i].DriverID); err == nil {
				orders[i].Driver = driver
			}
		}

		if orders[i].CustomerID != nil {
			var customer models.User
			if err := db.First(&customer, *orders[i].CustomerID); err == nil {
				orders[i].Customer = customer
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"total_orders":     len(orders),
		"completed_orders": completedCount,
		"pending_orders":   pendingCount,
		"accepted_orders":  acceptedCount,
		"orders":           orders,
	})
}

// DebugDrivers - Temporary function to see all drivers in database
func DebugDrivers(c *gin.Context) {
	db := database.GetDB()

	var drivers []models.Driver
	if err := db.Preload("User").Find(&drivers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch drivers"})
		return
	}

	// Count completed orders for each driver
	for i := range drivers {
		var completedCount int64
		var totalEarnings sql.NullFloat64
		db.Model(&models.Order{}).Where("driver_id = ? AND status = ?", drivers[i].ID, models.OrderStatusCompleted).Count(&completedCount)
		db.Model(&models.Order{}).Where("driver_id = ? AND status = ?", drivers[i].ID, models.OrderStatusCompleted).Select("SUM(price)").Scan(&totalEarnings)

		drivers[i].TotalTrips = int(completedCount)
		if totalEarnings.Valid {
			drivers[i].TotalEarnings = totalEarnings.Float64
		} else {
			drivers[i].TotalEarnings = 0
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"total_drivers": len(drivers),
		"drivers":       drivers,
	})
}

// DebugDriverByUserID - Temporary function to find driver by user_id
func DebugDriverByUserID(c *gin.Context) {
	db := database.GetDB()
	userID := c.Param("user_id")

	var driver models.Driver
	if err := db.Preload("User").Where("user_id = ?", userID).First(&driver).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Driver not found for user_id: " + userID})
		return
	}

	// Count completed orders for this driver
	var completedCount int64
	var totalEarnings sql.NullFloat64
	db.Model(&models.Order{}).Where("driver_id = ? AND status = ?", driver.ID, models.OrderStatusCompleted).Count(&completedCount)
	db.Model(&models.Order{}).Where("driver_id = ? AND status = ?", driver.ID, models.OrderStatusCompleted).Select("SUM(price)").Scan(&totalEarnings)

	c.JSON(http.StatusOK, gin.H{
		"driver": gin.H{
			"id":               driver.ID,
			"user_id":          driver.UserID,
			"name":             driver.Name,
			"completed_orders": completedCount,
			"total_earnings": func() float64 {
				if totalEarnings.Valid {
					return totalEarnings.Float64
				} else {
					return 0
				}
			}(),
		},
	})
}
