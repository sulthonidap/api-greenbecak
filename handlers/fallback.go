package handlers

import (
	"net/http"
	"time"

	"greenbecak-backend/utils"

	"github.com/gin-gonic/gin"
)

// FallbackResponse represents a fallback response when database is not available
type FallbackResponse struct {
	Success   bool                 `json:"success"`
	Message   string               `json:"message"`
	Data      interface{}          `json:"data,omitempty"`
	Timestamp time.Time            `json:"timestamp"`
	Database  utils.DatabaseStatus `json:"database"`
}

// CheckDatabaseAndRespond checks if database is available and responds accordingly
func CheckDatabaseAndRespond(c *gin.Context, successHandler func(*gin.Context)) {
	if !utils.IsDatabaseConnected() {
		// Database not available, return fallback response
		fallback := FallbackResponse{
			Success:   false,
			Message:   "Service temporarily unavailable - Database connection issue",
			Data:      nil,
			Timestamp: time.Now(),
			Database:  utils.GetDatabaseStatus(),
		}
		c.JSON(http.StatusServiceUnavailable, fallback)
		return
	}

	// Database available, proceed with normal handler
	successHandler(c)
}

// GetFallbackResponse returns a standard fallback response
func GetFallbackResponse(message string, data interface{}) FallbackResponse {
	return FallbackResponse{
		Success:   false,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
		Database:  utils.GetDatabaseStatus(),
	}
}

// RespondWithFallback sends a fallback response
func RespondWithFallback(c *gin.Context, statusCode int, message string, data interface{}) {
	fallback := GetFallbackResponse(message, data)
	c.JSON(statusCode, fallback)
}

// RespondWithDatabaseError sends a database error response
func RespondWithDatabaseError(c *gin.Context) {
	RespondWithFallback(c, http.StatusServiceUnavailable,
		"Database temporarily unavailable. Please try again later.", nil)
}
