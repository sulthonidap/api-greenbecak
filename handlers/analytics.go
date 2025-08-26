package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"greenbecak-backend/database"
	"greenbecak-backend/models"
)

func GetAnalytics(c *gin.Context) {
	db := database.GetDB()

	// Get total orders
	var totalOrders int64
	db.Model(&models.Order{}).Count(&totalOrders)

	// Get completed orders
	var completedOrders int64
	db.Model(&models.Order{}).Where("status = ?", models.OrderStatusCompleted).Count(&completedOrders)

	// Get total revenue
	var totalRevenue float64
	db.Model(&models.Order{}).Where("status = ?", models.OrderStatusCompleted).Select("SUM(price)").Scan(&totalRevenue)

	// Get total drivers
	var totalDrivers int64
	db.Model(&models.Driver{}).Where("is_active = ?", true).Count(&totalDrivers)

	// Get total customers
	var totalCustomers int64
	db.Model(&models.User{}).Where("role = ? AND is_active = ?", models.RoleCustomer, true).Count(&totalCustomers)

	// Get recent orders
	var recentOrders []models.Order
	db.Preload("Customer").Preload("Driver").Preload("Tariff").Order("created_at DESC").Limit(10).Find(&recentOrders)

	analytics := gin.H{
		"total_orders":      totalOrders,
		"completed_orders":  completedOrders,
		"total_revenue":     totalRevenue,
		"total_drivers":     totalDrivers,
		"total_customers":   totalCustomers,
		"recent_orders":     recentOrders,
		"completion_rate":   float64(completedOrders) / float64(totalOrders) * 100,
	}

	c.JSON(http.StatusOK, gin.H{"analytics": analytics})
}

func GetRevenueAnalytics(c *gin.Context) {
	db := database.GetDB()

	// Get revenue by period
	period := c.Query("period")
	if period == "" {
		period = "month"
	}

	var revenueData []gin.H

	switch period {
	case "week":
		// Last 7 days
		for i := 6; i >= 0; i-- {
			date := time.Now().AddDate(0, 0, -i)
			var dailyRevenue float64
			db.Model(&models.Order{}).
				Where("status = ? AND DATE(created_at) = ?", models.OrderStatusCompleted, date.Format("2006-01-02")).
				Select("COALESCE(SUM(price), 0)").
				Scan(&dailyRevenue)

			revenueData = append(revenueData, gin.H{
				"date":   date.Format("2006-01-02"),
				"revenue": dailyRevenue,
			})
		}
	case "month":
		// Last 30 days
		for i := 29; i >= 0; i-- {
			date := time.Now().AddDate(0, 0, -i)
			var dailyRevenue float64
			db.Model(&models.Order{}).
				Where("status = ? AND DATE(created_at) = ?", models.OrderStatusCompleted, date.Format("2006-01-02")).
				Select("COALESCE(SUM(price), 0)").
				Scan(&dailyRevenue)

			revenueData = append(revenueData, gin.H{
				"date":   date.Format("2006-01-02"),
				"revenue": dailyRevenue,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"revenue_analytics": revenueData})
}

func GetOrderAnalytics(c *gin.Context) {
	db := database.GetDB()

	// Get orders by status
	var ordersByStatus []gin.H
	db.Model(&models.Order{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&ordersByStatus)

	// Get orders by period
	period := c.Query("period")
	if period == "" {
		period = "month"
	}

	var orderData []gin.H

	switch period {
	case "week":
		// Last 7 days
		for i := 6; i >= 0; i-- {
			date := time.Now().AddDate(0, 0, -i)
			var dailyOrders int64
			db.Model(&models.Order{}).
				Where("DATE(created_at) = ?", date.Format("2006-01-02")).
				Count(&dailyOrders)

			orderData = append(orderData, gin.H{
				"date":   date.Format("2006-01-02"),
				"orders": dailyOrders,
			})
		}
	case "month":
		// Last 30 days
		for i := 29; i >= 0; i-- {
			date := time.Now().AddDate(0, 0, -i)
			var dailyOrders int64
			db.Model(&models.Order{}).
				Where("DATE(created_at) = ?", date.Format("2006-01-02")).
				Count(&dailyOrders)

			orderData = append(orderData, gin.H{
				"date":   date.Format("2006-01-02"),
				"orders": dailyOrders,
			})
		}
	}

	analytics := gin.H{
		"orders_by_status": ordersByStatus,
		"order_trends":     orderData,
	}

	c.JSON(http.StatusOK, gin.H{"order_analytics": analytics})
}
