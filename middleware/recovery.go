package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			log.Printf("Panic recovered: %s", err)
		} else {
			log.Printf("Panic recovered: %v", recovered)
		}
		
		// Log stack trace
		log.Printf("Stack trace: %s", debug.Stack())
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error. Please try again later.",
		})
	})
}
