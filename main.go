package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"greenbecak-backend/config"
	"greenbecak-backend/database"
	"greenbecak-backend/middleware"
	"greenbecak-backend/monitoring"
	"greenbecak-backend/routes"
	"greenbecak-backend/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Validate environment variables
	if err := utils.ValidateEnvironment(); err != nil {
		log.Fatal("Environment validation failed:", err)
	}

	// Initialize database (non-blocking)
	db := database.InitDB()
	defer database.CloseDB()

	// Initialize Firebase service (non-blocking)
	config.InitFirebase()

	// Set Gin mode
	gin.SetMode(os.Getenv("SERVER_MODE"))

	// Initialize router with custom recovery
	r := gin.New()
	r.Use(middleware.RecoveryMiddleware())

	// Load HTML templates
	r.LoadHTMLGlob("templates/*")

	// CORS configuration
	corsConfig := cors.DefaultConfig()

	// Get allowed origins from environment variable
	corsOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	if corsOrigins == "" {
		// Default to localhost if not set
		corsConfig.AllowOrigins = []string{"http://localhost:3000", "http://localhost:5173"}
	} else {
		// Split comma-separated origins
		origins := strings.Split(corsOrigins, ",")
		for i, origin := range origins {
			origins[i] = strings.TrimSpace(origin)
		}
		corsConfig.AllowOrigins = origins
	}

	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"}
	corsConfig.AllowCredentials = true
	corsConfig.ExposeHeaders = []string{"Content-Length", "Content-Type"}

	// Log CORS configuration for debugging
	log.Printf("CORS Allowed Origins: %v", corsConfig.AllowOrigins)

	r.Use(cors.New(corsConfig))

	// Add logging middleware
	r.Use(middleware.LoggingMiddleware())
	r.Use(middleware.ErrorLoggingMiddleware())
	r.Use(middleware.ValidationMiddleware())
	r.Use(middleware.MetricsMiddleware())

	// Initialize routes
	routes.SetupRoutes(r, db)

	// Setup Swagger documentation
	routes.SetupSwagger(r)

	// Start monitoring schedulers with delay to ensure database is ready
	go func() {
		time.Sleep(10 * time.Second) // Wait for database connection to be ready
		monitoring.StartAllSchedulers()
	}()

	// Start server
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    "0.0.0.0:" + port,
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()

	// Give server time to start
	time.Sleep(3 * time.Second)
	log.Println("Server is ready to accept connections")

	// Test health endpoint
	go func() {
		time.Sleep(5 * time.Second)
		log.Println("Testing health endpoint...")
		resp, err := http.Get("http://localhost:" + port + "/health")
		if err != nil {
			log.Printf("Health check failed: %v", err)
		} else {
			log.Printf("Health check status: %d", resp.StatusCode)
			resp.Body.Close()
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	utils.GracefulShutdown(srv)
}
