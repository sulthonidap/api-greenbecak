package utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"greenbecak-backend/config"
)

// DatabaseBackup creates a backup of the database
func DatabaseBackup() error {
	cfg := config.LoadConfig()
	
	// Create backup directory if it doesn't exist
	backupDir := "backups"
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %v", err)
	}

	// Generate backup filename with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	backupFile := fmt.Sprintf("%s/greenbecak_backup_%s.sql", backupDir, timestamp)

	// Create mysqldump command
	cmd := exec.Command("mysqldump",
		"-h", cfg.DBHost,
		"-P", fmt.Sprintf("%d", cfg.DBPort),
		"-u", cfg.DBUser,
		fmt.Sprintf("-p%s", cfg.DBPassword),
		cfg.DBName,
	)

	// Create output file
	outputFile, err := os.Create(backupFile)
	if err != nil {
		return fmt.Errorf("failed to create backup file: %v", err)
	}
	defer outputFile.Close()

	// Set command output
	cmd.Stdout = outputFile
	cmd.Stderr = os.Stderr

	// Execute backup
	log.Printf("Creating database backup: %s", backupFile)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create backup: %v", err)
	}

	log.Printf("Database backup completed: %s", backupFile)
	return nil
}

// RestoreDatabase restores database from backup file
func RestoreDatabase(backupFile string) error {
	cfg := config.LoadConfig()

	// Check if backup file exists
	if _, err := os.Stat(backupFile); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found: %s", backupFile)
	}

	// Create mysql command to restore
	cmd := exec.Command("mysql",
		"-h", cfg.DBHost,
		"-P", fmt.Sprintf("%d", cfg.DBPort),
		"-u", cfg.DBUser,
		fmt.Sprintf("-p%s", cfg.DBPassword),
		cfg.DBName,
	)

	// Open backup file
	inputFile, err := os.Open(backupFile)
	if err != nil {
		return fmt.Errorf("failed to open backup file: %v", err)
	}
	defer inputFile.Close()

	// Set command input
	cmd.Stdin = inputFile
	cmd.Stderr = os.Stderr

	// Execute restore
	log.Printf("Restoring database from: %s", backupFile)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to restore database: %v", err)
	}

	log.Printf("Database restore completed from: %s", backupFile)
	return nil
}

// CleanupOldBackups removes backup files older than specified days
func CleanupOldBackups(daysToKeep int) error {
	backupDir := "backups"
	
	// Read backup directory
	files, err := os.ReadDir(backupDir)
	if err != nil {
		return fmt.Errorf("failed to read backup directory: %v", err)
	}

	cutoffTime := time.Now().AddDate(0, 0, -daysToKeep)
	deletedCount := 0

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Check if file is a backup file
		if len(file.Name()) < 20 || file.Name()[:19] != "greenbecak_backup_" {
			continue
		}

		// Parse timestamp from filename
		timestampStr := file.Name()[19 : len(file.Name())-4] // Remove prefix and .sql
		fileTime, err := time.Parse("2006-01-02_15-04-05", timestampStr)
		if err != nil {
			log.Printf("Warning: Could not parse timestamp from filename: %s", file.Name())
			continue
		}

		// Delete old files
		if fileTime.Before(cutoffTime) {
			filePath := fmt.Sprintf("%s/%s", backupDir, file.Name())
			if err := os.Remove(filePath); err != nil {
				log.Printf("Warning: Failed to delete old backup file: %s", filePath)
			} else {
				deletedCount++
				log.Printf("Deleted old backup file: %s", filePath)
			}
		}
	}

	log.Printf("Cleanup completed. Deleted %d old backup files.", deletedCount)
	return nil
}
