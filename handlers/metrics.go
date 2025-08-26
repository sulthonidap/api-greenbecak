package handlers

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type Metrics struct {
	TotalRequests    int64            `json:"total_requests"`
	RequestsPerMinute map[string]int64 `json:"requests_per_minute"`
	ErrorCount       int64            `json:"error_count"`
	AverageResponseTime float64        `json:"average_response_time"`
	Uptime           string           `json:"uptime"`
	LastUpdated      time.Time        `json:"last_updated"`
}

var (
	metrics     Metrics
	metricsLock sync.RWMutex
	metricsStartTime   = time.Now()
)

func UpdateMetrics(method, path string, statusCode int, responseTime time.Duration) {
	metricsLock.Lock()
	defer metricsLock.Unlock()

	metrics.TotalRequests++
	
	// Track requests per minute
	minuteKey := time.Now().Format("2006-01-02 15:04")
	if metrics.RequestsPerMinute == nil {
		metrics.RequestsPerMinute = make(map[string]int64)
	}
	metrics.RequestsPerMinute[minuteKey]++

	// Track errors
	if statusCode >= 400 {
		metrics.ErrorCount++
	}

	// Update average response time
	if metrics.AverageResponseTime == 0 {
		metrics.AverageResponseTime = float64(responseTime.Milliseconds())
	} else {
		metrics.AverageResponseTime = (metrics.AverageResponseTime + float64(responseTime.Milliseconds())) / 2
	}

	metrics.Uptime = time.Since(metricsStartTime).String()
	metrics.LastUpdated = time.Now()
}

func GetMetrics(c *gin.Context) {
	metricsLock.RLock()
	defer metricsLock.RUnlock()

	// Clean old minute data (keep only last 60 minutes)
	now := time.Now()
	cleanMetrics := make(map[string]int64)
	for minute, count := range metrics.RequestsPerMinute {
		minuteTime, err := time.Parse("2006-01-02 15:04", minute)
		if err == nil && now.Sub(minuteTime) <= 60*time.Minute {
			cleanMetrics[minute] = count
		}
	}
	metrics.RequestsPerMinute = cleanMetrics

	c.JSON(http.StatusOK, gin.H{
		"metrics": metrics,
	})
}

func ResetMetrics(c *gin.Context) {
	metricsLock.Lock()
	defer metricsLock.Unlock()

	metrics = Metrics{
		RequestsPerMinute: make(map[string]int64),
		LastUpdated:       time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Metrics reset successfully",
	})
}
