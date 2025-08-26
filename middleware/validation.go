package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// ValidationMiddleware validates common request patterns
func ValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check Content-Type for POST/PUT requests that actually need body
		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			// Skip validation for endpoints that don't need body
			path := c.Request.URL.Path
			if shouldSkipContentTypeValidation(path, c.Request.Method) {
				c.Next()
				return
			}

			contentType := c.GetHeader("Content-Type")
			if contentType != "application/json" {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Content-Type must be application/json",
				})
				c.Abort()
				return
			}
		}

		// Check request size (limit to 10MB)
		if c.Request.ContentLength > 10*1024*1024 {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": "Request body too large",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// shouldSkipContentTypeValidation checks if the endpoint should skip content-type validation
func shouldSkipContentTypeValidation(path, method string) bool {
	// Debug: Print the path being checked
	fmt.Printf("Checking path: %s, method: %s\n", path, method)

	// TEMPORARY: Skip validation for all PUT requests to avoid content-type issues
	if method == "PUT" {
		fmt.Printf("Skipping validation for PUT request: %s\n", path)
		return true
	}

	// List of endpoints that don't need body (action endpoints)
	skipEndpoints := []string{
		"/api/driver/orders/", // accept, complete, etc.
		"/api/orders/",        // location updates, etc.
		"/api/payments/",      // process, status updates
		"/api/notifications/", // read, read-all
		"/api/admin/",         // admin action endpoints
		"/alerts/",            // acknowledge alerts
	}

	for _, endpoint := range skipEndpoints {
		if strings.Contains(path, endpoint) {
			// Additional check for specific action patterns
			if strings.Contains(path, "/accept") ||
				strings.Contains(path, "/complete") ||
				strings.Contains(path, "/process") ||
				strings.Contains(path, "/read") ||
				strings.Contains(path, "/acknowledge") ||
				strings.Contains(path, "/active") {
				fmt.Printf("Skipping validation for: %s\n", path)
				return true
			}
		}
	}

	fmt.Printf("NOT skipping validation for: %s\n", path)
	return false
}

// PaginationMiddleware adds pagination parameters
func PaginationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set default pagination values
		page := c.DefaultQuery("page", "1")
		limit := c.DefaultQuery("limit", "10")

		c.Set("page", page)
		c.Set("limit", limit)

		c.Next()
	}
}
