package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// Firebase Cloud Messaging Service
// ===============================
// Service untuk kirim push notifications ke driver app
// Menggunakan Firebase Cloud Messaging (FCM) HTTP v1 API

type FirebaseService struct {
	serverKey          string
	projectID          string
	serviceAccountPath string
	oauthClient        *http.Client
}

// FCM Message Structure
type FCMMessage struct {
	Message struct {
		Token        string                  `json:"token,omitempty"`
		Topic        string                  `json:"topic,omitempty"`
		Notification *FCMMessageNotification `json:"notification,omitempty"`
		Data         map[string]string       `json:"data,omitempty"`
		Android      *FCMAndroidConfig       `json:"android,omitempty"`
		APNS         *FCMAPNSConfig          `json:"apns,omitempty"`
	} `json:"message"`
}

type FCMMessageNotification struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

type FCMAndroidConfig struct {
	Priority     string                  `json:"priority,omitempty"`
	Notification *FCMAndroidNotification `json:"notification,omitempty"`
	Data         map[string]string       `json:"data,omitempty"`
}

type FCMAndroidNotification struct {
	Title       string `json:"title"`
	Body        string `json:"body"`
	Sound       string `json:"sound,omitempty"`
	Priority    string `json:"priority,omitempty"`
	ChannelID   string `json:"channel_id,omitempty"`
	ClickAction string `json:"click_action,omitempty"`
}

type FCMAPNSConfig struct {
	Payload *FCMAPNSPayload `json:"payload,omitempty"`
}

type FCMAPNSPayload struct {
	APS *FCMAPSAps `json:"aps,omitempty"`
}

type FCMAPSAps struct {
	Alert *FCMAPSAlert `json:"alert,omitempty"`
	Sound string       `json:"sound,omitempty"`
	Badge int          `json:"badge,omitempty"`
}

type FCMAPSAlert struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// New Firebase Service
func NewFirebaseService(serverKey, projectID string) *FirebaseService {
	return &FirebaseService{
		serverKey: serverKey,
		projectID: projectID,
	}
}

// New Firebase Service with Service Account
func NewFirebaseServiceWithServiceAccount(serviceAccountPath, projectID string) *FirebaseService {
	return &FirebaseService{
		serviceAccountPath: serviceAccountPath,
		projectID:          projectID,
	}
}

// Send notification to specific device
func (fs *FirebaseService) SendToDevice(token string, notification FCMMessageNotification, data map[string]string) error {
	message := FCMMessage{}
	message.Message.Token = token
	message.Message.Notification = &notification
	message.Message.Data = data

	// Add Android specific config
	message.Message.Android = &FCMAndroidConfig{
		Priority: "high",
		Notification: &FCMAndroidNotification{
			Title:       notification.Title,
			Body:        notification.Body,
			Sound:       "default",
			Priority:    "high",
			ChannelID:   "orders",
			ClickAction: "FLUTTER_NOTIFICATION_CLICK",
		},
		Data: data,
	}

	return fs.sendMessage(message)
}

// Send notification to topic
func (fs *FirebaseService) SendToTopic(topic string, notification FCMMessageNotification, data map[string]string) error {
	message := FCMMessage{}
	message.Message.Topic = topic
	message.Message.Notification = &notification
	message.Message.Data = data

	// Add Android specific config
	message.Message.Android = &FCMAndroidConfig{
		Priority: "high",
		Notification: &FCMAndroidNotification{
			Title:       notification.Title,
			Body:        notification.Body,
			Sound:       "default",
			Priority:    "high",
			ChannelID:   "orders",
			ClickAction: "FLUTTER_NOTIFICATION_CLICK",
		},
		Data: data,
	}

	return fs.sendMessage(message)
}

// Send message to FCM using HTTP v1 API
func (fs *FirebaseService) sendMessage(message FCMMessage) error {
	// HTTP v1 API endpoint format: https://fcm.googleapis.com/v1/projects/{project-id}/messages:send
	url := fmt.Sprintf("https://fcm.googleapis.com/v1/projects/%s/messages:send", fs.projectID)

	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal FCM message: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Authentication: Try Service Account first, then fallback to legacy server key
	if fs.serviceAccountPath != "" {
		// TODO: Implement OAuth 2.0 token from service account JSON
		// For now, return error to guide user to use legacy approach
		return fmt.Errorf("service account authentication not yet implemented. Please use legacy server key for now")
	} else if fs.serverKey != "" {
		// Legacy approach (deprecated but functional)
		req.Header.Set("Authorization", "Bearer "+fs.serverKey)
	} else {
		return fmt.Errorf("no authentication method configured. Please set FIREBASE_SERVER_KEY or implement service account")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send FCM request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("FCM request failed with status: %d, body: %s", resp.StatusCode, string(body))
	}

	log.Printf("FCM message sent successfully using HTTP v1 API")
	return nil
}

// Send new order notification to driver
func (fs *FirebaseService) SendNewOrderNotification(driverToken string, orderData map[string]interface{}) error {
	notification := FCMMessageNotification{
		Title: "🚗 Pesanan Baru!",
		Body:  fmt.Sprintf("Rp %v - %v", orderData["price"], orderData["pickup_location"]),
	}

	data := map[string]string{
		"type":            "new_order",
		"orderId":         fmt.Sprintf("%v", orderData["id"]),
		"price":           fmt.Sprintf("%v", orderData["price"]),
		"pickup_location": fmt.Sprintf("%v", orderData["pickup_location"]),
		"drop_location":   fmt.Sprintf("%v", orderData["drop_location"]),
		"distance":        fmt.Sprintf("%v", orderData["distance"]),
		"eta":             fmt.Sprintf("%v", orderData["eta"]),
	}

	return fs.SendToDevice(driverToken, notification, data)
}

// Send order accepted notification
func (fs *FirebaseService) SendOrderAcceptedNotification(driverToken string, orderData map[string]interface{}) error {
	notification := FCMMessageNotification{
		Title: "✅ Pesanan Diterima",
		Body:  "Pesanan telah diterima dan sedang dalam proses",
	}

	data := map[string]string{
		"type":    "order_accepted",
		"orderId": fmt.Sprintf("%v", orderData["id"]),
	}

	return fs.SendToDevice(driverToken, notification, data)
}

// Send order completed notification
func (fs *FirebaseService) SendOrderCompletedNotification(driverToken string, orderData map[string]interface{}) error {
	notification := FCMMessageNotification{
		Title: "🎉 Pesanan Selesai",
		Body:  "Pesanan telah selesai dan pembayaran telah diterima",
	}

	data := map[string]string{
		"type":    "order_completed",
		"orderId": fmt.Sprintf("%v", orderData["id"]),
		"price":   fmt.Sprintf("%v", orderData["price"]),
	}

	return fs.SendToDevice(driverToken, notification, data)
}

// Send withdrawal approved notification
func (fs *FirebaseService) SendWithdrawalApprovedNotification(driverToken string, withdrawalData map[string]interface{}) error {
	notification := FCMMessageNotification{
		Title: "💰 Penarikan Disetujui",
		Body:  "Permintaan penarikan dana Anda telah disetujui",
	}

	data := map[string]string{
		"type":         "withdrawal_approved",
		"withdrawalId": fmt.Sprintf("%v", withdrawalData["id"]),
		"amount":       fmt.Sprintf("%v", withdrawalData["amount"]),
	}

	return fs.SendToDevice(driverToken, notification, data)
}

// Send notification to all online drivers
func (fs *FirebaseService) SendToAllOnlineDrivers(notification FCMMessageNotification, data map[string]string) error {
	// Send to topic for all online drivers
	return fs.SendToTopic("online_drivers", notification, data)
}

// Send notification to drivers in specific area
func (fs *FirebaseService) SendToDriversInArea(area string, notification FCMMessageNotification, data map[string]string) error {
	// Send to topic for specific area
	topic := fmt.Sprintf("drivers_area_%s", area)
	return fs.SendToTopic(topic, notification, data)
}
