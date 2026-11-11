package handlers

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"greenbecak-backend/database"
	"greenbecak-backend/models"
	"greenbecak-backend/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// currentDriver mengambil data driver dari user yang login. Menulis respons 404 kalau tidak ada.
func currentDriver(c *gin.Context, db *gorm.DB) (models.Driver, bool) {
	userID, _ := c.Get("user_id")
	var driver models.Driver
	if err := db.Where("user_id = ?", userID).First(&driver).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Driver not found"})
		return driver, false
	}
	return driver, true
}

// GET /api/driver/bookings/available?radius=3
// Booking yang menunggu driver, sesuai jenis kendaraan driver, diurutkan dari yang terdekat.
// Kalau lokasi driver belum diketahui, semua booking dikembalikan (location_known=false).
func GetAvailableBookings(c *gin.Context) {
	db := database.GetDB()
	driver, ok := currentDriver(c, db)
	if !ok {
		return
	}

	radius := bookingDriverRadiusKm()
	if r, err := strconv.ParseFloat(c.Query("radius"), 64); err == nil && r > 0 && r <= 15 {
		radius = r
	}

	var orders []models.Order
	db.Where("order_type = ? AND status = ? AND driver_id IS NULL AND (expires_at IS NULL OR expires_at > ?)",
		models.OrderTypeBooking, models.OrderStatusPending, time.Now()).
		Where("vehicle_type = ? OR vehicle_type = ''", driver.VehicleType).
		// Booking yang memakai kode becak tertentu tidak disiarkan ke driver lain
		Where("becak_code = '' OR becak_code IS NULL").
		Order("created_at ASC").
		Find(&orders)

	loc, locationKnown := freshDriverLocation(db, driver.ID)

	type item struct {
		ID                 uint       `json:"id"`
		OrderNumber        string     `json:"order_number"`
		PickupLocation     string     `json:"pickup_location"`
		PickupLat          *float64   `json:"pickup_lat"`
		PickupLng          *float64   `json:"pickup_lng"`
		DropLocation       string     `json:"drop_location"`
		DropLat            *float64   `json:"drop_lat"`
		DropLng            *float64   `json:"drop_lng"`
		Distance           float64    `json:"distance"`
		Price              float64    `json:"price"`
		VehicleType        string     `json:"vehicle_type"`
		CustomerName       string     `json:"customer_name"`
		Notes              string     `json:"notes"`
		DistanceToPickupKm *float64   `json:"distance_to_pickup_km"`
		CreatedAt          time.Time  `json:"created_at"`
		ExpiresAt          *time.Time `json:"expires_at"`
	}

	items := []item{}
	for _, o := range orders {
		it := item{
			ID: o.ID, OrderNumber: o.OrderNumber,
			PickupLocation: o.PickupLocation, PickupLat: o.PickupLat, PickupLng: o.PickupLng,
			DropLocation: o.DropLocation, DropLat: o.DropLat, DropLng: o.DropLng,
			Distance: o.Distance, Price: o.Price, VehicleType: o.VehicleType,
			CustomerName: o.CustomerName, Notes: o.Notes,
			CreatedAt: o.CreatedAt, ExpiresAt: o.ExpiresAt,
		}
		if locationKnown && o.PickupLat != nil && o.PickupLng != nil {
			d := round1(services.HaversineKm(loc.Latitude, loc.Longitude, *o.PickupLat, *o.PickupLng))
			if d > radius {
				continue
			}
			it.DistanceToPickupKm = &d
		}
		items = append(items, it)
	}

	if locationKnown {
		sort.Slice(items, func(i, j int) bool {
			return *items[i].DistanceToPickupKm < *items[j].DistanceToPickupKm
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"bookings":       items,
		"count":          len(items),
		"location_known": locationKnown,
		"radius_km":      radius,
		"busy":           driverBusy(db, driver.ID, 0),
	})
}

// PUT /api/driver/orders/:id/pickup — penumpang sudah dijemput (accepted → picked_up)
func PickupOrder(c *gin.Context) {
	db := database.GetDB()
	driver, ok := currentDriver(c, db)
	if !ok {
		return
	}

	now := time.Now()
	res := db.Model(&models.Order{}).
		Where("id = ? AND driver_id = ? AND status = ?", c.Param("id"), driver.ID, models.OrderStatusAccepted).
		Updates(map[string]interface{}{
			"status":       models.OrderStatusPickedUp,
			"picked_up_at": now,
		})
	if res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui order"})
		return
	}
	if res.RowsAffected == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order tidak ditemukan atau belum diterima"})
		return
	}

	var order models.Order
	db.First(&order, c.Param("id"))
	c.JSON(http.StatusOK, gin.H{"message": "Penumpang sudah dijemput", "order": order})
}

// PUT /api/driver/orders/:id/release — driver batal mengambil booking (sebelum menjemput).
// Booking kembali ke status menunggu driver dan disiarkan ulang.
func ReleaseBooking(c *gin.Context) {
	db := database.GetDB()
	driver, ok := currentDriver(c, db)
	if !ok {
		return
	}

	var current models.Order
	if err := db.First(&current, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	// Booking dengan kode becak dipilih penumpang khusus untuk driver ini:
	// kalau driver tidak bisa, booking dibatalkan (tidak disiarkan ke driver lain).
	if current.OrderType == models.OrderTypeBooking && current.BecakCode != "" {
		now := time.Now()
		res := db.Model(&models.Order{}).
			Where("id = ? AND driver_id = ? AND status IN ?", current.ID, driver.ID,
				[]models.OrderStatus{models.OrderStatusPending, models.OrderStatusAccepted}).
			Updates(map[string]interface{}{
				"status":        models.OrderStatusCancelled,
				"cancelled_at":  now,
				"cancel_reason": "Driver tidak dapat mengambil pesanan ini",
			})
		if res.Error != nil || res.RowsAffected == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Booking tidak bisa dilepas"})
			return
		}
		db.Model(&driver).Update("status", models.DriverStatusActive)
		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Booking %s dibatalkan", current.OrderNumber)})
		return
	}

	expires := time.Now().Add(time.Duration(bookingExpiryMinutes()) * time.Minute)
	res := db.Model(&models.Order{}).
		Where("id = ? AND driver_id = ? AND status = ? AND order_type = ?",
			c.Param("id"), driver.ID, models.OrderStatusAccepted, models.OrderTypeBooking).
		Updates(map[string]interface{}{
			"status":      models.OrderStatusPending,
			"driver_id":   nil,
			"accepted_at": nil,
			"expires_at":  expires,
		})
	if res.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal melepas booking"})
		return
	}
	if res.RowsAffected == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Booking tidak bisa dilepas"})
		return
	}

	db.Model(&driver).Update("status", models.DriverStatusActive)

	var order models.Order
	if err := db.First(&order, c.Param("id")).Error; err == nil {
		go notifyDriversForBooking(order)
	}
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("Booking %s dilepas dan dicarikan driver lain", order.OrderNumber)})
}
