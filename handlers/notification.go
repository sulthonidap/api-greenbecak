package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"greenbecak-backend/database"
	"greenbecak-backend/models"
)

type NotificationRequest struct {
	UserID   uint   `json:"user_id" binding:"required"`
	Title    string `json:"title" binding:"required"`
	Message  string `json:"message" binding:"required"`
	Type     string `json:"type" binding:"required"`
	Priority string `json:"priority"`
	Data     map[string]interface{} `json:"data"`
}

type NotificationResponse struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Type      string    `json:"type"`
	Priority  string    `json:"priority"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// CreateNotification - Membuat notifikasi baru
func CreateNotification(c *gin.Context) {
	var req NotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()

	// Validate notification type
	validTypes := []string{"order", "payment", "system", "promo", "driver"}
	isValidType := false
	for _, t := range validTypes {
		if req.Type == t {
			isValidType = true
			break
		}
	}

	if !isValidType {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification type"})
		return
	}

	// Validate priority
	if req.Priority == "" {
		req.Priority = "normal"
	}

	validPriorities := []string{"low", "normal", "high", "urgent"}
	isValidPriority := false
	for _, p := range validPriorities {
		if req.Priority == p {
			isValidPriority = true
			break
		}
	}

	if !isValidPriority {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid priority level"})
		return
	}

	// Check if user exists
	var user models.User
	if err := db.First(&user, req.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Convert data to JSON string
	dataJSON := "{}"
	if req.Data != nil {
		if dataBytes, err := json.Marshal(req.Data); err == nil {
			dataJSON = string(dataBytes)
		}
	}

	notification := models.Notification{
		UserID:    req.UserID,
		Title:     req.Title,
		Message:   req.Message,
		Type:      models.NotificationType(req.Type),
		Priority:  models.NotificationPriority(req.Priority),
		IsRead:    false,
		Data:      dataJSON,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.Create(&notification).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}

	// Send real-time notification (simulasi)
	go sendRealTimeNotification(notification)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Notification created successfully",
		"notification": notification,
	})
}

// GetNotifications - Mendapatkan daftar notifikasi user
func GetNotifications(c *gin.Context) {
	db := database.GetDB()
	userID, _ := c.Get("user_id")

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset := (page - 1) * limit

	// Filter options
	notificationType := c.Query("type")
	isRead := c.Query("read")
	priority := c.Query("priority")

	var notifications []models.Notification
	query := db.Where("user_id = ?", userID)

	// Apply filters
	if notificationType != "" {
		query = query.Where("type = ?", notificationType)
	}
	if isRead != "" {
		if isRead == "true" {
			query = query.Where("is_read = ?", true)
		} else if isRead == "false" {
			query = query.Where("is_read = ?", false)
		}
	}
	if priority != "" {
		query = query.Where("priority = ?", priority)
	}

	var total int64
	query.Model(&models.Notification{}).Count(&total)

	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&notifications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetNotification - Mendapatkan detail notifikasi
func GetNotification(c *gin.Context) {
	db := database.GetDB()
	notificationID := c.Param("id")
	userID, _ := c.Get("user_id")

	var notification models.Notification
	if err := db.Where("id = ? AND user_id = ?", notificationID, userID).First(&notification).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	// Mark as read
	if !notification.IsRead {
		notification.IsRead = true
		notification.UpdatedAt = time.Now()
		db.Save(&notification)
	}

	c.JSON(http.StatusOK, gin.H{"notification": notification})
}

// MarkNotificationAsRead - Menandai notifikasi sebagai sudah dibaca
func MarkNotificationAsRead(c *gin.Context) {
	db := database.GetDB()
	notificationID := c.Param("id")
	userID, _ := c.Get("user_id")

	var notification models.Notification
	if err := db.Where("id = ? AND user_id = ?", notificationID, userID).First(&notification).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	notification.IsRead = true
	notification.UpdatedAt = time.Now()

	if err := db.Save(&notification).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Notification marked as read",
		"notification": notification,
	})
}

// MarkAllNotificationsAsRead - Menandai semua notifikasi sebagai sudah dibaca
func MarkAllNotificationsAsRead(c *gin.Context) {
	db := database.GetDB()
	userID, _ := c.Get("user_id")

	if err := db.Model(&models.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Updates(map[string]interface{}{
			"is_read":     true,
			"updated_at":  time.Now(),
		}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All notifications marked as read"})
}

// DeleteNotification - Menghapus notifikasi
func DeleteNotification(c *gin.Context) {
	db := database.GetDB()
	notificationID := c.Param("id")
	userID, _ := c.Get("user_id")

	var notification models.Notification
	if err := db.Where("id = ? AND user_id = ?", notificationID, userID).First(&notification).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	if err := db.Delete(&notification).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification deleted successfully"})
}

// GetNotificationStats - Mendapatkan statistik notifikasi
func GetNotificationStats(c *gin.Context) {
	db := database.GetDB()
	userID, _ := c.Get("user_id")

	var stats struct {
		TotalNotifications int64 `json:"total_notifications"`
		UnreadCount       int64 `json:"unread_count"`
		ReadCount         int64 `json:"read_count"`
		ByType            map[string]int64 `json:"by_type"`
		ByPriority        map[string]int64 `json:"by_priority"`
	}

	// Total notifications
	db.Model(&models.Notification{}).Where("user_id = ?", userID).Count(&stats.TotalNotifications)

	// Unread count
	db.Model(&models.Notification{}).Where("user_id = ? AND is_read = ?", userID, false).Count(&stats.UnreadCount)

	// Read count
	db.Model(&models.Notification{}).Where("user_id = ? AND is_read = ?", userID, true).Count(&stats.ReadCount)

	// Count by type
	stats.ByType = make(map[string]int64)
	types := []string{"order", "payment", "system", "promo", "driver"}
	for _, t := range types {
		var count int64
		db.Model(&models.Notification{}).Where("user_id = ? AND type = ?", userID, t).Count(&count)
		stats.ByType[t] = count
	}

	// Count by priority
	stats.ByPriority = make(map[string]int64)
	priorities := []string{"low", "normal", "high", "urgent"}
	for _, p := range priorities {
		var count int64
		db.Model(&models.Notification{}).Where("user_id = ? AND priority = ?", userID, p).Count(&count)
		stats.ByPriority[p] = count
	}

	c.JSON(http.StatusOK, gin.H{"stats": stats})
}

// SendBulkNotification - Mengirim notifikasi ke multiple users
func SendBulkNotification(c *gin.Context) {
	var req struct {
		UserIDs  []uint                 `json:"user_ids" binding:"required"`
		Title    string                 `json:"title" binding:"required"`
		Message  string                 `json:"message" binding:"required"`
		Type     string                 `json:"type" binding:"required"`
		Priority string                 `json:"priority"`
		Data     map[string]interface{} `json:"data"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()

	// Validate notification type
	validTypes := []string{"order", "payment", "system", "promo", "driver"}
	isValidType := false
	for _, t := range validTypes {
		if req.Type == t {
			isValidType = true
			break
		}
	}

	if !isValidType {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification type"})
		return
	}

	// Set default priority
	if req.Priority == "" {
		req.Priority = "normal"
	}

	// Convert data to JSON string
	dataJSON := "{}"
	if req.Data != nil {
		if dataBytes, err := json.Marshal(req.Data); err == nil {
			dataJSON = string(dataBytes)
		}
	}

	var notifications []models.Notification
	var successCount int64

	for _, userID := range req.UserIDs {
		// Check if user exists
		var user models.User
		if err := db.First(&user, userID).Error; err != nil {
			continue // Skip if user not found
		}

		notification := models.Notification{
			UserID:    userID,
			Title:     req.Title,
			Message:   req.Message,
			Type:      models.NotificationType(req.Type),
			Priority:  models.NotificationPriority(req.Priority),
			IsRead:    false,
			Data:      dataJSON,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := db.Create(&notification).Error; err == nil {
			notifications = append(notifications, notification)
			successCount++
			
			// Send real-time notification
			go sendRealTimeNotification(notification)
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": fmt.Sprintf("Bulk notification sent successfully. %d/%d notifications created", successCount, len(req.UserIDs)),
		"notifications_sent": successCount,
		"total_users": len(req.UserIDs),
		"notifications": notifications,
	})
}

// sendRealTimeNotification - Simulasi pengiriman notifikasi real-time
func sendRealTimeNotification(notification models.Notification) {
	// Simulasi delay pengiriman
	time.Sleep(100 * time.Millisecond)
	
	// Di sini bisa diintegrasikan dengan WebSocket, Firebase Cloud Messaging, atau layanan push notification lainnya
	fmt.Printf("Real-time notification sent to user %d: %s\n", notification.UserID, notification.Title)
}
