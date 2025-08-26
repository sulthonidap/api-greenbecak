package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"greenbecak-backend/database"
	"greenbecak-backend/models"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	
	// Setup test database
	db := database.GetDB()
	
	// Setup routes
	SetupRoutes(r, db)
	
	return r
}

func TestCreatePayment(t *testing.T) {
	router := setupTestRouter()
	
	// Test data
	paymentData := map[string]interface{}{
		"order_id": 1,
		"method":   "cash",
		"amount":   25000,
		"notes":    "Test payment",
	}
	
	jsonData, _ := json.Marshal(paymentData)
	req, _ := http.NewRequest("POST", "/api/payments", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "message")
	assert.Contains(t, response, "payment")
}

func TestGetPayments(t *testing.T) {
	router := setupTestRouter()
	
	req, _ := http.NewRequest("GET", "/api/payments", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "payments")
	assert.Contains(t, response, "pagination")
}

func TestCreateNotification(t *testing.T) {
	router := setupTestRouter()
	
	// Test data
	notificationData := map[string]interface{}{
		"user_id":  1,
		"title":    "Test Notification",
		"message":  "This is a test notification",
		"type":     "system",
		"priority": "normal",
		"data": map[string]interface{}{
			"test_key": "test_value",
		},
	}
	
	jsonData, _ := json.Marshal(notificationData)
	req, _ := http.NewRequest("POST", "/api/admin/notifications", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer admin-token")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "message")
	assert.Contains(t, response, "notification")
}

func TestGetNotifications(t *testing.T) {
	router := setupTestRouter()
	
	req, _ := http.NewRequest("GET", "/api/notifications", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "notifications")
	assert.Contains(t, response, "pagination")
}

func TestUpdateDriverLocation(t *testing.T) {
	router := setupTestRouter()
	
	// Test data
	locationData := map[string]interface{}{
		"latitude":  -7.797068,
		"longitude": 110.370529,
		"accuracy":  10.5,
		"speed":     25.0,
		"heading":   90.0,
	}
	
	jsonData, _ := json.Marshal(locationData)
	req, _ := http.NewRequest("POST", "/api/driver/location", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer driver-token")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "message")
	assert.Contains(t, response, "location")
}

func TestGetNearbyDrivers(t *testing.T) {
	router := setupTestRouter()
	
	req, _ := http.NewRequest("GET", "/api/location/drivers/nearby?lat=-7.797068&lng=110.370529&radius=5", nil)
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "drivers")
	assert.Contains(t, response, "count")
	assert.Contains(t, response, "radius")
}

func TestGetDriverRoute(t *testing.T) {
	router := setupTestRouter()
	
	req, _ := http.NewRequest("GET", "/api/location/routes/1", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "order_id")
	assert.Contains(t, response, "driver_location")
	assert.Contains(t, response, "route")
}

func TestPaymentValidation(t *testing.T) {
	router := setupTestRouter()
	
	// Test invalid payment method
	invalidPaymentData := map[string]interface{}{
		"order_id": 1,
		"method":   "invalid_method",
		"amount":   25000,
	}
	
	jsonData, _ := json.Marshal(invalidPaymentData)
	req, _ := http.NewRequest("POST", "/api/payments", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
}

func TestNotificationValidation(t *testing.T) {
	router := setupTestRouter()
	
	// Test invalid notification type
	invalidNotificationData := map[string]interface{}{
		"user_id":  1,
		"title":    "Test",
		"message":  "Test message",
		"type":     "invalid_type",
		"priority": "normal",
	}
	
	jsonData, _ := json.Marshal(invalidNotificationData)
	req, _ := http.NewRequest("POST", "/api/admin/notifications", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer admin-token")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
}

func TestLocationValidation(t *testing.T) {
	router := setupTestRouter()
	
	// Test invalid coordinates
	invalidLocationData := map[string]interface{}{
		"latitude":  200.0, // Invalid latitude
		"longitude": 110.370529,
		"accuracy":  10.5,
	}
	
	jsonData, _ := json.Marshal(invalidLocationData)
	req, _ := http.NewRequest("POST", "/api/driver/location", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer driver-token")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "error")
}

func TestBulkNotification(t *testing.T) {
	router := setupTestRouter()
	
	// Test data
	bulkNotificationData := map[string]interface{}{
		"user_ids":  []int{1, 2, 3},
		"title":     "Bulk Test",
		"message":   "This is a bulk notification test",
		"type":      "system",
		"priority":  "normal",
	}
	
	jsonData, _ := json.Marshal(bulkNotificationData)
	req, _ := http.NewRequest("POST", "/api/admin/notifications/bulk", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer admin-token")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "message")
	assert.Contains(t, response, "notifications_sent")
	assert.Contains(t, response, "total_users")
}

func TestPaymentStats(t *testing.T) {
	router := setupTestRouter()
	
	req, _ := http.NewRequest("GET", "/api/payments/stats", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "stats")
}

func TestNotificationStats(t *testing.T) {
	router := setupTestRouter()
	
	req, _ := http.NewRequest("GET", "/api/notifications/stats", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "stats")
}
