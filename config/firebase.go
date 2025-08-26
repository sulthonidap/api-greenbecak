package config

import (
	"greenbecak-backend/services"
	"os"
)

// Firebase configuration
var FirebaseService *services.FirebaseService

// Initialize Firebase service
func InitFirebase() {
	projectID := os.Getenv("FIREBASE_PROJECT_ID")

	// Try Service Account first, then fallback to legacy server key
	if serviceAccountPath := os.Getenv("FIREBASE_SERVICE_ACCOUNT_PATH"); serviceAccountPath != "" && projectID != "" {
		FirebaseService = services.NewFirebaseServiceWithServiceAccount(serviceAccountPath, projectID)
	} else if serverKey := os.Getenv("FIREBASE_SERVER_KEY"); serverKey != "" && projectID != "" {
		FirebaseService = services.NewFirebaseService(serverKey, projectID)
	}
}
