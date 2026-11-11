package handlers

import (
	"net/http"
	"regexp"
	"strings"

	"greenbecak-backend/database"
	"greenbecak-backend/models"
	"greenbecak-backend/services"

	"github.com/gin-gonic/gin"
)

// ===================== Hotel =====================

type HotelRequest struct {
	Name      string  `json:"name" binding:"required"`
	Slug      string  `json:"slug"`
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
	Phone     string  `json:"phone"`
	IsActive  *bool   `json:"is_active"`
}

var slugInvalid = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	s = slugInvalid.ReplaceAllString(strings.ToLower(strings.TrimSpace(s)), "-")
	s = strings.Trim(s, "-")
	if len(s) > 90 {
		s = strings.Trim(s[:90], "-")
	}
	return s
}

func checkInServiceArea(c *gin.Context, lat, lng float64) bool {
	if !services.ServiceArea().Contains(lat, lng) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Koordinat berada di luar area layanan"})
		return false
	}
	return true
}

// GET /api/admin/hotels
func GetHotels(c *gin.Context) {
	db := database.GetDB()
	var hotels []models.Hotel
	db.Order("name ASC").Find(&hotels)

	// Jumlah order per hotel
	type countRow struct {
		HotelID uint
		Total   int64
	}
	var rows []countRow
	db.Model(&models.Order{}).Select("hotel_id, COUNT(*) AS total").
		Where("hotel_id IS NOT NULL").Group("hotel_id").Scan(&rows)
	counts := map[uint]int64{}
	for _, r := range rows {
		counts[r.HotelID] = r.Total
	}

	result := make([]gin.H, 0, len(hotels))
	for _, h := range hotels {
		result = append(result, gin.H{
			"id": h.ID, "name": h.Name, "slug": h.Slug, "address": h.Address,
			"latitude": h.Latitude, "longitude": h.Longitude, "phone": h.Phone,
			"is_active": h.IsActive, "created_at": h.CreatedAt, "updated_at": h.UpdatedAt,
			"total_orders": counts[h.ID],
		})
	}
	c.JSON(http.StatusOK, gin.H{"hotels": result})
}

// POST /api/admin/hotels
func CreateHotel(c *gin.Context) {
	var req HotelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nama dan koordinat hotel wajib diisi"})
		return
	}
	if !checkInServiceArea(c, req.Latitude, req.Longitude) {
		return
	}
	slug := slugify(req.Slug)
	if slug == "" {
		slug = slugify(req.Name)
	}
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Slug tidak valid"})
		return
	}

	db := database.GetDB()
	var exists int64
	db.Model(&models.Hotel{}).Unscoped().Where("slug = ?", slug).Count(&exists)
	if exists > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Slug hotel sudah dipakai, gunakan slug lain"})
		return
	}

	hotel := models.Hotel{
		Name: strings.TrimSpace(req.Name), Slug: slug, Address: strings.TrimSpace(req.Address),
		Latitude: req.Latitude, Longitude: req.Longitude, Phone: strings.TrimSpace(req.Phone), IsActive: true,
	}
	if req.IsActive != nil {
		hotel.IsActive = *req.IsActive
	}
	if err := db.Create(&hotel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan hotel"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Hotel ditambahkan", "hotel": hotel})
}

// PUT /api/admin/hotels/:id
func UpdateHotel(c *gin.Context) {
	db := database.GetDB()
	var hotel models.Hotel
	if err := db.First(&hotel, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Hotel tidak ditemukan"})
		return
	}
	var req HotelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nama dan koordinat hotel wajib diisi"})
		return
	}
	if !checkInServiceArea(c, req.Latitude, req.Longitude) {
		return
	}
	if slug := slugify(req.Slug); slug != "" && slug != hotel.Slug {
		var exists int64
		db.Model(&models.Hotel{}).Unscoped().Where("slug = ? AND id <> ?", slug, hotel.ID).Count(&exists)
		if exists > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Slug hotel sudah dipakai"})
			return
		}
		// Catatan: mengubah slug membuat QR hotel lama tidak dikenali lagi
		hotel.Slug = slug
	}
	hotel.Name = strings.TrimSpace(req.Name)
	hotel.Address = strings.TrimSpace(req.Address)
	hotel.Latitude = req.Latitude
	hotel.Longitude = req.Longitude
	hotel.Phone = strings.TrimSpace(req.Phone)
	if req.IsActive != nil {
		hotel.IsActive = *req.IsActive
	}
	if err := db.Save(&hotel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan hotel"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Hotel diperbarui", "hotel": hotel})
}

// DELETE /api/admin/hotels/:id (soft delete; riwayat order tetap ada)
func DeleteHotel(c *gin.Context) {
	if err := database.GetDB().Delete(&models.Hotel{}, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus hotel"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Hotel dihapus"})
}

// ===================== Destinasi =====================

type DestinationRequest struct {
	Name      string  `json:"name" binding:"required"`
	Category  string  `json:"category"`
	Address   string  `json:"address"`
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
	SortOrder int     `json:"sort_order"`
	IsActive  *bool   `json:"is_active"`
}

// GET /api/admin/destinations
func GetDestinations(c *gin.Context) {
	var dests []models.Destination
	database.GetDB().Order("sort_order ASC, name ASC").Find(&dests)
	c.JSON(http.StatusOK, gin.H{"destinations": dests})
}

// POST /api/admin/destinations
func CreateDestination(c *gin.Context) {
	var req DestinationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nama dan koordinat destinasi wajib diisi"})
		return
	}
	if !checkInServiceArea(c, req.Latitude, req.Longitude) {
		return
	}
	d := models.Destination{
		Name: strings.TrimSpace(req.Name), Category: strings.TrimSpace(req.Category), Address: strings.TrimSpace(req.Address),
		Latitude: req.Latitude, Longitude: req.Longitude, SortOrder: req.SortOrder, IsActive: true,
	}
	if req.IsActive != nil {
		d.IsActive = *req.IsActive
	}
	if err := database.GetDB().Create(&d).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan destinasi"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Destinasi ditambahkan", "destination": d})
}

// PUT /api/admin/destinations/:id
func UpdateDestination(c *gin.Context) {
	db := database.GetDB()
	var d models.Destination
	if err := db.First(&d, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Destinasi tidak ditemukan"})
		return
	}
	var req DestinationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nama dan koordinat destinasi wajib diisi"})
		return
	}
	if !checkInServiceArea(c, req.Latitude, req.Longitude) {
		return
	}
	d.Name = strings.TrimSpace(req.Name)
	d.Category = strings.TrimSpace(req.Category)
	d.Address = strings.TrimSpace(req.Address)
	d.Latitude = req.Latitude
	d.Longitude = req.Longitude
	d.SortOrder = req.SortOrder
	if req.IsActive != nil {
		d.IsActive = *req.IsActive
	}
	if err := db.Save(&d).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan destinasi"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Destinasi diperbarui", "destination": d})
}

// DELETE /api/admin/destinations/:id
func DeleteDestination(c *gin.Context) {
	if err := database.GetDB().Delete(&models.Destination{}, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus destinasi"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Destinasi dihapus"})
}

// ===================== Tarif booking per km =====================

type BookingRateRequest struct {
	VehicleType   string   `json:"vehicle_type" binding:"required"`
	Label         string   `json:"label" binding:"required"`
	BaseFare      *float64 `json:"base_fare" binding:"required"`
	PerKm         *float64 `json:"per_km" binding:"required"`
	MinFare       *float64 `json:"min_fare" binding:"required"`
	MaxDistanceKm *float64 `json:"max_distance_km" binding:"required"`
	Rounding      *float64 `json:"rounding"`
	IsActive      *bool    `json:"is_active"`
}

var allowedVehicleTypes = map[string]bool{
	string(models.VehicleTypeBecakManual): true, string(models.VehicleTypeBecakMotor): true,
	string(models.VehicleTypeBecakListrik): true, string(models.VehicleTypeAndong): true,
}

func applyRateRequest(c *gin.Context, rate *models.BookingRate, req BookingRateRequest) bool {
	if !allowedVehicleTypes[req.VehicleType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Jenis kendaraan tidak dikenal"})
		return false
	}
	if *req.BaseFare < 0 || *req.PerKm < 0 || *req.MinFare < 0 || *req.MaxDistanceKm <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nilai tarif tidak valid"})
		return false
	}
	rate.VehicleType = req.VehicleType
	rate.Label = strings.TrimSpace(req.Label)
	rate.BaseFare = *req.BaseFare
	rate.PerKm = *req.PerKm
	rate.MinFare = *req.MinFare
	rate.MaxDistanceKm = *req.MaxDistanceKm
	if req.Rounding != nil && *req.Rounding >= 0 {
		rate.Rounding = *req.Rounding
	} else if rate.Rounding == 0 {
		rate.Rounding = 1000
	}
	if req.IsActive != nil {
		rate.IsActive = *req.IsActive
	}
	return true
}

// GET /api/admin/booking-rates
func GetBookingRates(c *gin.Context) {
	var rates []models.BookingRate
	database.GetDB().Order("id ASC").Find(&rates)
	c.JSON(http.StatusOK, gin.H{"rates": rates})
}

// POST /api/admin/booking-rates
func CreateBookingRate(c *gin.Context) {
	var req BookingRateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Semua nilai tarif wajib diisi"})
		return
	}
	rate := models.BookingRate{IsActive: true}
	if !applyRateRequest(c, &rate, req) {
		return
	}
	db := database.GetDB()
	var exists int64
	db.Model(&models.BookingRate{}).Where("vehicle_type = ?", rate.VehicleType).Count(&exists)
	if exists > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Tarif untuk jenis kendaraan ini sudah ada"})
		return
	}
	if err := db.Create(&rate).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan tarif"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Tarif booking ditambahkan", "rate": rate})
}

// PUT /api/admin/booking-rates/:id
func UpdateBookingRate(c *gin.Context) {
	db := database.GetDB()
	var rate models.BookingRate
	if err := db.First(&rate, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tarif tidak ditemukan"})
		return
	}
	var req BookingRateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Semua nilai tarif wajib diisi"})
		return
	}
	if req.VehicleType != rate.VehicleType {
		var exists int64
		db.Model(&models.BookingRate{}).Where("vehicle_type = ? AND id <> ?", req.VehicleType, rate.ID).Count(&exists)
		if exists > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Tarif untuk jenis kendaraan ini sudah ada"})
			return
		}
	}
	if !applyRateRequest(c, &rate, req) {
		return
	}
	if err := db.Save(&rate).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan tarif"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Tarif booking diperbarui", "rate": rate})
}
