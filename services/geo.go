package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ===== Konfigurasi =====
//
// OSRM_URL           : server rute OSRM. Default profil SEPEDA dari routing.openstreetmap.de karena
//                      paling mirip becak (boleh lewat Malioboro, tidak memutar seperti mobil).
//                      Server publik untuk fair-use; untuk trafik besar sebaiknya self-host.
// OSRM_DURATION_FACTOR: pengali durasi OSRM (default 1.0)
// PHOTON_URL         : server pencarian alamat berbasis OpenStreetMap (default: photon.komoot.io)
// GEO_USER_AGENT     : User-Agent untuk request ke layanan peta (isi dengan nama aplikasi + kontak)
// SERVICE_AREA_BBOX  : "minLat,minLng,maxLat,maxLng" area layanan (default: DI Yogyakarta & sekitarnya)

func envOr(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return strings.TrimRight(v, "/")
	}
	return def
}

var httpClient = &http.Client{Timeout: 4 * time.Second}

func geoUserAgent() string {
	return envOr("GEO_USER_AGENT", "becak-booking-api/1.0")
}

// ===== Area layanan =====

type BBox struct {
	MinLat float64 `json:"min_lat"`
	MinLng float64 `json:"min_lng"`
	MaxLat float64 `json:"max_lat"`
	MaxLng float64 `json:"max_lng"`
}

// Default: kotak yang mencakup Kota Yogyakarta, Sleman, dan Bantul bagian utara.
var defaultServiceArea = BBox{MinLat: -8.10, MinLng: 110.25, MaxLat: -7.65, MaxLng: 110.50}

func ServiceArea() BBox {
	raw := os.Getenv("SERVICE_AREA_BBOX")
	if raw == "" {
		return defaultServiceArea
	}
	parts := strings.Split(raw, ",")
	if len(parts) != 4 {
		return defaultServiceArea
	}
	vals := make([]float64, 4)
	for i, p := range parts {
		v, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
		if err != nil {
			return defaultServiceArea
		}
		vals[i] = v
	}
	return BBox{MinLat: vals[0], MinLng: vals[1], MaxLat: vals[2], MaxLng: vals[3]}
}

func (b BBox) Contains(lat, lng float64) bool {
	return lat >= b.MinLat && lat <= b.MaxLat && lng >= b.MinLng && lng <= b.MaxLng
}

// ===== Jarak =====

// HaversineKm menghitung jarak garis lurus (km) antara dua titik.
func HaversineKm(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*math.Sin(dLng/2)*math.Sin(dLng/2)
	return R * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

// Faktor pengali garis lurus → perkiraan jarak jalan (dipakai kalau layanan rute gagal).
const straightLineFactor = 1.3

type RouteResult struct {
	DistanceKm  float64 `json:"distance_km"`
	DurationMin float64 `json:"duration_min"`
	Source      string  `json:"source"` // "route" atau "estimate"
}

var routeCache = newTTLCache(30 * time.Minute)

// RouteDistance menghitung jarak via rute jalan (OSRM). Kalau gagal, pakai garis lurus × 1,3.
func RouteDistance(lat1, lng1, lat2, lng2 float64) RouteResult {
	key := fmt.Sprintf("%.5f,%.5f;%.5f,%.5f", lat1, lng1, lat2, lng2)
	if v, ok := routeCache.get(key); ok {
		return v.(RouteResult)
	}

	if res, err := osrmRoute(lat1, lng1, lat2, lng2); err == nil {
		routeCache.set(key, res)
		return res
	}

	km := HaversineKm(lat1, lng1, lat2, lng2) * straightLineFactor
	// Asumsi kecepatan becak ±12 km/jam
	return RouteResult{DistanceKm: km, DurationMin: km / 12 * 60, Source: "estimate"}
}

func osrmRoute(lat1, lng1, lat2, lng2 float64) (RouteResult, error) {
	base := envOr("OSRM_URL", "https://routing.openstreetmap.de/routed-bike")
	u := fmt.Sprintf("%s/route/v1/driving/%f,%f;%f,%f?overview=false", base, lng1, lat1, lng2, lat2)

	req, _ := http.NewRequest(http.MethodGet, u, nil)
	req.Header.Set("User-Agent", geoUserAgent())
	resp, err := httpClient.Do(req)
	if err != nil {
		return RouteResult{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return RouteResult{}, fmt.Errorf("osrm status %d", resp.StatusCode)
	}

	var body struct {
		Code   string `json:"code"`
		Routes []struct {
			Distance float64 `json:"distance"` // meter
			Duration float64 `json:"duration"` // detik
		} `json:"routes"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return RouteResult{}, err
	}
	if body.Code != "Ok" || len(body.Routes) == 0 {
		return RouteResult{}, errors.New("osrm: no route")
	}
	km := body.Routes[0].Distance / 1000
	factor := 1.0
	if v, err := strconv.ParseFloat(os.Getenv("OSRM_DURATION_FACTOR"), 64); err == nil && v > 0 {
		factor = v
	}
	return RouteResult{DistanceKm: km, DurationMin: body.Routes[0].Duration / 60 * factor, Source: "route"}, nil
}

// ===== Pencarian alamat (Photon / OpenStreetMap) =====

type Place struct {
	Name     string  `json:"name"`
	Address  string  `json:"address"`
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
	Type     string  `json:"type"`     // "destination" (dari admin), "place" (OSM)
	Category string  `json:"category"` // mis. hotel, restaurant, attraction
}

var (
	searchCache  = newTTLCache(15 * time.Minute)
	reverseCache = newTTLCache(60 * time.Minute)
	// Batasi request keluar supaya tidak melanggar fair-use server publik
	geoThrottle = newThrottle(300 * time.Millisecond)
)

type photonResponse struct {
	Features []struct {
		Geometry struct {
			Coordinates []float64 `json:"coordinates"` // [lng, lat]
		} `json:"geometry"`
		Properties map[string]interface{} `json:"properties"`
	} `json:"features"`
}

// SearchPlaces mencari tempat berdasarkan teks, dibatasi area layanan.
// biasLat/biasLng (opsional, 0 = tidak ada) memprioritaskan hasil dekat posisi penumpang.
func SearchPlaces(query string, biasLat, biasLng float64, limit int) ([]Place, error) {
	query = strings.TrimSpace(query)
	if len([]rune(query)) < 3 {
		return []Place{}, nil
	}
	area := ServiceArea()
	key := fmt.Sprintf("%s|%.3f|%.3f|%d", strings.ToLower(query), biasLat, biasLng, limit)
	if v, ok := searchCache.get(key); ok {
		return v.([]Place), nil
	}

	params := url.Values{}
	params.Set("q", query)
	params.Set("limit", strconv.Itoa(limit))
	params.Set("bbox", fmt.Sprintf("%f,%f,%f,%f", area.MinLng, area.MinLat, area.MaxLng, area.MaxLat))
	if biasLat != 0 && biasLng != 0 {
		params.Set("lat", fmt.Sprintf("%f", biasLat))
		params.Set("lon", fmt.Sprintf("%f", biasLng))
	}

	body, err := photonGet("/api/?" + params.Encode())
	if err != nil {
		return nil, err
	}

	places := make([]Place, 0, len(body.Features))
	for _, f := range body.Features {
		if len(f.Geometry.Coordinates) < 2 {
			continue
		}
		lat, lng := f.Geometry.Coordinates[1], f.Geometry.Coordinates[0]
		if !area.Contains(lat, lng) {
			continue
		}
		name, address := describePhoton(f.Properties)
		if name == "" {
			continue
		}
		places = append(places, Place{
			Name:     name,
			Address:  address,
			Lat:      lat,
			Lng:      lng,
			Type:     "place",
			Category: propString(f.Properties, "osm_value"),
		})
	}
	searchCache.set(key, places)
	return places, nil
}

// ReverseGeocode mengubah koordinat (dari pin peta) menjadi nama tempat/jalan.
func ReverseGeocode(lat, lng float64) (Place, error) {
	key := fmt.Sprintf("%.5f,%.5f", lat, lng)
	if v, ok := reverseCache.get(key); ok {
		return v.(Place), nil
	}

	params := url.Values{}
	params.Set("lat", fmt.Sprintf("%f", lat))
	params.Set("lon", fmt.Sprintf("%f", lng))
	params.Set("limit", "1")

	body, err := photonGet("/reverse?" + params.Encode())
	if err != nil {
		return Place{}, err
	}

	place := Place{Lat: lat, Lng: lng, Type: "pin"}
	if len(body.Features) > 0 {
		place.Name, place.Address = describePhoton(body.Features[0].Properties)
		place.Category = propString(body.Features[0].Properties, "osm_value")
	}
	if place.Name == "" {
		place.Name = fmt.Sprintf("Titik %.5f, %.5f", lat, lng)
	}
	reverseCache.set(key, place)
	return place, nil
}

func photonGet(pathAndQuery string) (*photonResponse, error) {
	geoThrottle.wait()

	base := envOr("PHOTON_URL", "https://photon.komoot.io")
	req, _ := http.NewRequest(http.MethodGet, base+pathAndQuery, nil)
	req.Header.Set("User-Agent", geoUserAgent())
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("photon status %d", resp.StatusCode)
	}
	var body photonResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	return &body, nil
}

func propString(p map[string]interface{}, key string) string {
	if v, ok := p[key].(string); ok {
		return v
	}
	return ""
}

// describePhoton membentuk nama & alamat yang enak dibaca dari properti Photon.
func describePhoton(p map[string]interface{}) (name, address string) {
	name = propString(p, "name")
	street := propString(p, "street")
	if hn := propString(p, "housenumber"); hn != "" && street != "" {
		street = street + " " + hn
	}
	if name == "" {
		name = street
		street = ""
	}

	parts := []string{}
	// "county" sengaja tidak dipakai: data OSM di Jogja sering memberi nilai yang membingungkan (mis. "Depok")
	for _, v := range []string{street, propString(p, "locality"), propString(p, "district"), propString(p, "city")} {
		if v != "" && v != name && !containsStr(parts, v) {
			parts = append(parts, v)
		}
	}
	return name, strings.Join(parts, ", ")
}

func containsStr(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// ===== Util: cache & throttle =====

type ttlEntry struct {
	value   interface{}
	expires time.Time
}

type ttlCache struct {
	mu   sync.Mutex
	ttl  time.Duration
	data map[string]ttlEntry
}

const maxCacheEntries = 5000

func newTTLCache(ttl time.Duration) *ttlCache {
	return &ttlCache{ttl: ttl, data: make(map[string]ttlEntry)}
}

func (c *ttlCache) get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.data[key]
	if !ok || time.Now().After(e.expires) {
		delete(c.data, key)
		return nil, false
	}
	return e.value, true
}

func (c *ttlCache) set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.data) >= maxCacheEntries {
		// Sederhana: kosongkan saat penuh
		c.data = make(map[string]ttlEntry)
	}
	c.data[key] = ttlEntry{value: value, expires: time.Now().Add(c.ttl)}
}

type throttle struct {
	mu       sync.Mutex
	interval time.Duration
	last     time.Time
}

func newThrottle(interval time.Duration) *throttle {
	return &throttle{interval: interval}
}

func (t *throttle) wait() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if d := t.interval - time.Since(t.last); d > 0 {
		time.Sleep(d)
	}
	t.last = time.Now()
}
