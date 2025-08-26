package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"greenbecak-backend/database"
	"greenbecak-backend/models"
)

type PaymentRequest struct {
	OrderID uint   `json:"order_id" binding:"required"`
	Method  string `json:"method" binding:"required"`
	Amount  float64 `json:"amount" binding:"required"`
	Notes   string  `json:"notes"`
}

type PaymentStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// CreatePayment - Membuat pembayaran baru
func CreatePayment(c *gin.Context) {
	var req PaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()

	// Check if order exists
	var order models.Order
	if err := db.First(&order, req.OrderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	// Check if payment already exists for this order
	var existingPayment models.Payment
	if err := db.Where("order_id = ?", req.OrderID).First(&existingPayment).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Payment already exists for this order"})
		return
	}

	// Validate payment method
	if req.Method != string(models.PaymentMethodCash) && 
	   req.Method != string(models.PaymentMethodTransfer) && 
	   req.Method != string(models.PaymentMethodQR) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment method"})
		return
	}

	payment := models.Payment{
		OrderID:    req.OrderID,
		Amount:     req.Amount,
		Method:     models.PaymentMethod(req.Method),
		Status:     models.PaymentStatusPending,
		Reference:  req.Notes, // Using Notes as Reference
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := db.Create(&payment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Payment created successfully",
		"payment": payment,
	})
}

// GetPayments - Mendapatkan daftar pembayaran
func GetPayments(c *gin.Context) {
	db := database.GetDB()
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	var payments []models.Payment
	query := db.Preload("Order")

	// Filter berdasarkan role
	if role == "admin" {
		// Admin bisa lihat semua pembayaran
	} else if role == "driver" {
		// Driver hanya bisa lihat pembayaran order yang dia terima
		query = query.Joins("JOIN orders ON payments.order_id = orders.id").
			Where("orders.driver_id = ?", userID)
	} else {
		// Customer hanya bisa lihat pembayaran order mereka
		query = query.Joins("JOIN orders ON payments.order_id = orders.id").
			Where("orders.customer_id = ?", userID)
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	var total int64
	query.Model(&models.Payment{}).Count(&total)

	if err := query.Offset(offset).Limit(limit).Find(&payments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"payments": payments,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetPayment - Mendapatkan detail pembayaran
func GetPayment(c *gin.Context) {
	db := database.GetDB()
	paymentID := c.Param("id")

	var payment models.Payment
	if err := db.Preload("Order").First(&payment, paymentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
		return
	}

	// Check authorization
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	if role != "admin" {
		if role == "driver" && (payment.Order.DriverID == nil || *payment.Order.DriverID != userID.(uint)) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		if role == "customer" && (payment.Order.CustomerID == nil || *payment.Order.CustomerID != userID.(uint)) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"payment": payment})
}

// UpdatePaymentStatus - Update status pembayaran
func UpdatePaymentStatus(c *gin.Context) {
	var req PaymentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	paymentID := c.Param("id")

	var payment models.Payment
	if err := db.First(&payment, paymentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
		return
	}

	// Validate status
	validStatuses := []string{
		string(models.PaymentStatusPending),
		string(models.PaymentStatusPaid),
		string(models.PaymentStatusFailed),
		string(models.PaymentStatusRefunded),
	}

	isValidStatus := false
	for _, status := range validStatuses {
		if req.Status == status {
			isValidStatus = true
			break
		}
	}

	if !isValidStatus {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment status"})
		return
	}

	payment.Status = models.PaymentStatus(req.Status)
	payment.UpdatedAt = time.Now()

	if err := db.Save(&payment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment"})
		return
	}

	// If payment is paid, update order status
	if req.Status == string(models.PaymentStatusPaid) {
		var order models.Order
		if err := db.First(&order, payment.OrderID).Error; err == nil {
			order.PaymentStatus = "paid"
			order.UpdatedAt = time.Now()
			db.Save(&order)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Payment status updated successfully",
		"payment": payment,
	})
}

// GetPaymentStats - Mendapatkan statistik pembayaran
func GetPaymentStats(c *gin.Context) {
	db := database.GetDB()
	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	var stats struct {
		TotalPayments   int64   `json:"total_payments"`
		TotalAmount     float64 `json:"total_amount"`
		PendingPayments int64   `json:"pending_payments"`
		PaidPayments    int64   `json:"paid_payments"`
		FailedPayments  int64   `json:"failed_payments"`
	}

	query := db.Model(&models.Payment{})

	// Filter berdasarkan role
	if role == "driver" {
		query = query.Joins("JOIN orders ON payments.order_id = orders.id").
			Where("orders.driver_id = ?", userID)
	} else if role == "customer" {
		query = query.Joins("JOIN orders ON payments.order_id = orders.id").
			Where("orders.customer_id = ?", userID)
	}

	// Get total payments and amount
	query.Count(&stats.TotalPayments)
	query.Select("COALESCE(SUM(amount), 0)").Scan(&stats.TotalAmount)

	// Get payments by status
	db.Model(&models.Payment{}).Where("status = ?", models.PaymentStatusPending).Count(&stats.PendingPayments)
	db.Model(&models.Payment{}).Where("status = ?", models.PaymentStatusPaid).Count(&stats.PaidPayments)
	db.Model(&models.Payment{}).Where("status = ?", models.PaymentStatusFailed).Count(&stats.FailedPayments)

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

// ProcessPayment - Memproses pembayaran (simulasi)
func ProcessPayment(c *gin.Context) {
	paymentID := c.Param("id")
	db := database.GetDB()

	var payment models.Payment
	if err := db.First(&payment, paymentID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
		return
	}

	// Simulasi proses pembayaran
	if payment.Status == models.PaymentStatusPending {
		// Simulasi delay processing
		time.Sleep(2 * time.Second)

		// 90% success rate untuk simulasi
		success := time.Now().Unix()%10 < 9
		if success {
			payment.Status = models.PaymentStatusPaid
		} else {
			payment.Status = models.PaymentStatusFailed
		}

		payment.UpdatedAt = time.Now()
		db.Save(&payment)

		// Update order status jika berhasil
		if success {
			var order models.Order
			if err := db.First(&order, payment.OrderID).Error; err == nil {
				order.PaymentStatus = "paid"
				order.UpdatedAt = time.Now()
				db.Save(&order)
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Payment processed successfully",
			"payment": payment,
		})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payment is not in pending status"})
	}
}
