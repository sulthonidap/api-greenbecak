package main

import (
	"log"
	"net/http"
	"os"

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

	// Initialize database
	db := database.InitDB()
	defer database.CloseDB()

	// Initialize Firebase service
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
	corsConfig.AllowOrigins = []string{"http://localhost:3000", "http://localhost:5173"}
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	corsConfig.AllowCredentials = true
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

	// Start monitoring schedulers
	monitoring.StartAllSchedulers()

	// Start server
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	utils.GracefulShutdown(srv)
}
