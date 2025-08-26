package monitoring

import (
	"log"
	"time"
)

type Scheduler struct {
	stopChan chan bool
}

var scheduler = &Scheduler{
	stopChan: make(chan bool),
}

// StartHealthCheckScheduler starts periodic health checks
func StartHealthCheckScheduler(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		log.Printf("Health check scheduler started with %v interval", interval)

		for {
			select {
			case <-ticker.C:
				CheckHealthAndAlert()
			case <-scheduler.stopChan:
				log.Println("Health check scheduler stopped")
				return
			}
		}
	}()
}

// StopHealthCheckScheduler stops the health check scheduler
func StopHealthCheckScheduler() {
	scheduler.stopChan <- true
}

// StartAlertCleanupScheduler starts periodic alert cleanup
func StartAlertCleanupScheduler(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		log.Printf("Alert cleanup scheduler started with %v interval", interval)

		for {
			select {
			case <-ticker.C:
				alertManager := GetAlertManager()
				alertManager.ClearOldAlerts(24 * time.Hour) // Clear alerts older than 24 hours
			case <-scheduler.stopChan:
				log.Println("Alert cleanup scheduler stopped")
				return
			}
		}
	}()
}

// StartAllSchedulers starts all monitoring schedulers
func StartAllSchedulers() {
	// Start health check scheduler (every 30 seconds)
	StartHealthCheckScheduler(30 * time.Second)
	
	// Start alert cleanup scheduler (every hour)
	StartAlertCleanupScheduler(1 * time.Hour)
	
	log.Println("All monitoring schedulers started")
}

// StopAllSchedulers stops all monitoring schedulers
func StopAllSchedulers() {
	StopHealthCheckScheduler()
	log.Println("All monitoring schedulers stopped")
}
