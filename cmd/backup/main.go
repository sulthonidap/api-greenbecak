package main

import (
	"flag"
	"log"
	"os"

	"github.com/joho/godotenv"
	"greenbecak-backend/utils"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Parse command line flags
	var (
		backupFlag   = flag.Bool("backup", false, "Create a new database backup")
		restoreFlag  = flag.String("restore", "", "Restore database from backup file")
		cleanupFlag  = flag.Int("cleanup", 0, "Clean up backup files older than N days")
		listFlag     = flag.Bool("list", false, "List available backup files")
	)
	flag.Parse()

	// Execute commands
	if *backupFlag {
		if err := utils.DatabaseBackup(); err != nil {
			log.Fatal("Backup failed:", err)
		}
		log.Println("Backup completed successfully")
		return
	}

	if *restoreFlag != "" {
		if err := utils.RestoreDatabase(*restoreFlag); err != nil {
			log.Fatal("Restore failed:", err)
		}
		log.Println("Restore completed successfully")
		return
	}

	if *cleanupFlag > 0 {
		if err := utils.CleanupOldBackups(*cleanupFlag); err != nil {
			log.Fatal("Cleanup failed:", err)
		}
		log.Println("Cleanup completed successfully")
		return
	}

	if *listFlag {
		listBackupFiles()
		return
	}

	// Show usage if no valid command provided
	flag.Usage()
	os.Exit(1)
}

func listBackupFiles() {
	backupDir := "backups"
	
	files, err := os.ReadDir(backupDir)
	if err != nil {
		log.Printf("Failed to read backup directory: %v", err)
		return
	}

	log.Println("Available backup files:")
	for _, file := range files {
		if !file.IsDir() && len(file.Name()) > 20 && file.Name()[:19] == "greenbecak_backup_" {
			info, err := file.Info()
			if err == nil {
				log.Printf("  %s (%d bytes, %s)", file.Name(), info.Size(), info.ModTime().Format("2006-01-02 15:04:05"))
			} else {
				log.Printf("  %s", file.Name())
			}
		}
	}
}
