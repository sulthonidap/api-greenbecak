package handlers

import (
	"net/http"
	"time"

	"greenbecak-backend/database"
	"greenbecak-backend/models"

	"github.com/gin-gonic/gin"
)

type CreateWithdrawalRequest struct {
	Amount        float64 `json:"amount" binding:"required"`
	BankName      string  `json:"bank_name" binding:"required"`
	AccountNumber string  `json:"account_number" binding:"required"`
	AccountName   string  `json:"account_name" binding:"required"`
	Notes         string  `json:"notes"`
}

type UpdateWithdrawalRequest struct {
	Status     string `json:"status" binding:"required"`
	Notes      string `json:"notes"`
	ApprovedBy string `json:"approved_by"`
	RejectedBy string `json:"rejected_by"`
}

func CreateWithdrawal(c *gin.Context) {
	var req CreateWithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	userID, _ := c.Get("user_id")

	// Check if driver exists
	var driver models.Driver
	if err := db.Where("user_id = ?", userID).First(&driver).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Driver not found"})
		return
	}

	// Check if amount is valid
	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Amount must be greater than 0"})
		return
	}

	// Check if driver has sufficient balance
	// Calculate available balance: total_earnings - approved withdrawals
	var approvedWithdrawals float64
	db.Model(&models.Withdrawal{}).Where("driver_id = ? AND status = ?", driver.ID, "approved").Select("SUM(amount)").Scan(&approvedWithdrawals)

	availableBalance := driver.TotalEarnings - approvedWithdrawals

	if availableBalance < req.Amount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance", "available_balance": availableBalance, "requested_amount": req.Amount})
		return
	}

	withdrawal := models.Withdrawal{
		DriverID:      driver.ID,
		Amount:        req.Amount,
		Status:        models.WithdrawalStatusPending,
		BankName:      req.BankName,
		AccountNumber: req.AccountNumber,
		AccountName:   req.AccountName,
		Notes:         req.Notes,
	}

	if err := db.Create(&withdrawal).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create withdrawal request"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Withdrawal request created successfully",
		"withdrawal": withdrawal,
	})
}

func GetDriverWithdrawals(c *gin.Context) {
	db := database.GetDB()
	userID, _ := c.Get("user_id")

	// Get driver first
	var driver models.Driver
	if err := db.Where("user_id = ?", userID).First(&driver).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Driver not found"})
		return
	}

	var withdrawals []models.Withdrawal
	query := db.Where("driver_id = ?", driver.ID)

	// Add filters
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at DESC").Find(&withdrawals).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch withdrawals"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"withdrawals": withdrawals})
}

func GetWithdrawals(c *gin.Context) {
	db := database.GetDB()

	var withdrawals []models.Withdrawal
	query := db.Preload("Driver.User")

	// Add filters
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if driverID := c.Query("driver_id"); driverID != "" {
		query = query.Where("driver_id = ?", driverID)
	}

	if err := query.Order("created_at DESC").Find(&withdrawals).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch withdrawals"})
		return
	}

			// Calculate available balance for each withdrawal
		for i := range withdrawals {
			var approvedWithdrawals float64
			db.Model(&models.Withdrawal{}).Where("driver_id = ? AND status = ?", withdrawals[i].DriverID, "approved").Select("SUM(amount)").Scan(&approvedWithdrawals)

			// Add calculated fields to response
			withdrawals[i].Driver.AvailableBalance = withdrawals[i].Driver.TotalEarnings - approvedWithdrawals
			withdrawals[i].Driver.CompletedWithdrawals = approvedWithdrawals
		}

	c.JSON(http.StatusOK, gin.H{"withdrawals": withdrawals})
}

func GetWithdrawal(c *gin.Context) {
	withdrawalID := c.Param("id")
	db := database.GetDB()

	var withdrawal models.Withdrawal
	if err := db.Preload("Driver.User").First(&withdrawal, withdrawalID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Withdrawal not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"withdrawal": withdrawal})
}

func UpdateWithdrawal(c *gin.Context) {
	withdrawalID := c.Param("id")
	var req UpdateWithdrawalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()

	var withdrawal models.Withdrawal
	if err := db.Preload("Driver.User").First(&withdrawal, withdrawalID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Withdrawal not found"})
		return
	}

	// Update status
	withdrawal.Status = models.WithdrawalStatus(req.Status)
	now := time.Now()

	switch req.Status {
	case "approved":
		withdrawal.ApprovedAt = &now
		if req.ApprovedBy != "" {
			withdrawal.ApprovedBy = &req.ApprovedBy
		}

				// Update driver's total earnings when withdrawal is approved
		var driver models.Driver
		if err := db.First(&driver, withdrawal.DriverID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find driver"})
			return
		}

		// Check if driver has sufficient balance
		if driver.TotalEarnings < withdrawal.Amount {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient balance", "available_balance": driver.TotalEarnings, "requested_amount": withdrawal.Amount})
			return
		}

		// Reduce total earnings by withdrawal amount
		driver.TotalEarnings -= withdrawal.Amount
		if driver.TotalEarnings < 0 {
			driver.TotalEarnings = 0
		}

		if err := db.Save(&driver).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update driver earnings"})
			return
		}

	case "rejected":
		withdrawal.RejectedAt = &now
		if req.RejectedBy != "" {
			withdrawal.RejectedBy = &req.RejectedBy
		}

	case "completed":
		withdrawal.CompletedAt = &now
		// No additional logic needed for completed status since balance is already deducted on approval
	}

	if req.Notes != "" {
		withdrawal.Notes = req.Notes
	}

	if err := db.Save(&withdrawal).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update withdrawal"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Withdrawal updated successfully",
		"withdrawal": withdrawal,
	})
}

func DeleteWithdrawal(c *gin.Context) {
	withdrawalID := c.Param("id")
	db := database.GetDB()

	if err := db.Delete(&models.Withdrawal{}, withdrawalID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete withdrawal"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Withdrawal deleted successfully"})
}
