// Package pricing menghitung harga order jemput berdasarkan jarak rute.
package pricing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"greenbecak-backend/models"
)

// Breakdown adalah rincian harga yang dikirim ke customer.
type Breakdown struct {
	BaseFare        float64 `json:"base_fare"`
	DistanceFare    float64 `json:"distance_fare"`
	NightSurcharge  float64 `json:"night_surcharge"`
	Subtotal        float64 `json:"subtotal"`
	MinFareApplied  bool    `json:"min_fare_applied"`
	Total           float64 `json:"total"`
	IsNight         bool    `json:"is_night"`
	DistanceKm      float64 `json:"distance_km"`
	PerKm           float64 `json:"per_km"`
	NightPercentage float64 `json:"night_percentage"`
}

// Calculate menghitung harga dari jarak (km) dan waktu pemesanan.
// harga = tarif_dasar + jarak × per_km (+ tambahan malam), minimal tarif_minimum, dibulatkan ke atas.
func Calculate(s models.PricingSetting, distanceKm float64, at time.Time) Breakdown {
	km := math.Round(distanceKm*100) / 100
	b := Breakdown{
		BaseFare:     s.BaseFare,
		DistanceFare: km * s.PerKm,
		DistanceKm:   km,
		PerKm:        s.PerKm,
	}
	b.Subtotal = b.BaseFare + b.DistanceFare

	if s.NightSurchargePercent > 0 && IsNight(s, at) {
		b.IsNight = true
		b.NightPercentage = s.NightSurchargePercent
		b.NightSurcharge = b.Subtotal * s.NightSurchargePercent / 100
		b.Subtotal += b.NightSurcharge
	}

	total := b.Subtotal
	if total < s.MinFare {
		total = s.MinFare
		b.MinFareApplied = true
	}
	b.Total = RoundUp(total, s.RoundTo)
	return b
}

// RoundUp membulatkan ke atas ke kelipatan step (step <= 0 berarti hanya dibulatkan ke rupiah).
func RoundUp(v, step float64) float64 {
	if step <= 0 {
		return math.Ceil(v)
	}
	// Toleransi kecil supaya 18000.0000001 tidak jadi 19000
	return math.Ceil(v/step-1e-9) * step
}

var wib = func() *time.Location {
	if loc, err := time.LoadLocation("Asia/Jakarta"); err == nil {
		return loc
	}
	return time.FixedZone("WIB", 7*3600)
}()

// IsNight mengecek apakah waktu `at` (dikonversi ke WIB) masuk jam tarif malam.
// Mendukung rentang yang melewati tengah malam, mis. 22:00–05:00.
func IsNight(s models.PricingSetting, at time.Time) bool {
	start, err1 := parseHHMM(s.NightStart)
	end, err2 := parseHHMM(s.NightEnd)
	if err1 != nil || err2 != nil || start == end {
		return false
	}
	t := at.In(wib)
	now := t.Hour()*60 + t.Minute()
	if start < end {
		return now >= start && now < end
	}
	return now >= start || now < end
}

func parseHHMM(v string) (int, error) {
	parts := strings.Split(strings.TrimSpace(v), ":")
	if len(parts) != 2 {
		return 0, errors.New("format jam harus HH:MM")
	}
	h, err := strconv.Atoi(parts[0])
	if err != nil || h < 0 || h > 23 {
		return 0, errors.New("jam tidak valid")
	}
	m, err := strconv.Atoi(parts[1])
	if err != nil || m < 0 || m > 59 {
		return 0, errors.New("menit tidak valid")
	}
	return h*60 + m, nil
}

// ValidateSetting memeriksa nilai pengaturan dari admin.
func ValidateSetting(s models.PricingSetting) error {
	switch {
	case s.BaseFare < 0 || s.PerKm < 0 || s.MinFare < 0:
		return errors.New("tarif tidak boleh negatif")
	case s.PerKm == 0 && s.BaseFare == 0 && s.MinFare == 0:
		return errors.New("minimal salah satu tarif harus diisi")
	case s.MaxDistanceKm <= 0:
		return errors.New("jarak maksimal harus lebih dari 0 km")
	case s.RoundTo < 0:
		return errors.New("pembulatan tidak boleh negatif")
	case s.NightSurchargePercent < 0 || s.NightSurchargePercent > 200:
		return errors.New("tambahan tarif malam harus 0–200%")
	case s.ServiceRadiusKm < 0:
		return errors.New("radius area layanan tidak boleh negatif")
	case !ValidLatLng(s.ServiceCenterLat, s.ServiceCenterLng):
		return errors.New("titik pusat area layanan tidak valid")
	}
	if _, err := parseHHMM(s.NightStart); err != nil {
		return fmt.Errorf("jam mulai tarif malam: %w", err)
	}
	if _, err := parseHHMM(s.NightEnd); err != nil {
		return fmt.Errorf("jam selesai tarif malam: %w", err)
	}
	return nil
}

// ValidLatLng memeriksa rentang koordinat.
func ValidLatLng(lat, lng float64) bool {
	return lat >= -90 && lat <= 90 && lng >= -180 && lng <= 180 && !(lat == 0 && lng == 0)
}

// HaversineKm menghitung jarak garis lurus dua titik (km).
func HaversineKm(lat1, lng1, lat2, lng2 float64) float64 {
	const r = 6371.0
	toRad := func(d float64) float64 { return d * math.Pi / 180 }
	dLat := toRad(lat2 - lat1)
	dLng := toRad(lng2 - lng1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRad(lat1))*math.Cos(toRad(lat2))*math.Sin(dLng/2)*math.Sin(dLng/2)
	return 2 * r * math.Asin(math.Sqrt(a))
}

// Route adalah hasil perhitungan rute.
type Route struct {
	DistanceKm  float64      `json:"distance_km"`
	DurationMin int          `json:"duration_min"`
	Geometry    [][2]float64 `json:"geometry"`  // [lat, lng]
	Estimated   bool         `json:"estimated"` // true kalau OSRM gagal dan jarak diperkirakan
}

// Faktor pengali jarak garis lurus kalau layanan rute tidak bisa dihubungi.
const fallbackRoadFactor = 1.3

// Kecepatan rata-rata becak (km/jam) untuk estimasi durasi.
const becakSpeedKmh = 12.0

var httpClient = &http.Client{Timeout: 6 * time.Second}

func osrmBaseURL() string {
	if v := strings.TrimRight(os.Getenv("OSRM_URL"), "/"); v != "" {
		return v
	}
	return "https://router.project-osrm.org"
}

// GetRoute mengambil rute jalan dari OSRM. Kalau gagal, pakai perkiraan jarak garis lurus × 1,3.
func GetRoute(ctx context.Context, pLat, pLng, dLat, dLng float64) Route {
	if r, err := osrmRoute(ctx, pLat, pLng, dLat, dLng); err == nil {
		return r
	}
	km := HaversineKm(pLat, pLng, dLat, dLng) * fallbackRoadFactor
	return Route{
		DistanceKm:  math.Round(km*100) / 100,
		DurationMin: durationFor(km),
		Geometry:    [][2]float64{{pLat, pLng}, {dLat, dLng}},
		Estimated:   true,
	}
}

func durationFor(km float64) int {
	m := int(math.Ceil(km / becakSpeedKmh * 60))
	if m < 1 {
		m = 1
	}
	return m
}

func osrmRoute(ctx context.Context, pLat, pLng, dLat, dLng float64) (Route, error) {
	// OSRM memakai urutan lng,lat
	url := fmt.Sprintf("%s/route/v1/driving/%f,%f;%f,%f?overview=simplified&geometries=geojson",
		osrmBaseURL(), pLng, pLat, dLng, dLat)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Route{}, err
	}
	req.Header.Set("User-Agent", "GreenBecak/1.0")
	resp, err := httpClient.Do(req)
	if err != nil {
		return Route{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Route{}, fmt.Errorf("osrm status %d", resp.StatusCode)
	}

	var body struct {
		Code   string `json:"code"`
		Routes []struct {
			Distance float64 `json:"distance"` // meter
			Geometry struct {
				Coordinates [][]float64 `json:"coordinates"` // [lng, lat]
			} `json:"geometry"`
		} `json:"routes"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return Route{}, err
	}
	if body.Code != "Ok" || len(body.Routes) == 0 {
		return Route{}, fmt.Errorf("osrm: %s", body.Code)
	}

	r := body.Routes[0]
	km := r.Distance / 1000
	geom := make([][2]float64, 0, len(r.Geometry.Coordinates))
	for _, c := range r.Geometry.Coordinates {
		if len(c) >= 2 {
			geom = append(geom, [2]float64{c[1], c[0]})
		}
	}
	return Route{
		DistanceKm: math.Round(km*100) / 100,
		// Durasi dari OSRM adalah untuk mobil, jadi dihitung ulang dengan kecepatan becak
		DurationMin: durationFor(km),
		Geometry:    geom,
	}, nil
}
