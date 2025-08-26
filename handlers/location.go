package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"greenbecak-backend/database"
	"greenbecak-backend/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type LocationUpdate struct {
	Latitude  float64   `json:"latitude" binding:"required"`
	Longitude float64   `json:"longitude" binding:"required"`
	Accuracy  float64   `json:"accuracy"`
	Speed     float64   `json:"speed"`
	Heading   float64   `json:"heading"`
	Timestamp time.Time `json:"timestamp"`
}

type DriverLocation struct {
	DriverID  uint      `json:"driver_id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Accuracy  float64   `json:"accuracy"`
	Speed     float64   `json:"speed"`
	Heading   float64   `json:"heading"`
	IsOnline  bool      `json:"is_online"`
	LastSeen  time.Time `json:"last_seen"`
}

// UpdateDriverLocation - Update lokasi driver
func UpdateDriverLocation(c *gin.Context) {
	var req LocationUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	driverID, _ := c.Get("user_id")

	// Validate coordinates
	if req.Latitude < -90 || req.Latitude > 90 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid latitude"})
		return
	}
	if req.Longitude < -180 || req.Longitude > 180 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid longitude"})
		return
	}

	// Set timestamp if not provided
	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now()
	}

	// Update or create location record
	var location models.DriverLocation
	result := db.Where("driver_id = ?", driverID).First(&location)

	if result.Error == gorm.ErrRecordNotFound {
		// Create new location record
		location = models.DriverLocation{
			DriverID:  driverID.(uint),
			Latitude:  req.Latitude,
			Longitude: req.Longitude,
			Accuracy:  req.Accuracy,
			Speed:     req.Speed,
			Heading:   req.Heading,
			IsOnline:  true,
			LastSeen:  req.Timestamp,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		db.Create(&location)
	} else {
		// Update existing location
		location.Latitude = req.Latitude
		location.Longitude = req.Longitude
		location.Accuracy = req.Accuracy
		location.Speed = req.Speed
		location.Heading = req.Heading
		location.IsOnline = true
		location.LastSeen = req.Timestamp
		location.UpdatedAt = time.Now()
		db.Save(&location)
	}

	// Broadcast location update to connected clients (simulasi)
	go broadcastLocationUpdate(location)

	c.JSON(http.StatusOK, gin.H{
		"message":  "Location updated successfully",
		"location": location,
	})
}

// GetDriverLocation - Mendapatkan lokasi driver berdasarkan ID (public)
func GetDriverLocation(c *gin.Context) {
	db := database.GetDB()
	driverID := c.Param("id")

	var location models.DriverLocation
	if err := db.Where("driver_id = ?", driverID).First(&location).Error; err != nil {
		// Return default offline status if location not found
		c.JSON(http.StatusOK, gin.H{
			"location": gin.H{
				"driver_id": driverID,
				"latitude":  0.0,
				"longitude": 0.0,
				"is_online": false,
				"last_seen": time.Now(),
			},
		})
		return
	}

	// Check if driver is online (within last 5 minutes)
	isOnline := time.Since(location.LastSeen) < 5*time.Minute
	location.IsOnline = isOnline

	c.JSON(http.StatusOK, gin.H{"location": location})
}

// GetCurrentDriverLocation - Mendapatkan lokasi driver yang sedang login
func GetCurrentDriverLocation(c *gin.Context) {
	db := database.GetDB()
	userID, _ := c.Get("user_id")

	// Check if driver exists
	var driver models.Driver
	if err := db.Where("user_id = ?", userID).First(&driver).Error; err != nil {
		// Return default offline status if driver not found
		c.JSON(http.StatusOK, gin.H{
			"is_online": false,
			"latitude":  0.0,
			"longitude": 0.0,
			"last_seen": time.Now(),
		})
		return
	}

	var location models.DriverLocation
	if err := db.Where("driver_id = ?", driver.ID).First(&location).Error; err != nil {
		// Return default offline status if location not found
		c.JSON(http.StatusOK, gin.H{
			"is_online": false,
			"latitude":  0.0,
			"longitude": 0.0,
			"last_seen": time.Now(),
		})
		return
	}

	// Check if driver is online (within last 5 minutes)
	isOnline := time.Since(location.LastSeen) < 5*time.Minute
	location.IsOnline = isOnline

	c.JSON(http.StatusOK, gin.H{
		"is_online": location.IsOnline,
		"latitude":  location.Latitude,
		"longitude": location.Longitude,
		"accuracy":  location.Accuracy,
		"speed":     location.Speed,
		"heading":   location.Heading,
		"last_seen": location.LastSeen,
	})
}

// GetNearbyDrivers - Mendapatkan driver yang berada di sekitar lokasi
func GetNearbyDrivers(c *gin.Context) {
	db := database.GetDB()

	// Get query parameters
	lat, _ := strconv.ParseFloat(c.Query("lat"), 64)
	lng, _ := strconv.ParseFloat(c.Query("lng"), 64)
	radius, _ := strconv.ParseFloat(c.DefaultQuery("radius", "5"), 64) // Default 5km
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if lat == 0 || lng == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Latitude and longitude are required"})
		return
	}

	// Calculate bounding box for efficient querying
	// 1 degree latitude ≈ 111km, 1 degree longitude ≈ 111km * cos(latitude)
	latDelta := radius / 111.0
	lngDelta := radius / (111.0 * cos(lat*3.14159/180))

	var drivers []models.DriverLocation
	query := db.Where("is_online = ? AND last_seen > ?", true, time.Now().Add(-5*time.Minute)).
		Where("latitude BETWEEN ? AND ?", lat-latDelta, lat+latDelta).
		Where("longitude BETWEEN ? AND ?", lng-lngDelta, lng+lngDelta)

	if err := query.Limit(limit).Find(&drivers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch nearby drivers"})
		return
	}

	// Calculate actual distances and filter
	var nearbyDrivers []map[string]interface{}
	for _, driver := range drivers {
		distance := calculateDistance(lat, lng, driver.Latitude, driver.Longitude)
		if distance <= radius {
			driverData := map[string]interface{}{
				"driver_id": driver.DriverID,
				"latitude":  driver.Latitude,
				"longitude": driver.Longitude,
				"distance":  distance,
				"is_online": driver.IsOnline,
				"last_seen": driver.LastSeen,
			}
			nearbyDrivers = append(nearbyDrivers, driverData)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"drivers": nearbyDrivers,
		"count":   len(nearbyDrivers),
		"radius":  radius,
	})
}

// GetDriverRoute - Mendapatkan rute driver untuk order tertentu
func GetDriverRoute(c *gin.Context) {
	db := database.GetDB()
	orderID := c.Param("order_id")

	var order models.Order
	if err := db.Preload("Driver").First(&order, orderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	if order.DriverID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order has no assigned driver"})
		return
	}

	// Get driver location
	var driverLocation models.DriverLocation
	if err := db.Where("driver_id = ?", *order.DriverID).First(&driverLocation).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Driver location not found"})
		return
	}

	// Simulate route calculation (in real app, use Google Maps API or similar)
	route := calculateRoute(driverLocation.Latitude, driverLocation.Longitude, order.PickupLocation, order.DropLocation)

	c.JSON(http.StatusOK, gin.H{
		"order_id": orderID,
		"driver_location": gin.H{
			"latitude":  driverLocation.Latitude,
			"longitude": driverLocation.Longitude,
		},
		"pickup_location": order.PickupLocation,
		"drop_location":   order.DropLocation,
		"route":           route,
		"estimated_time":  route.EstimatedTime,
		"distance":        route.Distance,
	})
}

// SetDriverOnlineStatus - Set status online/offline driver
func SetDriverOnlineStatus(c *gin.Context) {
	var req struct {
		IsOnline bool `json:"is_online"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("Error binding JSON: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	driverID, _ := c.Get("user_id")

	fmt.Printf("Setting online status for driver ID: %v, is_online: %v\n", driverID, req.IsOnline)

	// Check if driver exists
	var driver models.Driver
	if err := db.Where("user_id = ?", driverID).First(&driver).Error; err != nil {
		fmt.Printf("Driver not found for user ID: %v, error: %v\n", driverID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Driver not found for this user"})
		return
	}

	fmt.Printf("Found driver: ID=%v, UserID=%v, Code=%v\n", driver.ID, driver.UserID, driver.DriverCode)

	var location models.DriverLocation
	result := db.Where("driver_id = ?", driver.ID).First(&location)

	if result.Error == gorm.ErrRecordNotFound {
		fmt.Printf("Driver location not found, creating new record for driver ID: %v\n", driver.ID)
		// Create new location record if not exists
		location = models.DriverLocation{
			DriverID:  driver.ID,
			Latitude:  0.0, // Default coordinates (will be updated when location is set)
			Longitude: 0.0,
			Accuracy:  0.0,
			Speed:     0.0,
			Heading:   0.0,
			IsOnline:  req.IsOnline,
			LastSeen:  time.Now(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := db.Create(&location).Error; err != nil {
			fmt.Printf("Error creating driver location: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create driver location: " + err.Error()})
			return
		}
		fmt.Printf("Successfully created driver location record\n")
	} else if result.Error != nil {
		fmt.Printf("Error finding driver location: %v\n", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find driver location: " + result.Error.Error()})
		return
	} else {
		fmt.Printf("Found existing driver location, updating status\n")
		// Update existing location
		location.IsOnline = req.IsOnline
		location.LastSeen = time.Now()
		location.UpdatedAt = time.Now()

		if err := db.Save(&location).Error; err != nil {
			fmt.Printf("Error updating driver location: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update online status: " + err.Error()})
			return
		}
		fmt.Printf("Successfully updated driver location\n")
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   fmt.Sprintf("Driver status updated to %s", map[bool]string{true: "online", false: "offline"}[req.IsOnline]),
		"is_online": req.IsOnline,
	})
}

// GetLocationHistory - Mendapatkan history lokasi driver
func GetLocationHistory(c *gin.Context) {
	db := database.GetDB()
	driverID := c.Param("id")

	// Get query parameters
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))

	query := db.Where("driver_id = ?", driverID)

	if startDate != "" {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("created_at <= ?", endDate)
	}

	var locations []models.DriverLocation
	if err := query.Order("created_at DESC").Limit(limit).Find(&locations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch location history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"driver_id": driverID,
		"locations": locations,
		"count":     len(locations),
	})
}

// Helper functions
func cos(x float64) float64 {
	// Simple cosine approximation
	return 1 - x*x/2 + x*x*x*x/24
}

func calculateDistance(lat1, lng1, lat2, lng2 float64) float64 {
	// Haversine formula for calculating distance between two points
	const R = 6371 // Earth's radius in kilometers

	lat1Rad := lat1 * 3.14159 / 180
	lat2Rad := lat2 * 3.14159 / 180
	deltaLat := (lat2 - lat1) * 3.14159 / 180
	deltaLng := (lng2 - lng1) * 3.14159 / 180

	a := sin(deltaLat/2)*sin(deltaLat/2) + cos(lat1Rad)*cos(lat2Rad)*sin(deltaLng/2)*sin(deltaLng/2)
	c := 2 * atan2(sqrt(a), sqrt(1-a))

	return R * c
}

func sin(x float64) float64 {
	// Simple sine approximation
	return x - x*x*x/6 + x*x*x*x*x/120
}

func atan2(y, x float64) float64 {
	// Simple atan2 approximation
	if x > 0 {
		return atan(y / x)
	} else if x < 0 && y >= 0 {
		return atan(y/x) + 3.14159
	} else if x < 0 && y < 0 {
		return atan(y/x) - 3.14159
	} else if x == 0 && y > 0 {
		return 3.14159 / 2
	} else if x == 0 && y < 0 {
		return -3.14159 / 2
	} else {
		return 0
	}
}

func atan(x float64) float64 {
	// Simple arctangent approximation
	return x - x*x*x/3 + x*x*x*x*x/5
}

func sqrt(x float64) float64 {
	// Simple square root approximation
	if x < 0 {
		return 0
	}
	if x == 0 {
		return 0
	}

	z := x / 2
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}

type Route struct {
	Waypoints     []map[string]float64 `json:"waypoints"`
	Distance      float64              `json:"distance"`
	EstimatedTime int                  `json:"estimated_time"`
}

func calculateRoute(startLat, startLng float64, pickupLocation, dropLocation string) Route {
	// Simulate route calculation
	// In real implementation, use Google Maps Directions API or similar

	// Mock waypoints
	waypoints := []map[string]float64{
		{"lat": startLat, "lng": startLng},
		{"lat": startLat + 0.001, "lng": startLng + 0.001}, // Pickup location
		{"lat": startLat + 0.002, "lng": startLng + 0.002}, // Drop location
	}

	return Route{
		Waypoints:     waypoints,
		Distance:      2.5, // km
		EstimatedTime: 15,  // minutes
	}
}

func broadcastLocationUpdate(location models.DriverLocation) {
	// Simulate broadcasting location update to connected clients
	// In real implementation, use WebSocket or Server-Sent Events
	fmt.Printf("Broadcasting location update for driver %d: %.6f, %.6f\n",
		location.DriverID, location.Latitude, location.Longitude)
}
