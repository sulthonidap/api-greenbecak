package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"greenbecak-backend/database"
	"greenbecak-backend/models"
	"greenbecak-backend/pricing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ===== Pengaturan tarif per-km =====

// loadPricingSetting mengambil pengaturan tarif per-km, membuat baris default kalau belum ada.
func loadPricingSetting(db *gorm.DB) (models.PricingSetting, error) {
	var s models.PricingSetting
	err := db.Where("id = ?", 1).Attrs(models.DefaultPricingSetting()).FirstOrCreate(&s).Error
	return s, err
}

// GetPricingSettingsPublic - info tarif per-km untuk halaman pemesanan (tanpa login)
func GetPricingSettingsPublic(c *gin.Context) {
	s, err := loadPricingSetting(database.GetDB())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat pengaturan tarif"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"pricing": gin.H{
		"enabled":                 s.Enabled,
		"base_fare":               s.BaseFare,
		"per_km":                  s.PerKm,
		"min_fare":                s.MinFare,
		"max_distance_km":         s.MaxDistanceKm,
		"night_surcharge_percent": s.NightSurchargePercent,
		"night_start":             s.NightStart,
		"night_end":               s.NightEnd,
		"service_center_lat":      s.ServiceCenterLat,
		"service_center_lng":      s.ServiceCenterLng,
		"service_radius_km":       s.ServiceRadiusKm,
	}})
}

// GetPricingSettings - admin
func GetPricingSettings(c *gin.Context) {
	s, err := loadPricingSetting(database.GetDB())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat pengaturan tarif"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"pricing": s})
}

// UpdatePricingSettings - admin
func UpdatePricingSettings(c *gin.Context) {
	db := database.GetDB()
	current, err := loadPricingSetting(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat pengaturan tarif"})
		return
	}

	var req models.PricingSetting
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid: " + err.Error()})
		return
	}
	req.ID = 1
	req.CreatedAt = current.CreatedAt
	req.NightStart = strings.TrimSpace(req.NightStart)
	req.NightEnd = strings.TrimSpace(req.NightEnd)

	if err := pricing.ValidateSetting(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.Save(&req).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan pengaturan tarif"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Pengaturan tarif per km disimpan", "pricing": req})
}

// ===== Order jemput (customer) =====

type pickupPoints struct {
	PickupLat float64 `json:"pickup_lat" binding:"required"`
	PickupLng float64 `json:"pickup_lng" binding:"required"`
	DropLat   float64 `json:"drop_lat" binding:"required"`
	DropLng   float64 `json:"drop_lng" binding:"required"`
}

type PickupQuoteRequest struct {
	pickupPoints
}

type CreatePickupOrderRequest struct {
	pickupPoints
	PickupAddress string `json:"pickup_address" binding:"required"`
	DropAddress   string `json:"drop_address" binding:"required"`
	PickupNote    string `json:"pickup_note"`
	CustomerName  string `json:"customer_name" binding:"required"`
	CustomerPhone string `json:"customer_phone" binding:"required"`
	Notes         string `json:"notes"`
}

type pickupQuote struct {
	Route     pricing.Route     `json:"route"`
	Breakdown pricing.Breakdown `json:"breakdown"`
	Price     float64           `json:"price"`
}

var errPickupDisabled = errors.New("Layanan jemput sedang tidak tersedia")

// buildQuote memvalidasi titik dan menghitung harga. Dipakai oleh quote dan create
// supaya harga selalu dihitung ulang di server.
func buildQuote(c *gin.Context, s models.PricingSetting, p pickupPoints) (pickupQuote, int, error) {
	if !s.Enabled {
		return pickupQuote{}, http.StatusServiceUnavailable, errPickupDisabled
	}
	if !pricing.ValidLatLng(p.PickupLat, p.PickupLng) || !pricing.ValidLatLng(p.DropLat, p.DropLng) {
		return pickupQuote{}, http.StatusBadRequest, errors.New("Koordinat lokasi tidak valid")
	}
	if s.ServiceRadiusKm > 0 {
		if pricing.HaversineKm(s.ServiceCenterLat, s.ServiceCenterLng, p.PickupLat, p.PickupLng) > s.ServiceRadiusKm {
			return pickupQuote{}, http.StatusUnprocessableEntity, errors.New("Lokasi jemput di luar area layanan")
		}
		if pricing.HaversineKm(s.ServiceCenterLat, s.ServiceCenterLng, p.DropLat, p.DropLng) > s.ServiceRadiusKm {
			return pickupQuote{}, http.StatusUnprocessableEntity, errors.New("Tujuan di luar area layanan")
		}
	}
	if pricing.HaversineKm(p.PickupLat, p.PickupLng, p.DropLat, p.DropLng) < 0.05 {
		return pickupQuote{}, http.StatusUnprocessableEntity, errors.New("Lokasi jemput dan tujuan terlalu dekat")
	}

	route := pricing.GetRoute(c.Request.Context(), p.PickupLat, p.PickupLng, p.DropLat, p.DropLng)
	if route.DistanceKm > s.MaxDistanceKm {
		return pickupQuote{Route: route}, http.StatusUnprocessableEntity,
			errors.New("Jarak perjalanan melebihi batas maksimal " + trimFloat(s.MaxDistanceKm) + " km")
	}

	b := pricing.Calculate(s, route.DistanceKm, time.Now())
	return pickupQuote{Route: route, Breakdown: b, Price: b.Total}, http.StatusOK, nil
}

// QuotePickupOrder - hitung jarak & harga sebelum pesan (tanpa login)
func QuotePickupOrder(c *gin.Context) {
	var req PickupQuoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lokasi jemput dan tujuan wajib diisi"})
		return
	}
	s, err := loadPricingSetting(database.GetDB())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat pengaturan tarif"})
		return
	}
	q, status, err := buildQuote(c, s, req.pickupPoints)
	if err != nil {
		resp := gin.H{"error": err.Error()}
		if q.Route.DistanceKm > 0 {
			resp["route"] = q.Route
		}
		c.JSON(status, resp)
		return
	}
	c.JSON(http.StatusOK, q)
}

var nonDigit = regexp.MustCompile(`\D`)

// normalizePhone mengubah nomor ke format 08xxx.
func normalizePhone(v string) (string, bool) {
	d := nonDigit.ReplaceAllString(v, "")
	switch {
	case strings.HasPrefix(d, "62"):
		d = "0" + d[2:]
	case strings.HasPrefix(d, "8"):
		d = "0" + d
	}
	return d, strings.HasPrefix(d, "08") && len(d) >= 10 && len(d) <= 14
}

// CreatePickupOrder - buat order jemput (tanpa login)
func CreatePickupOrder(c *gin.Context) {
	var req CreatePickupOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Lengkapi lokasi, nama, dan nomor WhatsApp"})
		return
	}
	req.CustomerName = strings.TrimSpace(req.CustomerName)
	if req.CustomerName == "" || len(req.CustomerName) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nama pemesan wajib diisi"})
		return
	}
	phone, ok := normalizePhone(req.CustomerPhone)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nomor WhatsApp tidak valid"})
		return
	}

	db := database.GetDB()
	s, err := loadPricingSetting(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat pengaturan tarif"})
		return
	}
	q, status, err := buildQuote(c, s, req.pickupPoints)
	if err != nil {
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	token := randomHex(16)
	pLat, pLng, dLat, dLng := req.PickupLat, req.PickupLng, req.DropLat, req.DropLng
	order := models.Order{
		OrderNumber:    generateOrderNumber(),
		OrderType:      models.OrderTypeJemput,
		BecakCode:      "",
		PickupLocation: truncate(strings.TrimSpace(req.PickupAddress), 250),
		DropLocation:   truncate(strings.TrimSpace(req.DropAddress), 250),
		PickupLat:      &pLat,
		PickupLng:      &pLng,
		DropLat:        &dLat,
		DropLng:        &dLng,
		PickupNote:     truncate(strings.TrimSpace(req.PickupNote), 200),
		Distance:       q.Route.DistanceKm,
		DurationMin:    q.Route.DurationMin,
		Price:          q.Price,
		Status:         models.OrderStatusPending,
		PaymentStatus:  "pending",
		CustomerName:   req.CustomerName,
		CustomerPhone:  phone,
		Notes:          truncate(strings.TrimSpace(req.Notes), 500),
		TrackingToken:  &token,
	}
	if err := db.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat pesanan"})
		return
	}

	go sendNewOrderNotification(order)

	c.JSON(http.StatusCreated, gin.H{
		"message":        "Pesanan dibuat, sedang mencari pengayuh",
		"order_number":   order.OrderNumber,
		"tracking_token": token,
		"price":          order.Price,
		"distance_km":    order.Distance,
	})
}

func findPickupOrderByToken(db *gorm.DB, token string) (models.Order, error) {
	var order models.Order
	if len(token) != 32 {
		return order, gorm.ErrRecordNotFound
	}
	err := db.Preload("Driver").
		Where("tracking_token = ? AND order_type = ?", token, models.OrderTypeJemput).
		First(&order).Error
	return order, err
}

// TrackPickupOrder - status pesanan jemput untuk customer (tanpa login, pakai token)
func TrackPickupOrder(c *gin.Context) {
	db := database.GetDB()
	order, err := findPickupOrderByToken(db, c.Param("token"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pesanan tidak ditemukan"})
		return
	}

	resp := gin.H{
		"order_number":    order.OrderNumber,
		"status":          order.Status,
		"payment_status":  order.PaymentStatus,
		"price":           order.Price,
		"distance_km":     order.Distance,
		"duration_min":    order.DurationMin,
		"pickup_location": order.PickupLocation,
		"pickup_note":     order.PickupNote,
		"pickup_lat":      order.PickupLat,
		"pickup_lng":      order.PickupLng,
		"drop_location":   order.DropLocation,
		"drop_lat":        order.DropLat,
		"drop_lng":        order.DropLng,
		"customer_name":   order.CustomerName,
		"created_at":      order.CreatedAt,
		"accepted_at":     order.AcceptedAt,
		"completed_at":    order.CompletedAt,
		"cancelled_at":    order.CancelledAt,
	}

	// Info pengayuh hanya ditampilkan setelah pesanan diterima
	if order.DriverID != nil && order.Status != models.OrderStatusPending {
		driver := gin.H{
			"name":        order.Driver.Name,
			"phone":       order.Driver.Phone,
			"driver_code": order.Driver.DriverCode,
		}
		if order.Status == models.OrderStatusAccepted {
			var loc models.DriverLocation
			if err := db.Where("driver_id = ? AND last_seen > ? AND NOT (latitude = 0 AND longitude = 0)", *order.DriverID, time.Now().Add(-5*time.Minute)).
				Order("last_seen DESC").First(&loc).Error; err == nil {
				driver["lat"] = loc.Latitude
				driver["lng"] = loc.Longitude
				driver["last_seen"] = loc.LastSeen
			}
		}
		resp["driver"] = driver
	}

	c.JSON(http.StatusOK, gin.H{"order": resp})
}

// CancelPickupOrder - customer membatalkan pesanan yang belum diterima pengayuh
func CancelPickupOrder(c *gin.Context) {
	db := database.GetDB()
	order, err := findPickupOrderByToken(db, c.Param("token"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pesanan tidak ditemukan"})
		return
	}

	now := time.Now()
	res := db.Model(&models.Order{}).
		Where("id = ? AND status = ? AND driver_id IS NULL", order.ID, models.OrderStatusPending).
		Updates(map[string]interface{}{"status": models.OrderStatusCancelled, "cancelled_at": now})
	if res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membatalkan pesanan"})
		return
	}
	if res.RowsAffected == 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Pesanan sudah diterima pengayuh, hubungi pengayuh untuk membatalkan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Pesanan dibatalkan"})
}

// ===== helpers =====

func randomHex(nBytes int) string {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		// Sangat jarang; pakai waktu sebagai cadangan
		return hex.EncodeToString([]byte(time.Now().Format("150405.000000")))[:nBytes*2]
	}
	return hex.EncodeToString(b)
}

func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max])
}

func trimFloat(f float64) string {
	return strings.Replace(strconv.FormatFloat(f, 'f', -1, 64), ".", ",", 1)
}
