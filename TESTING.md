# GreenBecak Backend Testing Guide

## Overview

Guide ini menjelaskan cara melakukan testing untuk GreenBecak Backend API.

## Testing Strategy

### 1. Unit Testing
- Test individual functions dan methods
- Mock external dependencies
- Fast execution

### 2. Integration Testing
- Test API endpoints
- Test database interactions
- Test authentication flows

### 3. End-to-End Testing
- Test complete user workflows
- Test real database
- Test production-like environment

## Running Tests

### 1. Run All Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### 2. Run Specific Tests

```bash
# Run specific package
go test ./handlers

# Run specific test function
go test -run TestLogin

# Run tests with pattern
go test -run "TestAuth.*"
```

### 3. Using Makefile

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run tests with race detection
make test-race
```

## Test Structure

### 1. Unit Tests

```go
// handlers/auth_test.go
func TestLogin(t *testing.T) {
    // Setup
    gin.SetMode(gin.TestMode)
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)

    // Test data
    loginReq := LoginRequest{
        Username: "testuser",
        Password: "password123",
    }

    // Execute
    jsonData, _ := json.Marshal(loginReq)
    req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    c.Request = req

    // Call function
    Login(c)

    // Assert
    assert.Equal(t, http.StatusOK, w.Code)
    
    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.Contains(t, response, "token")
}
```

### 2. Integration Tests

```go
// integration/auth_test.go
func TestLoginIntegration(t *testing.T) {
    // Setup test database
    db := setupTestDB()
    defer cleanupTestDB(db)

    // Create test user
    user := createTestUser(db)

    // Setup router
    r := setupTestRouter(db)

    // Test request
    loginData := map[string]interface{}{
        "username": user.Username,
        "password": "password123",
    }
    jsonData, _ := json.Marshal(loginData)

    req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()

    r.ServeHTTP(w, req)

    // Assert
    assert.Equal(t, http.StatusOK, w.Code)
    
    var response map[string]interface{}
    err := json.Unmarshal(w.Body.Bytes(), &response)
    assert.NoError(t, err)
    assert.Contains(t, response, "token")
}
```

### 3. Test Helpers

```go
// test_helpers.go
func setupTestDB() *gorm.DB {
    // Setup test database connection
    dsn := "test_user:test_password@tcp(localhost:3306)/greenbecak_test?charset=utf8mb4&parseTime=True&loc=Local"
    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal("Failed to connect to test database:", err)
    }
    
    // Run migrations
    db.AutoMigrate(&models.User{}, &models.Driver{}, &models.Order{})
    
    return db
}

func createTestUser(db *gorm.DB) *models.User {
    hashedPassword, _ := utils.HashPassword("password123")
    user := &models.User{
        Username: "testuser",
        Email:    "test@example.com",
        Password: hashedPassword,
        Name:     "Test User",
        Role:     models.RoleCustomer,
        IsActive: true,
    }
    
    db.Create(user)
    return user
}

func cleanupTestDB(db *gorm.DB) {
    // Clean up test data
    db.Exec("DELETE FROM users")
    db.Exec("DELETE FROM drivers")
    db.Exec("DELETE FROM orders")
}
```

## API Testing

### 1. Manual Testing with curl

```bash
# Health check
curl http://localhost:8080/health

# Register user
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123",
    "name": "Test User"
  }'

# Login
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "password123"
  }'

# Get orders (with token)
curl -X GET http://localhost:8080/api/orders \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

### 2. Using Postman

1. Import collection dari `postman/GreenBecak_API.postman_collection.json`
2. Set environment variables
3. Run tests

### 3. Using Insomnia

1. Import workspace dari `insomnia/GreenBecak_API.json`
2. Configure environment
3. Run requests

## Performance Testing

### 1. Load Testing with Apache Bench

```bash
# Test login endpoint
ab -n 1000 -c 10 -p login_data.json -T application/json http://localhost:8080/api/auth/login

# Test health endpoint
ab -n 10000 -c 100 http://localhost:8080/health
```

### 2. Load Testing with wrk

```bash
# Install wrk
# macOS: brew install wrk
# Ubuntu: sudo apt install wrk

# Test API endpoints
wrk -t12 -c400 -d30s http://localhost:8080/health
wrk -t12 -c400 -d30s -s login_script.lua http://localhost:8080/api/auth/login
```

### 3. Load Testing Script

```lua
-- login_script.lua
wrk.method = "POST"
wrk.headers["Content-Type"] = "application/json"
wrk.body = '{"username":"testuser","password":"password123"}'
```

## Security Testing

### 1. Authentication Tests

```go
func TestAuthentication(t *testing.T) {
    // Test invalid token
    req := httptest.NewRequest("GET", "/api/orders", nil)
    req.Header.Set("Authorization", "Bearer invalid_token")
    w := httptest.NewRecorder()
    
    r.ServeHTTP(w, req)
    assert.Equal(t, http.StatusUnauthorized, w.Code)
    
    // Test missing token
    req = httptest.NewRequest("GET", "/api/orders", nil)
    w = httptest.NewRecorder()
    
    r.ServeHTTP(w, req)
    assert.Equal(t, http.StatusUnauthorized, w.Code)
}
```

### 2. Authorization Tests

```go
func TestAuthorization(t *testing.T) {
    // Test admin-only endpoint with customer token
    customerToken := getCustomerToken()
    
    req := httptest.NewRequest("GET", "/api/admin/users", nil)
    req.Header.Set("Authorization", "Bearer "+customerToken)
    w := httptest.NewRecorder()
    
    r.ServeHTTP(w, req)
    assert.Equal(t, http.StatusForbidden, w.Code)
}
```

### 3. Input Validation Tests

```go
func TestInputValidation(t *testing.T) {
    // Test invalid email
    registerData := map[string]interface{}{
        "username": "testuser",
        "email":    "invalid-email",
        "password": "password123",
        "name":     "Test User",
    }
    
    jsonData, _ := json.Marshal(registerData)
    req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonData))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    
    r.ServeHTTP(w, req)
    assert.Equal(t, http.StatusBadRequest, w.Code)
}
```

## Database Testing

### 1. Database Connection Tests

```go
func TestDatabaseConnection(t *testing.T) {
    db := database.GetDB()
    
    // Test connection
    err := db.Raw("SELECT 1").Error
    assert.NoError(t, err)
    
    // Test migration
    err = db.AutoMigrate(&models.User{})
    assert.NoError(t, err)
}
```

### 2. Database Transaction Tests

```go
func TestDatabaseTransaction(t *testing.T) {
    db := database.GetDB()
    
    // Start transaction
    tx := db.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()
    
    // Create test data
    user := &models.User{
        Username: "testuser",
        Email:    "test@example.com",
        Password: "hashed_password",
        Name:     "Test User",
    }
    
    err := tx.Create(user).Error
    assert.NoError(t, err)
    
    // Commit transaction
    err = tx.Commit().Error
    assert.NoError(t, err)
}
```

## Mock Testing

### 1. Mock Database

```go
func TestWithMockDB(t *testing.T) {
    // Create mock database
    mockDB := newMockDB()
    
    // Test with mock
    user := &models.User{
        Username: "testuser",
        Email:    "test@example.com",
    }
    
    mockDB.On("Create", user).Return(nil)
    mockDB.On("First", mock.Anything, mock.Anything).Return(nil)
    
    // Test function
    result := createUser(mockDB, user)
    
    // Assert
    assert.NoError(t, result)
    mockDB.AssertExpectations(t)
}
```

### 2. Mock External Services

```go
func TestWithMockExternalService(t *testing.T) {
    // Mock email service
    mockEmailService := newMockEmailService()
    mockEmailService.On("SendEmail", mock.Anything).Return(nil)
    
    // Test function
    result := sendWelcomeEmail(mockEmailService, "test@example.com")
    
    // Assert
    assert.NoError(t, result)
    mockEmailService.AssertExpectations(t)
}
```

## Continuous Integration

### 1. GitHub Actions

```yaml
# .github/workflows/test.yml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      mysql:
        image: mysql:8.0
        env:
          MYSQL_ROOT_PASSWORD: password
          MYSQL_DATABASE: greenbecak_test
        ports:
          - 3306:3306
        options: --health-cmd="mysqladmin ping" --health-interval=10s --health-timeout=5s --health-retries=3
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: 1.21
    
    - name: Install dependencies
      run: go mod download
    
    - name: Run tests
      run: go test -v -coverprofile=coverage.out ./...
      env:
        DB_HOST: localhost
        DB_PORT: 3306
        DB_USER: root
        DB_PASSWORD: password
        DB_NAME: greenbecak_test
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
```

### 2. Local CI

```bash
#!/bin/bash
# scripts/ci.sh

echo "Running tests..."
go test -v -coverprofile=coverage.out ./...

echo "Running linting..."
golangci-lint run

echo "Running security scan..."
gosec ./...

echo "Building application..."
go build -o greenbecak-backend main.go

echo "CI completed successfully!"
```

## Test Data Management

### 1. Test Fixtures

```go
// fixtures/users.go
var TestUsers = []models.User{
    {
        Username: "admin",
        Email:    "admin@greenbecak.com",
        Password: "hashed_password",
        Name:     "Administrator",
        Role:     models.RoleAdmin,
    },
    {
        Username: "driver1",
        Email:    "driver1@greenbecak.com",
        Password: "hashed_password",
        Name:     "Budi Santoso",
        Role:     models.RoleCustomer,
    },
}

func LoadTestUsers(db *gorm.DB) {
    for _, user := range TestUsers {
        hashedPassword, _ := utils.HashPassword("password123")
        user.Password = hashedPassword
        db.Create(&user)
    }
}
```

### 2. Test Factories

```go
// factories/user_factory.go
func CreateUser(db *gorm.DB, overrides ...map[string]interface{}) *models.User {
    user := &models.User{
        Username: "testuser",
        Email:    "test@example.com",
        Password: "hashed_password",
        Name:     "Test User",
        Role:     models.RoleCustomer,
        IsActive: true,
    }
    
    // Apply overrides
    for _, override := range overrides {
        for key, value := range override {
            switch key {
            case "username":
                user.Username = value.(string)
            case "email":
                user.Email = value.(string)
            case "role":
                user.Role = models.UserRole(value.(string))
            }
        }
    }
    
    db.Create(user)
    return user
}
```

## Test Reports

### 1. Coverage Report

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# View in browser
open coverage.html
```

### 2. Test Results

```bash
# Run tests with detailed output
go test -v -json ./... > test-results.json

# Generate test report
go test -v -coverprofile=coverage.out -json ./... | jq '.'
```

## Best Practices

### 1. Test Organization

- Group related tests together
- Use descriptive test names
- Follow AAA pattern (Arrange, Act, Assert)
- Keep tests independent

### 2. Test Data

- Use factories for test data creation
- Clean up test data after each test
- Use unique identifiers to avoid conflicts
- Mock external dependencies

### 3. Performance

- Run tests in parallel when possible
- Use test databases for integration tests
- Mock heavy operations
- Cache test data when appropriate

### 4. Maintenance

- Update tests when code changes
- Remove obsolete tests
- Keep test code clean and readable
- Document complex test scenarios

## Troubleshooting

### 1. Common Issues

**Tests Failing Randomly**
```bash
# Run tests with race detection
go test -race ./...

# Check for shared state
go test -parallel 1 ./...
```

**Database Connection Issues**
```bash
# Check database status
sudo systemctl status mysql

# Test connection
mysql -u test_user -p test_database
```

**Memory Issues**
```bash
# Run tests with memory profiling
go test -memprofile=mem.out ./...
go tool pprof mem.out
```

### 2. Debug Tests

```bash
# Run specific test with verbose output
go test -v -run TestLogin

# Run with debugger
dlv test ./handlers -run TestLogin

# Run with trace
go test -trace=trace.out ./...
go tool trace trace.out
```

## Tools

### 1. Testing Frameworks

- **testify**: Assertions and mocking
- **httptest**: HTTP testing
- **gomock**: Mock generation
- **sqlmock**: Database mocking

### 2. Code Quality

- **golangci-lint**: Linting
- **gosec**: Security scanning
- **staticcheck**: Static analysis

### 3. Performance

- **pprof**: Profiling
- **wrk**: Load testing
- **ab**: Apache bench

## Support

Untuk bantuan testing:

- **Documentation**: https://docs.greenbecak.com/testing
- **GitHub Issues**: https://github.com/greenbecak/backend/issues
- **Email Support**: support@greenbecak.com
