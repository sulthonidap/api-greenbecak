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
	"greenbecak-backend/utils"
)

func setupTestDB() {
	// Setup test database connection
	// In a real test, you would use a test database
}

func TestLogin(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create test user
	hashedPassword, _ := utils.HashPassword("password123")
	user := models.User{
		Username: "testuser",
		Email:    "test@example.com",
		Password: hashedPassword,
		Name:     "Test User",
		Role:     models.RoleCustomer,
		IsActive: true,
	}

	// Mock database call (in real test, you'd use a test database)
	// db.Create(&user)

	// Test request
	loginReq := LoginRequest{
		Username: "testuser",
		Password: "password123",
	}
	jsonData, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	// Call function
	Login(c)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "token")
	assert.Contains(t, response, "user")
}

func TestRegister(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Test request
	registerReq := RegisterRequest{
		Username: "newuser",
		Email:    "newuser@example.com",
		Password: "password123",
		Name:     "New User",
		Phone:    "08123456789",
		Address:  "Test Address",
	}
	jsonData, _ := json.Marshal(registerReq)
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	// Call function
	Register(c)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "token")
	assert.Contains(t, response, "user")
}

func TestGetProfile(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Set user context (simulating authenticated user)
	c.Set("user_id", uint(1))
	c.Set("username", "testuser")
	c.Set("role", "customer")

	req := httptest.NewRequest("GET", "/api/profile", nil)
	c.Request = req

	// Call function
	GetProfile(c)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "user")
}
