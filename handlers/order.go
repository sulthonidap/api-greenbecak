package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"greenbecak-backend/config"
	"greenbecak-backend/database"
	"greenbecak-backend/models"

	"github.com/gin-gonic/gin"
)

type CreateOrderRequest struct {
	CustomerID     uint    `json:"customer_id" binding:"required"`
	TariffID       uint    `json:"tariff_id" binding:"required"`
	PickupLocation string  `json:"pickup_location" binding:"required"`
	DropLocation   string  `json:"drop_location" binding:"required"`
	Distance       float64 `json:"distance" binding:"required"`
	CustomerPhone  string  `json:"customer_phone"`
	CustomerName   string  `json:"customer_name"`
	Notes          string  `json:"notes"`
}

type CreateOrderPublicRequest struct {
	BecakCode     string `json:"becak_code" binding:"required"` // Kode dari sticker barcode
	TariffID      uint   `json:"tariff_id" binding:"required"`
	CustomerPhone string `json:"customer_phone" binding:"required"`
	CustomerName  string `json:"customer_name"`
	Notes         string `json:"notes"`
}

type UpdateOrderRequest struct {
	Status string `json:"status" binding:"required"`
}

type UpdateOrderLocationRequest struct {
	PickupLocation string  `json:"pickup_location" binding:"required"`
	DropLocation   string  `json:"drop_location" binding:"required"`
	Distance       float64 `json:"distance" binding:"required"`
}

// generateOrderNumber creates a unique order number
func generateOrderNumber() string {
	return fmt.Sprintf("ORD-%d", time.Now().Unix())
}

func CreateOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()

	// Get tariff
	var tariff models.Tariff
	if err := db.First(&tariff, req.TariffID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tariff not found"})
		return
	}

	// Calculate price based on distance range
	var price float64
	if req.Distance >= tariff.MinDistance && req.Distance <= tariff.MaxDistance {
		price = tariff.Price
	} else {
		// If distance is outside range, use the tariff price anyway (flat pricing)
		price = tariff.Price
	}

	order := models.Order{
		OrderNumber:    generateOrderNumber(),
		CustomerID:     &req.CustomerID,
		TariffID:       req.TariffID,
		PickupLocation: req.PickupLocation,
		DropLocation:   req.DropLocation,
		Distance:       req.Distance,
		Price:          price,
		Status:         "pending",
		PaymentStatus:  "pending",
		CustomerPhone:  req.CustomerPhone,
		CustomerName:   req.CustomerName,
		Notes:          req.Notes,
	}

	if err := db.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	// Send notification to available drivers
	go sendNewOrderNotification(order)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Order created successfully",
		"order":   order,
	})
}

// CreateOrderPublic allows customers without authentication to create orders
func CreateOrderPublic(c *gin.Context) {
	var req CreateOrderPublicRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()

	// Get tariff
	var tariff models.Tariff
	if err := db.First(&tariff, req.TariffID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tariff not found"})
		return
	}

	// Find driver by becak code
	var driver models.Driver
	if err := db.Where("driver_code = ?", req.BecakCode).First(&driver).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Becak code not found"})
		return
	}

	// Set default customer name if not provided
	if req.CustomerName == "" {
		req.CustomerName = "Customer"
	}

	// Use tariff's max distance as default distance (flat pricing)
	distance := tariff.MaxDistance
	price := tariff.Price

	order := models.Order{
		OrderNumber:    generateOrderNumber(),
		BecakCode:      req.BecakCode,
		DriverID:       &driver.ID,
		TariffID:       req.TariffID,
		PickupLocation: "", // Akan diisi nanti oleh sistem/driver
		DropLocation:   "", // Akan diisi nanti oleh sistem/driver
		Distance:       distance,
		Price:          price,
		Status:         "pending",
		PaymentStatus:  "pending",
		CustomerPhone:  req.CustomerPhone,
		CustomerName:   req.CustomerName,
		Notes:          req.Notes,
	}

	if err := db.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Order created successfully",
		"order":   order,
		"driver": gin.H{
			"name":  driver.Name,
			"phone": driver.Phone,
		},
	})
}

// GetOrderHistory returns orders by customer phone
func GetOrderHistory(c *gin.Context) {
	phone := c.Query("phone")
	if phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "phone is required"})
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	db := database.GetDB()
	var orders []models.Order
	var total int64

	db.Model(&models.Order{}).Where("customer_phone = ?", phone).Count(&total)
	if err := db.Where("customer_phone = ?", phone).Order("created_at DESC").Offset(offset).Limit(limit).Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"orders":     orders,
		"pagination": gin.H{"page": page, "limit": limit, "total": total},
	})
}

func GetOrders(c *gin.Context) {
	db := database.GetDB()
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	var orders []models.Order
	query := db.Preload("Customer").Preload("Driver").Preload("Tariff")

	// Filter based on role
	if role == "admin" {
		// Admin can see all orders
	} else if role == "driver" {
		// Driver can see their own orders
		query = query.Where("driver_id = ?", userID)
	} else {
		// Customer can see their own orders
		query = query.Where("customer_id = ?", userID)
	}

	// Add filters
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"orders": orders})
}

func GetOrder(c *gin.Context) {
	orderID := c.Param("id")
	db := database.GetDB()

	var order models.Order
	if err := db.Preload("Customer").Preload("Driver").Preload("Tariff").First(&order, orderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"order": order})
}

func UpdateOrder(c *gin.Context) {
	orderID := c.Param("id")
	var req UpdateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()

	var order models.Order
	if err := db.First(&order, orderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	// Update status
	order.Status = models.OrderStatus(req.Status)
	now := time.Now()

	switch req.Status {
	case "accepted":
		order.AcceptedAt = &now
		userID, _ := c.Get("user_id")
		driverID := uint(userID.(uint))
		order.DriverID = &driverID
	case "completed":
		order.CompletedAt = &now
	case "cancelled":
		order.CancelledAt = &now
	}

	if err := db.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order updated successfully",
		"order":   order,
	})
}

// UpdateOrderLocation allows driver to update pickup/drop location
func UpdateOrderLocation(c *gin.Context) {
	orderID := c.Param("id")
	var req UpdateOrderLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()

	var order models.Order
	if err := db.First(&order, orderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	// Update location details
	order.PickupLocation = req.PickupLocation
	order.DropLocation = req.DropLocation
	order.Distance = req.Distance

	if err := db.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order location"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order location updated successfully",
		"order":   order,
	})
}

func DeleteOrder(c *gin.Context) {
	orderID := c.Param("id")
	db := database.GetDB()

	if err := db.Delete(&models.Order{}, orderID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order deleted successfully"})
}

// sendNewOrderNotification sends push notification to available drivers
func sendNewOrderNotification(order models.Order) {
	if config.FirebaseService == nil {
		return
	}

	db := database.GetDB()

	// Get all active drivers with FCM tokens
	var drivers []models.Driver
	if err := db.Where("is_active = ? AND fcm_token != ''", true).Find(&drivers).Error; err != nil {
		return
	}

	// Prepare order data for notification
	orderData := map[string]interface{}{
		"id":              order.ID,
		"price":           order.Price,
		"pickup_location": order.PickupLocation,
		"drop_location":   order.DropLocation,
		"distance":        order.Distance,
		"eta":             order.ETA,
	}

	// Send notification to each driver
	for _, driver := range drivers {
		if driver.FCMToken != "" {
			err := config.FirebaseService.SendNewOrderNotification(driver.FCMToken, orderData)
			if err != nil {
				// Log error but continue with other drivers
				fmt.Printf("Failed to send notification to driver %d: %v\n", driver.ID, err)
			}
		}
	}
}
