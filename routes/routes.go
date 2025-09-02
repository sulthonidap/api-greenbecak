package routes

import (
	"greenbecak-backend/handlers"
	"greenbecak-backend/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB) {
	// Health check endpoints
	r.GET("/health", handlers.HealthCheck)
	r.GET("/ready", handlers.ReadinessCheck)
	r.GET("/live", handlers.LivenessCheck)

	// Metrics endpoints
	r.GET("/metrics", handlers.GetMetrics)
	r.POST("/metrics/reset", handlers.ResetMetrics)

	// Alert endpoints
	r.GET("/alerts", handlers.GetAlerts)
	r.GET("/alerts/active", handlers.GetActiveAlerts)
	r.POST("/alerts", handlers.CreateAlert)
	r.PUT("/alerts/:id/acknowledge", handlers.AcknowledgeAlert)
	r.DELETE("/alerts/old", handlers.ClearOldAlerts)

	// API routes
	api := r.Group("/api")
	{
		// Public routes
		// Auth routes
		auth := api.Group("/auth")
		auth.Use(middleware.StrictRateLimit())
		{
			auth.POST("/login", handlers.Login)
			auth.POST("/register", handlers.Register)
		}

		// Public location endpoints
		location := api.Group("/location")
		{
			location.GET("/drivers/nearby", handlers.GetNearbyDrivers)
			location.GET("/drivers/:id", handlers.GetDriverLocation)
			location.GET("/routes/:order_id", handlers.GetDriverRoute)
		}

		// Public order endpoints (no auth)
		api.POST("/orders/public", handlers.CreateOrderPublic)
		api.GET("/orders/history", handlers.GetOrderHistory)

		// Public admin creation endpoint (no auth required)
		api.POST("/admin/public", handlers.CreateAdminPublic)

		// Public tariff endpoints (no auth required)
		api.GET("/tariffs/public", handlers.GetTariffsPublic)
		api.GET("/tariffs/public/:id", handlers.GetTariffPublic)

		// Public driver orders endpoint (for testing) - must be before protected routes
		api.GET("/driver/:driver_id/orders", handlers.GetOrdersByDriverID)

		// Debug endpoint to see all orders
		api.GET("/debug/orders", handlers.DebugAllOrders)
		// Debug endpoint to see all drivers
		api.GET("/debug/drivers", handlers.DebugDrivers)
		// Debug endpoint to find driver by user_id
		api.GET("/debug/driver/user/:user_id", handlers.DebugDriverByUserID)

		// Protected routes
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			// Profile
			protected.GET("/profile", handlers.GetProfile)

			// Orders
			orders := protected.Group("/orders")
			{
				orders.POST("/", handlers.CreateOrder)
				orders.GET("/", handlers.GetOrders)
				orders.GET("/:id", handlers.GetOrder)
				orders.PUT("/:id", handlers.UpdateOrder)
				orders.PUT("/:id/location", handlers.UpdateOrderLocation)
				orders.DELETE("/:id", handlers.DeleteOrder)
			}

			// Tariffs (public for customers, admin only for management)
			tariffs := protected.Group("/tariffs")
			{
				tariffs.GET("/", handlers.GetTariffs)
				tariffs.GET("/:id", handlers.GetTariff)
			}

			// Payments
			payments := protected.Group("/payments")
			{
				payments.POST("/", handlers.CreatePayment)
				payments.GET("/", handlers.GetPayments)
				payments.GET("/:id", handlers.GetPayment)
				payments.PUT("/:id/status", handlers.UpdatePaymentStatus)
				payments.POST("/:id/process", handlers.ProcessPayment)
				payments.GET("/stats", handlers.GetPaymentStats)
			}

			// Notifications
			notifications := protected.Group("/notifications")
			{
				notifications.GET("/", handlers.GetNotifications)
				notifications.GET("/:id", handlers.GetNotification)
				notifications.PUT("/:id/read", handlers.MarkNotificationAsRead)
				notifications.PUT("/read-all", handlers.MarkAllNotificationsAsRead)
				notifications.DELETE("/:id", handlers.DeleteNotification)
				notifications.GET("/stats", handlers.GetNotificationStats)
			}
		}

		// Admin only routes
		admin := api.Group("/admin")
		admin.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
		{
			// User management
			users := admin.Group("/users")
			{
				users.POST("/", handlers.CreateUser)
				users.GET("/", handlers.GetUsers)
				users.GET("/:id", handlers.GetUser)
				users.PUT("/:id", handlers.UpdateUser)
				users.DELETE("/:id", handlers.DeleteUser)
				users.POST("/:id/reset-password", handlers.ResetUserPassword)
			}

			// Driver management
			drivers := admin.Group("/drivers")
			{
				drivers.POST("/", handlers.CreateDriver)
				drivers.GET("/", handlers.GetDrivers)
				drivers.GET("/:id", handlers.GetDriver)
				drivers.PUT("/:id", handlers.UpdateDriver)
				drivers.DELETE("/:id", handlers.DeleteDriver)
				drivers.GET("/:id/performance", handlers.GetDriverPerformance)
				drivers.GET("/financial-data", handlers.GetDriverFinancialData)
			}

			// Tariff management
			tariffs := admin.Group("/tariffs")
			{
				tariffs.POST("/", handlers.CreateTariff)
				tariffs.PUT("/:id", handlers.UpdateTariff)
				tariffs.PUT("/:id/active", handlers.ToggleTariffActive)
				tariffs.DELETE("/:id", handlers.DeleteTariff)
			}

			// Analytics
			admin.GET("/analytics", handlers.GetAnalytics)
			admin.GET("/analytics/revenue", handlers.GetRevenueAnalytics)
			admin.GET("/analytics/orders", handlers.GetOrderAnalytics)

			// Withdrawal management
			withdrawals := admin.Group("/withdrawals")
			{
				withdrawals.GET("/", handlers.GetWithdrawals)
				withdrawals.GET("/:id", handlers.GetWithdrawal)
				withdrawals.PUT("/:id", handlers.UpdateWithdrawal)
				withdrawals.DELETE("/:id", handlers.DeleteWithdrawal)
			}

			// Payment management (admin only)
			payments := admin.Group("/payments")
			{
				payments.GET("/", handlers.GetPayments)
				payments.GET("/:id", handlers.GetPayment)
				payments.PUT("/:id/status", handlers.UpdatePaymentStatus)
				payments.POST("/:id/process", handlers.ProcessPayment)
				payments.GET("/stats", handlers.GetPaymentStats)
			}

			// Notification management (admin only)
			notifications := admin.Group("/notifications")
			{
				notifications.POST("/", handlers.CreateNotification)
				notifications.POST("/bulk", handlers.SendBulkNotification)
				notifications.GET("/", handlers.GetNotifications)
				notifications.GET("/:id", handlers.GetNotification)
				notifications.PUT("/:id/read", handlers.MarkNotificationAsRead)
				notifications.PUT("/read-all", handlers.MarkAllNotificationsAsRead)
				notifications.DELETE("/:id", handlers.DeleteNotification)
				notifications.GET("/stats", handlers.GetNotificationStats)
			}
		}

		// Driver routes
		driver := api.Group("/driver")
		driver.Use(middleware.AuthMiddleware(), middleware.DriverMiddleware())
		{
			driver.GET("/orders", handlers.GetDriverOrders)
			driver.PUT("/orders/:id/accept", handlers.AcceptOrder)
			driver.PUT("/orders/:id/complete", handlers.CompleteOrder)
			driver.GET("/earnings", handlers.GetDriverEarnings)
			driver.POST("/withdrawals", handlers.CreateWithdrawal)
			driver.GET("/withdrawals", handlers.GetDriverWithdrawals)

			// FCM Token management
			driver.POST("/fcm-token", handlers.UpdateFCMToken)
			driver.GET("/fcm-token", handlers.GetFCMToken)
			driver.DELETE("/fcm-token", handlers.DeleteFCMToken)

			// Location tracking
			driver.POST("/location", handlers.UpdateDriverLocation)
			driver.GET("/location", handlers.GetCurrentDriverLocation)
			driver.PUT("/online-status", handlers.SetDriverOnlineStatus)
			driver.GET("/location/history", handlers.GetLocationHistory)
		}
	}
}
