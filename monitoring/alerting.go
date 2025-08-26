package monitoring

import (
	"log"
	"time"
)

type AlertLevel string

const (
	AlertLevelInfo    AlertLevel = "info"
	AlertLevelWarning AlertLevel = "warning"
	AlertLevelError   AlertLevel = "error"
	AlertLevelCritical AlertLevel = "critical"
)

type Alert struct {
	ID          string    `json:"id"`
	Level       AlertLevel `json:"level"`
	Message     string    `json:"message"`
	Service     string    `json:"service"`
	Timestamp   time.Time `json:"timestamp"`
	Acknowledged bool     `json:"acknowledged"`
}

type AlertManager struct {
	alerts []Alert
}

var alertManager = &AlertManager{
	alerts: make([]Alert, 0),
}

// NewAlert creates a new alert
func (am *AlertManager) NewAlert(level AlertLevel, message, service string) Alert {
	alert := Alert{
		ID:          generateAlertID(),
		Level:       level,
		Message:     message,
		Service:     service,
		Timestamp:   time.Now(),
		Acknowledged: false,
	}

	am.alerts = append(am.alerts, alert)
	
	// Log alert
	log.Printf("[ALERT] %s - %s: %s", level, service, message)
	
	return alert
}

// GetAlerts returns all alerts
func (am *AlertManager) GetAlerts() []Alert {
	return am.alerts
}

// GetActiveAlerts returns unacknowledged alerts
func (am *AlertManager) GetActiveAlerts() []Alert {
	var activeAlerts []Alert
	for _, alert := range am.alerts {
		if !alert.Acknowledged {
			activeAlerts = append(activeAlerts, alert)
		}
	}
	return activeAlerts
}

// AcknowledgeAlert marks an alert as acknowledged
func (am *AlertManager) AcknowledgeAlert(alertID string) bool {
	for i, alert := range am.alerts {
		if alert.ID == alertID {
			am.alerts[i].Acknowledged = true
			return true
		}
	}
	return false
}

// ClearOldAlerts removes alerts older than specified duration
func (am *AlertManager) ClearOldAlerts(olderThan time.Duration) {
	var newAlerts []Alert
	cutoff := time.Now().Add(-olderThan)
	
	for _, alert := range am.alerts {
		if alert.Timestamp.After(cutoff) {
			newAlerts = append(newAlerts, alert)
		}
	}
	
	am.alerts = newAlerts
}

// CheckHealthAndAlert checks system health and creates alerts if needed
func CheckHealthAndAlert() {
	status := RunHealthCheck()
	
	// Check database health
	if !status.Database {
		alertManager.NewAlert(
			AlertLevelCritical,
			"Database connection failed",
			"database",
		)
	}
	
	// Check API health
	if !status.API {
		alertManager.NewAlert(
			AlertLevelCritical,
			"API service is down",
			"api",
		)
	}
	
	// Check memory health
	if !status.Memory {
		alertManager.NewAlert(
			AlertLevelWarning,
			"Memory usage is high",
			"memory",
		)
	}
	
	// Check disk health
	if !status.Disk {
		alertManager.NewAlert(
			AlertLevelWarning,
			"Disk space is low",
			"disk",
		)
	}
}

// GetAlertManager returns the alert manager instance
func GetAlertManager() *AlertManager {
	return alertManager
}

// generateAlertID generates a unique alert ID
func generateAlertID() string {
	return time.Now().Format("20060102150405")
}
