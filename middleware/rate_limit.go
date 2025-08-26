package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		
		rl.mutex.Lock()
		defer rl.mutex.Unlock()
		
		now := time.Now()
		windowStart := now.Add(-rl.window)
		
		// Clean old requests
		if requests, exists := rl.requests[clientIP]; exists {
			var validRequests []time.Time
			for _, reqTime := range requests {
				if reqTime.After(windowStart) {
					validRequests = append(validRequests, reqTime)
				}
			}
			rl.requests[clientIP] = validRequests
		}
		
		// Check if limit exceeded
		if len(rl.requests[clientIP]) >= rl.limit {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}
		
		// Add current request
		rl.requests[clientIP] = append(rl.requests[clientIP], now)
		
		c.Next()
	}
}

// Default rate limiter: 100 requests per minute
func DefaultRateLimit() gin.HandlerFunc {
	limiter := NewRateLimiter(100, time.Minute)
	return limiter.RateLimit()
}

// Strict rate limiter: 10 requests per minute (for auth endpoints)
func StrictRateLimit() gin.HandlerFunc {
	limiter := NewRateLimiter(10, time.Minute)
	return limiter.RateLimit()
}
