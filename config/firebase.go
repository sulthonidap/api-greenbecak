package config

import (
	"log"
	"os"

	"greenbecak-backend/services"
)

// Firebase configuration
var FirebaseService *services.FirebaseService

// Initialize Firebase service
func InitFirebase() {
	projectID := os.Getenv("FIREBASE_PROJECT_ID")

	// Only initialize Firebase if project ID is provided
	if projectID == "" {
		log.Println("Firebase project ID not provided, skipping Firebase initialization")
		return
	}

	// Try Service Account first, then fallback to legacy server key
	if serviceAccountPath := os.Getenv("FIREBASE_SERVICE_ACCOUNT_PATH"); serviceAccountPath != "" {
		FirebaseService = services.NewFirebaseServiceWithServiceAccount(serviceAccountPath, projectID)
		log.Println("Firebase initialized with service account")
	} else if serverKey := os.Getenv("FIREBASE_SERVER_KEY"); serverKey != "" {
		FirebaseService = services.NewFirebaseService(serverKey, projectID)
		log.Println("Firebase initialized with server key")
	} else {
		log.Println("Firebase credentials not provided, skipping Firebase initialization")
	}
}
