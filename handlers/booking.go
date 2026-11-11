package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"greenbecak-backend/config"
	"greenbecak-backend/database"
	"greenbecak-backend/models"
	"greenbecak-backend/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ===================== Konfigurasi =====================

func envInt(key string, def int) int {
	if v, err := strconv.Atoi(strings.TrimSpace(os.Getenv(key))); err == nil && v > 0 {
		return v
	}
	return def
}

func envFloat(key string, def float64) float64 {
	if v, err := strconv.ParseFloat(strings.TrimSpace(os.Getenv(key)), 64); err == nil && v > 0 {
		return v
	}
	return def
}

// Lama booking menunggu driver sebelum otomatis batal (menit)
func bookingExpiryMinutes() int { return envInt("BOOKING_EXPIRY_MINUTES", 10) }

// Radius pencarian driver dari titik jemput (km)
func bookingDriverRadiusKm() float64 { return envFloat("BOOKING_DRIVER_RADIUS_KM", 3) }

// Maksimal booking aktif per nomor HP (anti-spam)
func maxActiveBookingsPerPhone() int { return envInt("BOOKING_MAX_ACTIVE_PER_PHONE", 2) }

// Lokasi driver dianggap masih valid jika diperbarui dalam rentang ini
const driverLocationFreshness = 10 * time.Minute

// Order langsung dianggap masih berjalan selama ini (driver tidak menerima booking)
const walkUpBusyWindow = 30 * time.Minute

// ===================== Helper umum =====================

var sourcePattern = regexp.MustCompile(`[^a-z0-9_]`)

// sanitizeSource menormalkan sumber order: huruf kecil, "-" jadi "_", maks 50 karakter.
func sanitizeSource(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "-", "_")
	s = sourcePattern.ReplaceAllString(s, "")
	if len(s) > 50 {
		s = s[:50]
	}
	return s
}

func sanitizeSourceRef(s string) string {
	s = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1
		}
		return r
	}, strings.TrimSpace(s))
	if len([]rune(s)) > 100 {
		s = string([]rune(s)[:100])
	}
	return s
}

func validLatPtr(v *float64) *float64 {
	if v == nil || *v == 0 || *v < -90 || *v > 90 {
		return nil
	}
	return v
}

func validLngPtr(v *float64) *float64 {
	if v == nil || *v == 0 || *v < -180 || *v > 180 {
		return nil
	}
	return v
}

var nonDigit = regexp.MustCompile(`\D`)

// normalizePhone mengubah 62812..., +62 812..., 0812-... menjadi 0812...
func normalizePhone(p string) (string, bool) {
	d := nonDigit.ReplaceAllString(p, "")
	if strings.HasPrefix(d, "62") {
		d = "0" + d[2:]
	} else if strings.HasPrefix(d, "8") {
		d = "0" + d
	}
	if !strings.HasPrefix(d, "08") || len(d) < 10 || len(d) > 14 {
		return "", false
	}
	return d, true
}

func randomToken(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func round1(v float64) float64 { return math.Round(v*10) / 10 }

func floatPtr(v float64) *float64 { return &v }

// CalculateBookingPrice: max(MinFare, BaseFare + km × PerKm), dibulatkan ke atas ke kelipatan Rounding.
func CalculateBookingPrice(rate models.BookingRate, km float64) float64 {
	raw := rate.BaseFare + km*rate.PerKm
	if raw < rate.MinFare {
		raw = rate.MinFare
	}
	if rate.Rounding > 0 {
		raw = math.Ceil(raw/rate.Rounding) * rate.Rounding
	}
	return raw
}

// ===================== Driver: sibuk & terdekat =====================

// driverBusy: driver sedang membawa penumpang (booking aktif, atau order langsung < 30 menit).
func driverBusy(db *gorm.DB, driverID uint, excludeOrderID uint) bool {
	now := time.Now()
	var n int64
	db.Model(&models.Order{}).
		Where("driver_id = ? AND id <> ?", driverID, excludeOrderID).
		Where(
			"(order_type = ? AND status IN ? AND accepted_at > ?) OR (order_type <> ? AND status IN ? AND created_at > ?)",
			models.OrderTypeBooking, []models.OrderStatus{models.OrderStatusAccepted, models.OrderStatusPickedUp}, now.Add(-3*time.Hour),
			models.OrderTypeBooking, []models.OrderStatus{models.OrderStatusPending, models.OrderStatusAccepted}, now.Add(-walkUpBusyWindow),
		).
		Count(&n)
	return n > 0
}

type nearbyDriver struct {
	Driver     models.Driver
	DistanceKm float64
}

// freshDriverLocation mengembalikan lokasi driver kalau masih baru dan valid.
func freshDriverLocation(db *gorm.DB, driverID uint) (*models.DriverLocation, bool) {
	var loc models.DriverLocation
	err := db.Where("driver_id = ? AND last_seen > ? AND latitude <> 0", driverID, time.Now().Add(-driverLocationFreshness)).
		First(&loc).Error
	if err != nil {
		return nil, false
	}
	return &loc, true
}

// findNearbyDrivers: driver aktif, online, jenis kendaraan cocok, tidak sibuk, dalam radius.
func findNearbyDrivers(db *gorm.DB, lat, lng float64, vehicleType string, radiusKm float64) []nearbyDriver {
	var locs []models.DriverLocation
	db.Where("is_online = ? AND last_seen > ? AND latitude <> 0", true, time.Now().Add(-driverLocationFreshness)).Find(&locs)
	if len(locs) == 0 {
		return nil
	}

	distByDriver := map[uint]float64{}
	ids := []uint{}
	for _, l := range locs {
		d := services.HaversineKm(lat, lng, l.Latitude, l.Longitude)
		if d <= radiusKm {
			distByDriver[l.DriverID] = d
			ids = append(ids, l.DriverID)
		}
	}
	if len(ids) == 0 {
		return nil
	}

	var drivers []models.Driver
	q := db.Where("id IN ? AND is_active = ?", ids, true)
	if vehicleType != "" {
		q = q.Where("vehicle_type = ?", vehicleType)
	}
	q.Find(&drivers)

	result := []nearbyDriver{}
	for _, d := range drivers {
		if driverBusy(db, d.ID, 0) {
			continue
		}
		result = append(result, nearbyDriver{Driver: d, DistanceKm: distByDriver[d.ID]})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].DistanceKm < result[j].DistanceKm })
	return result
}

// notifyDriversForBooking mengirim notifikasi ke driver terdekat.
// Kalau tidak ada driver yang lokasinya diketahui di sekitar, kirim ke semua driver
// dengan jenis kendaraan yang sama (fallback, selama aplikasi driver belum mengirim lokasi).
func notifyDriversForBooking(order models.Order) {
	if config.FirebaseService == nil {
		return
	}
	db := database.GetDB()

	targets := []models.Driver{}
	if order.PickupLat != nil && order.PickupLng != nil {
		for _, n := range findNearbyDrivers(db, *order.PickupLat, *order.PickupLng, order.VehicleType, bookingDriverRadiusKm()) {
			targets = append(targets, n.Driver)
		}
	}
	if len(targets) == 0 {
		var all []models.Driver
		q := db.Where("is_active = ? AND fcm_token <> ''", true)
		if order.VehicleType != "" {
			q = q.Where("vehicle_type = ?", order.VehicleType)
		}
		q.Find(&all)
		for _, d := range all {
			if !driverBusy(db, d.ID, 0) {
				targets = append(targets, d)
			}
		}
		log.Printf("Booking %s: tidak ada driver terdekat dengan lokasi, fallback ke %d driver", order.OrderNumber, len(targets))
	}

	notification := services.FCMMessageNotification{
		Title: "🛺 Booking Baru!",
		Body:  fmt.Sprintf("%s → %s · %.1f km · Rp %s", order.PickupLocation, order.DropLocation, order.Distance, formatRupiah(order.Price)),
	}
	data := map[string]string{
		"type":            "new_booking",
		"orderId":         fmt.Sprintf("%d", order.ID),
		"price":           fmt.Sprintf("%.0f", order.Price),
		"pickup_location": order.PickupLocation,
		"drop_location":   order.DropLocation,
		"distance":        fmt.Sprintf("%.1f", order.Distance),
	}
	for _, d := range targets {
		if d.FCMToken == "" {
			continue
		}
		if err := config.FirebaseService.SendToDevice(d.FCMToken, notification, data); err != nil {
			log.Printf("Gagal kirim notifikasi booking ke driver %d: %v", d.ID, err)
		}
	}
}

func notifyDriver(driverID uint, title, body string, data map[string]string) {
	if config.FirebaseService == nil {
		return
	}
	var d models.Driver
	if err := database.GetDB().First(&d, driverID).Error; err != nil || d.FCMToken == "" {
		return
	}
	if err := config.FirebaseService.SendToDevice(d.FCMToken, services.FCMMessageNotification{Title: title, Body: body}, data); err != nil {
		log.Printf("Gagal kirim notifikasi ke driver %d: %v", driverID, err)
	}
}

func formatRupiah(v float64) string {
	s := strconv.FormatInt(int64(math.Round(v)), 10)
	var out []byte
	for i, c := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, '.')
		}
		out = append(out, c)
	}
	return string(out)
}

// ===================== Quote (cek harga) =====================

type GeoPoint struct {
	Lat     float64 `json:"lat" binding:"required"`
	Lng     float64 `json:"lng" binding:"required"`
	Name    string  `json:"name"`
	Address string  `json:"address"`
}

func (p GeoPoint) label() string {
	if n := strings.TrimSpace(p.Name); n != "" {
		return n
	}
	if a := strings.TrimSpace(p.Address); a != "" {
		return a
	}
	return fmt.Sprintf("%.5f, %.5f", p.Lat, p.Lng)
}

type BookingQuoteRequest struct {
	Pickup      GeoPoint `json:"pickup" binding:"required"`
	Drop        GeoPoint `json:"drop" binding:"required"`
	VehicleType string   `json:"vehicle_type"`
}

type BookingQuote struct {
	VehicleType    string  `json:"vehicle_type"`
	Label          string  `json:"label"`
	DistanceKm     float64 `json:"distance_km"`
	DurationMin    int     `json:"duration_min"`
	DistanceSource string  `json:"distance_source"`
	Price          float64 `json:"price"`
	BaseFare       float64 `json:"base_fare"`
	PerKm          float64 `json:"per_km"`
	MinFare        float64 `json:"min_fare"`
	NearbyDrivers  int     `json:"nearby_drivers"`
}

type quoteError struct {
	status  int
	message string
}

func (e *quoteError) Error() string { return e.message }

func computeQuote(db *gorm.DB, req BookingQuoteRequest) (*BookingQuote, *models.BookingRate, error) {
	area := services.ServiceArea()
	if !area.Contains(req.Pickup.Lat, req.Pickup.Lng) {
		return nil, nil, &quoteError{http.StatusBadRequest, "Titik jemput berada di luar area layanan"}
	}
	if !area.Contains(req.Drop.Lat, req.Drop.Lng) {
		return nil, nil, &quoteError{http.StatusBadRequest, "Tujuan berada di luar area layanan"}
	}

	var rate models.BookingRate
	q := db.Where("is_active = ?", true)
	if vt := strings.TrimSpace(req.VehicleType); vt != "" {
		q = q.Where("vehicle_type = ?", vt)
	}
	if err := q.Order("id ASC").First(&rate).Error; err != nil {
		return nil, nil, &quoteError{http.StatusBadRequest, "Jenis kendaraan tidak tersedia untuk booking"}
	}

	route := services.RouteDistance(req.Pickup.Lat, req.Pickup.Lng, req.Drop.Lat, req.Drop.Lng)
	km := round1(route.DistanceKm)
	if km < 0.1 {
		return nil, nil, &quoteError{http.StatusBadRequest, "Titik jemput dan tujuan terlalu dekat"}
	}
	if rate.MaxDistanceKm > 0 && km > rate.MaxDistanceKm {
		return nil, nil, &quoteError{http.StatusBadRequest,
			fmt.Sprintf("Jarak %.1f km melebihi batas maksimal booking %.0f km", km, rate.MaxDistanceKm)}
	}

	nearby := findNearbyDrivers(db, req.Pickup.Lat, req.Pickup.Lng, rate.VehicleType, bookingDriverRadiusKm())

	return &BookingQuote{
		VehicleType:    rate.VehicleType,
		Label:          rate.Label,
		DistanceKm:     km,
		DurationMin:    int(math.Ceil(route.DurationMin)),
		DistanceSource: route.Source,
		Price:          CalculateBookingPrice(rate, km),
		BaseFare:       rate.BaseFare,
		PerKm:          rate.PerKm,
		MinFare:        rate.MinFare,
		NearbyDrivers:  len(nearby),
	}, &rate, nil
}

func respondQuoteError(c *gin.Context, err error) {
	var qe *quoteError
	if errors.As(err, &qe) {
		c.JSON(qe.status, gin.H{"error": qe.message})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghitung harga"})
}

// POST /api/booking/quote
func GetBookingQuote(c *gin.Context) {
	var req BookingQuoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Titik jemput dan tujuan wajib diisi"})
		return
	}
	quote, _, err := computeQuote(database.GetDB(), req)
	if err != nil {
		respondQuoteError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"quote": quote})
}

// ===================== Konfigurasi booking =====================

// GET /api/booking/config?hotel=<slug>
func GetBookingConfig(c *gin.Context) {
	db := database.GetDB()

	var rates []models.BookingRate
	db.Where("is_active = ?", true).Order("id ASC").Find(&rates)

	var destinations []models.Destination
	db.Where("is_active = ?", true).Order("sort_order ASC, name ASC").Find(&destinations)

	resp := gin.H{
		"rates":          rates,
		"destinations":   destinations,
		"service_area":   services.ServiceArea(),
		"expiry_minutes": bookingExpiryMinutes(),
	}

	if slug := strings.TrimSpace(c.Query("hotel")); slug != "" {
		var hotel models.Hotel
		if err := db.Where("slug = ? AND is_active = ?", slug, true).First(&hotel).Error; err == nil {
			resp["hotel"] = hotel
		}
	}

	c.JSON(http.StatusOK, resp)
}

// GET /api/hotels/public/:slug
func GetHotelPublic(c *gin.Context) {
	var hotel models.Hotel
	if err := database.GetDB().Where("slug = ? AND is_active = ?", c.Param("slug"), true).First(&hotel).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Hotel tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"hotel": hotel})
}

// ===================== Buat booking =====================

type CreateBookingRequest struct {
	Pickup        GeoPoint `json:"pickup" binding:"required"`
	Drop          GeoPoint `json:"drop" binding:"required"`
	VehicleType   string   `json:"vehicle_type"`
	CustomerName  string   `json:"customer_name" binding:"required"`
	CustomerPhone string   `json:"customer_phone" binding:"required"`
	Notes         string   `json:"notes"`
	Source        string   `json:"source"`
	SourceRef     string   `json:"source_ref"`
	HotelSlug     string   `json:"hotel_slug"`
	CustomerLat   *float64 `json:"customer_lat"`
	CustomerLng   *float64 `json:"customer_lng"`
	// Harga yang dilihat penumpang; jika berbeda dari hitungan server, booking ditolak agar penumpang cek ulang
	ExpectedPrice *float64 `json:"expected_price"`
}

// POST /api/booking
func CreateBooking(c *gin.Context) {
	var req CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data booking belum lengkap: titik jemput, tujuan, nama, dan nomor WA wajib diisi"})
		return
	}

	name := strings.TrimSpace(req.CustomerName)
	if name == "" || len([]rune(name)) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nama tidak valid"})
		return
	}
	phone, ok := normalizePhone(req.CustomerPhone)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nomor WhatsApp tidak valid"})
		return
	}
	notes := strings.TrimSpace(req.Notes)
	if len([]rune(notes)) > 300 {
		notes = string([]rune(notes)[:300])
	}

	db := database.GetDB()

	// Anti-spam: batasi booking aktif per nomor
	var active int64
	db.Model(&models.Order{}).
		Where("customer_phone = ? AND order_type = ? AND status IN ?", phone, models.OrderTypeBooking,
			[]models.OrderStatus{models.OrderStatusPending, models.OrderStatusAccepted, models.OrderStatusPickedUp}).
		Count(&active)
	if int(active) >= maxActiveBookingsPerPhone() {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Nomor ini masih punya booking yang aktif. Selesaikan atau batalkan dulu."})
		return
	}

	// Asal order
	source := sanitizeSource(req.Source)
	sourceRef := sanitizeSourceRef(req.SourceRef)
	var hotelID *uint
	if slug := strings.TrimSpace(req.HotelSlug); slug != "" {
		var hotel models.Hotel
		if err := db.Where("slug = ? AND is_active = ?", slug, true).First(&hotel).Error; err == nil {
			hotelID = &hotel.ID
			if source == "" {
				source = models.OrderSourceQRHotel
			}
			if sourceRef == "" {
				sourceRef = hotel.Slug
			}
		}
	}
	if source == "" {
		source = models.OrderSourceWeb
	}

	quote, _, err := computeQuote(db, BookingQuoteRequest{Pickup: req.Pickup, Drop: req.Drop, VehicleType: req.VehicleType})
	if err != nil {
		respondQuoteError(c, err)
		return
	}
	if req.ExpectedPrice != nil && math.Abs(*req.ExpectedPrice-quote.Price) >= 1 {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Harga berubah, silakan cek kembali sebelum memesan",
			"quote": quote,
		})
		return
	}

	token, err := randomToken(16)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat booking"})
		return
	}
	expires := time.Now().Add(time.Duration(bookingExpiryMinutes()) * time.Minute)

	order := models.Order{
		OrderNumber:    generateOrderNumber(),
		BecakCode:      "",
		PickupLocation: req.Pickup.label(),
		DropLocation:   req.Drop.label(),
		Distance:       quote.DistanceKm,
		Price:          quote.Price,
		Status:         models.OrderStatusPending,
		PaymentStatus:  "pending",
		CustomerPhone:  phone,
		CustomerName:   name,
		Notes:          notes,
		OrderType:      models.OrderTypeBooking,
		Source:         source,
		SourceRef:      sourceRef,
		HotelID:        hotelID,
		PickupLat:      floatPtr(req.Pickup.Lat),
		PickupLng:      floatPtr(req.Pickup.Lng),
		DropLat:        floatPtr(req.Drop.Lat),
		DropLng:        floatPtr(req.Drop.Lng),
		CustomerLat:    validLatPtr(req.CustomerLat),
		CustomerLng:    validLngPtr(req.CustomerLng),
		VehicleType:    quote.VehicleType,
		DistanceSource: quote.DistanceSource,
		TrackingToken:  &token,
		ExpiresAt:      &expires,
	}

	if err := db.Create(&order).Error; err != nil {
		log.Printf("CreateBooking: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat booking"})
		return
	}

	go notifyDriversForBooking(order)

	c.JSON(http.StatusCreated, gin.H{
		"message":        "Booking berhasil dibuat, sedang mencari becak",
		"tracking_token": token,
		"booking":        bookingStatusView(db, order),
	})
}

// ===================== Status & batal (penumpang) =====================

var bookingStatusLabels = map[models.OrderStatus]string{
	models.OrderStatusPending:   "Mencari becak",
	models.OrderStatusAccepted:  "Driver menuju lokasi jemput",
	models.OrderStatusPickedUp:  "Dalam perjalanan",
	models.OrderStatusCompleted: "Selesai",
	models.OrderStatusCancelled: "Dibatalkan",
}

func bookingStatusView(db *gorm.DB, order models.Order) gin.H {
	view := gin.H{
		"order_number":    order.OrderNumber,
		"status":          order.Status,
		"status_label":    bookingStatusLabels[order.Status],
		"pickup":          gin.H{"name": order.PickupLocation, "lat": order.PickupLat, "lng": order.PickupLng},
		"drop":            gin.H{"name": order.DropLocation, "lat": order.DropLat, "lng": order.DropLng},
		"distance_km":     order.Distance,
		"distance_source": order.DistanceSource,
		"price":           order.Price,
		"vehicle_type":    order.VehicleType,
		"customer_name":   order.CustomerName,
		"payment_status":  order.PaymentStatus,
		"created_at":      order.CreatedAt,
		"expires_at":      order.ExpiresAt,
		"accepted_at":     order.AcceptedAt,
		"picked_up_at":    order.PickedUpAt,
		"completed_at":    order.CompletedAt,
		"cancelled_at":    order.CancelledAt,
		"cancel_reason":   order.CancelReason,
		"can_cancel":      order.Status == models.OrderStatusPending || order.Status == models.OrderStatusAccepted,
	}

	showDriver := order.DriverID != nil && (order.Status == models.OrderStatusAccepted ||
		order.Status == models.OrderStatusPickedUp || order.Status == models.OrderStatusCompleted)
	if showDriver {
		var d models.Driver
		if err := db.First(&d, *order.DriverID).Error; err == nil {
			driver := gin.H{
				"name":           d.Name,
				"phone":          d.Phone,
				"vehicle_number": d.VehicleNumber,
				"vehicle_type":   d.VehicleType,
				"rating":         d.Rating,
			}
			if order.Status != models.OrderStatusCompleted {
				if loc, ok := freshDriverLocation(db, d.ID); ok {
					driver["location"] = gin.H{"lat": loc.Latitude, "lng": loc.Longitude, "last_seen": loc.LastSeen}
					if order.Status == models.OrderStatusAccepted && order.PickupLat != nil && order.PickupLng != nil {
						dist := services.HaversineKm(loc.Latitude, loc.Longitude, *order.PickupLat, *order.PickupLng) * 1.3
						driver["distance_to_pickup_km"] = round1(dist)
						driver["eta_min"] = int(math.Ceil(dist / 12 * 60))
					}
				}
			}
			view["driver"] = driver
		}
	}
	return view
}

func findBookingByToken(db *gorm.DB, token string) (*models.Order, error) {
	token = strings.TrimSpace(token)
	if len(token) < 16 {
		return nil, gorm.ErrRecordNotFound
	}
	var order models.Order
	if err := db.Where("tracking_token = ? AND order_type = ?", token, models.OrderTypeBooking).First(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

// GET /api/booking/status/:token
func GetBookingStatus(c *gin.Context) {
	db := database.GetDB()
	order, err := findBookingByToken(db, c.Param("token"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking tidak ditemukan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"booking": bookingStatusView(db, *order)})
}

// POST /api/booking/status/:token/cancel
func CancelBookingByCustomer(c *gin.Context) {
	var req struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&req)

	db := database.GetDB()
	order, err := findBookingByToken(db, c.Param("token"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking tidak ditemukan"})
		return
	}

	reason := sanitizeSourceRef(req.Reason)
	if reason == "" {
		reason = "Dibatalkan penumpang"
	}
	now := time.Now()
	res := db.Model(&models.Order{}).
		Where("id = ? AND status IN ?", order.ID, []models.OrderStatus{models.OrderStatusPending, models.OrderStatusAccepted}).
		Updates(map[string]interface{}{
			"status":        models.OrderStatusCancelled,
			"cancelled_at":  now,
			"cancel_reason": reason,
		})
	if res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membatalkan booking"})
		return
	}
	if res.RowsAffected == 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Booking tidak bisa dibatalkan karena penumpang sudah dijemput atau booking sudah selesai"})
		return
	}

	if order.DriverID != nil && order.Status == models.OrderStatusAccepted {
		db.Model(&models.Driver{}).Where("id = ?", *order.DriverID).Update("status", models.DriverStatusActive)
		go notifyDriver(*order.DriverID, "Booking dibatalkan", fmt.Sprintf("Penumpang membatalkan booking %s", order.OrderNumber),
			map[string]string{"type": "booking_cancelled", "orderId": fmt.Sprintf("%d", order.ID)})
	}

	db.First(order, order.ID)
	c.JSON(http.StatusOK, gin.H{"message": "Booking dibatalkan", "booking": bookingStatusView(db, *order)})
}

// ===================== Pencarian alamat =====================

// GET /api/geo/search?q=kraton&lat=&lng=
func SearchPlaces(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	if len([]rune(q)) < 2 {
		c.JSON(http.StatusOK, gin.H{"results": []services.Place{}})
		return
	}
	if len([]rune(q)) > 100 {
		q = string([]rune(q)[:100])
	}
	biasLat, _ := strconv.ParseFloat(c.Query("lat"), 64)
	biasLng, _ := strconv.ParseFloat(c.Query("lng"), 64)

	db := database.GetDB()
	results := []services.Place{}
	seen := map[string]bool{}

	// 1. Destinasi populer dari admin (paling atas)
	var dests []models.Destination
	like := "%" + strings.ReplaceAll(strings.ReplaceAll(q, "%", ""), "_", "") + "%"
	db.Where("is_active = ? AND (name LIKE ? OR address LIKE ?)", true, like, like).
		Order("sort_order ASC, name ASC").Limit(5).Find(&dests)
	for _, d := range dests {
		results = append(results, services.Place{
			Name: d.Name, Address: d.Address, Lat: d.Latitude, Lng: d.Longitude, Type: "destination", Category: d.Category,
		})
		seen[strings.ToLower(d.Name)] = true
	}

	// 2. OpenStreetMap
	resp := gin.H{}
	places, err := services.SearchPlaces(q, biasLat, biasLng, 8)
	if err != nil {
		log.Printf("SearchPlaces: %v", err)
		resp["warning"] = "Pencarian alamat sedang bermasalah, silakan pilih lewat peta"
	}
	for _, p := range places {
		key := strings.ToLower(p.Name)
		if seen[key] {
			continue
		}
		seen[key] = true
		results = append(results, p)
	}

	resp["results"] = results
	c.JSON(http.StatusOK, resp)
}

// GET /api/geo/reverse?lat=&lng=
func ReverseGeocode(c *gin.Context) {
	lat, err1 := strconv.ParseFloat(c.Query("lat"), 64)
	lng, err2 := strconv.ParseFloat(c.Query("lng"), 64)
	if err1 != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "lat dan lng wajib diisi"})
		return
	}
	inArea := services.ServiceArea().Contains(lat, lng)

	// Kalau pin dekat (≤ 80 m) dengan destinasi populer, pakai nama destinasi
	db := database.GetDB()
	var dests []models.Destination
	db.Where("is_active = ? AND latitude BETWEEN ? AND ? AND longitude BETWEEN ? AND ?",
		true, lat-0.001, lat+0.001, lng-0.001, lng+0.001).Find(&dests)
	for _, d := range dests {
		if services.HaversineKm(lat, lng, d.Latitude, d.Longitude) <= 0.08 {
			c.JSON(http.StatusOK, gin.H{"place": services.Place{
				Name: d.Name, Address: d.Address, Lat: lat, Lng: lng, Type: "destination", Category: d.Category,
			}, "in_service_area": inArea})
			return
		}
	}

	place, err := services.ReverseGeocode(lat, lng)
	if err != nil {
		log.Printf("ReverseGeocode: %v", err)
		place = services.Place{Name: fmt.Sprintf("Titik %.5f, %.5f", lat, lng), Lat: lat, Lng: lng, Type: "pin"}
	}
	c.JSON(http.StatusOK, gin.H{"place": place, "in_service_area": inArea})
}

// ===================== Auto-batal booking =====================

// ExpirePendingBookings membatalkan booking yang tidak diambil driver sampai batas waktunya.
func ExpirePendingBookings() {
	db := database.GetDB()
	if db == nil {
		return
	}
	now := time.Now()
	res := db.Model(&models.Order{}).
		Where("order_type = ? AND status = ? AND expires_at IS NOT NULL AND expires_at < ?",
			models.OrderTypeBooking, models.OrderStatusPending, now).
		Updates(map[string]interface{}{
			"status":        models.OrderStatusCancelled,
			"cancelled_at":  now,
			"cancel_reason": "Tidak ada driver yang menerima",
		})
	if res.Error != nil {
		log.Printf("ExpirePendingBookings: %v", res.Error)
	} else if res.RowsAffected > 0 {
		log.Printf("ExpirePendingBookings: %d booking dibatalkan otomatis", res.RowsAffected)
	}
}

// StartBookingExpiryScheduler menjalankan ExpirePendingBookings setiap menit.
func StartBookingExpiryScheduler() {
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		log.Println("Booking expiry scheduler started")
		for range ticker.C {
			ExpirePendingBookings()
		}
	}()
}
