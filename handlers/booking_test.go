package handlers

import (
	"math"
	"testing"

	"greenbecak-backend/models"
	"greenbecak-backend/services"
)

func TestCalculateBookingPrice(t *testing.T) {
	becak := models.BookingRate{BaseFare: 5000, PerKm: 4000, MinFare: 10000, Rounding: 1000}

	cases := []struct {
		km   float64
		want float64
	}{
		{0.5, 10000}, // 5000 + 2000 = 7000 → minimum 10000
		{1.2, 10000}, // 5000 + 4800 = 9800 → minimum 10000
		{2.4, 15000}, // 5000 + 9600 = 14600 → dibulatkan 15000
		{3.0, 17000}, // 5000 + 12000 = 17000 (pas)
		{3.1, 18000}, // 17400 → 18000
	}
	for _, c := range cases {
		if got := CalculateBookingPrice(becak, c.km); got != c.want {
			t.Errorf("km=%.1f: got %.0f, want %.0f", c.km, got, c.want)
		}
	}

	noRounding := models.BookingRate{BaseFare: 5000, PerKm: 4000, MinFare: 0, Rounding: 0}
	if got := CalculateBookingPrice(noRounding, 2.4); math.Abs(got-14600) > 0.001 {
		t.Errorf("tanpa pembulatan: got %.2f, want 14600", got)
	}
}

func TestNormalizePhone(t *testing.T) {
	ok := map[string]string{
		"081234567890":     "081234567890",
		"+62 812-3456-7890": "081234567890",
		"6281234567890":    "081234567890",
		"81234567890":      "081234567890",
	}
	for in, want := range ok {
		got, valid := normalizePhone(in)
		if !valid || got != want {
			t.Errorf("normalizePhone(%q) = %q,%v want %q", in, got, valid, want)
		}
	}
	for _, bad := range []string{"", "12345", "0212345678", "08123", "0812345678901234"} {
		if _, valid := normalizePhone(bad); valid {
			t.Errorf("normalizePhone(%q) seharusnya tidak valid", bad)
		}
	}
}

func TestSanitizeSource(t *testing.T) {
	cases := map[string]string{
		"qr-statis":        "qr_statis",
		" Instagram ":      "instagram",
		"qr_hotel":         "qr_hotel",
		"<script>x</script>": "scriptxscript",
	}
	for in, want := range cases {
		if got := sanitizeSource(in); got != want {
			t.Errorf("sanitizeSource(%q) = %q want %q", in, got, want)
		}
	}
}

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"Grand Inna Malioboro": "grand-inna-malioboro",
		"  Hotel  #1 (Jogja) ": "hotel-1-jogja",
	}
	for in, want := range cases {
		if got := slugify(in); got != want {
			t.Errorf("slugify(%q) = %q want %q", in, got, want)
		}
	}
}

func TestFormatRupiah(t *testing.T) {
	cases := map[float64]string{0: "0", 999: "999", 15000: "15.000", 1250000: "1.250.000"}
	for in, want := range cases {
		if got := formatRupiah(in); got != want {
			t.Errorf("formatRupiah(%v) = %q want %q", in, got, want)
		}
	}
}

func TestHaversineAndServiceArea(t *testing.T) {
	// Tugu → Keraton ±2,5 km garis lurus
	d := services.HaversineKm(-7.78284, 110.36703, -7.80529, 110.36421)
	if d < 2.3 || d > 2.7 {
		t.Errorf("jarak Tugu-Keraton = %.2f km, di luar perkiraan", d)
	}
	area := services.ServiceArea()
	if !area.Contains(-7.79254, 110.36580) {
		t.Error("Malioboro seharusnya di dalam area layanan")
	}
	if area.Contains(-6.2, 106.8) {
		t.Error("Jakarta seharusnya di luar area layanan")
	}
}
