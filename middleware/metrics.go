package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"greenbecak-backend/handlers"
)

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Process request
		c.Next()
		
		// Calculate response time
		duration := time.Since(start)
		
		// Update metrics
		handlers.UpdateMetrics(
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			duration,
		)
	}
}
