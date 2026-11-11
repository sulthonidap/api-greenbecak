package pricing

import (
	"testing"
	"time"

	"greenbecak-backend/models"
)

func setting() models.PricingSetting {
	s := models.DefaultPricingSetting()
	s.Enabled = true
	return s
}

func at(hhmm string) time.Time {
	t, _ := time.ParseInLocation("2006-01-02 15:04", "2026-11-11 "+hhmm, wib)
	return t
}

func TestCalculateBasic(t *testing.T) {
	// 5000 + 3.2 × 4000 = 17800 → dibulatkan 18000
	b := Calculate(setting(), 3.2, at("10:00"))
	if b.Total != 18000 {
		t.Fatalf("total = %v, want 18000", b.Total)
	}
	if b.MinFareApplied || b.IsNight {
		t.Fatalf("unexpected flags: %+v", b)
	}
}

func TestCalculateMinFare(t *testing.T) {
	// 5000 + 0.5 × 4000 = 7000 < 10000
	b := Calculate(setting(), 0.5, at("10:00"))
	if b.Total != 10000 || !b.MinFareApplied {
		t.Fatalf("got %+v", b)
	}
}

func TestCalculateExactMultipleNotBumped(t *testing.T) {
	// 5000 + 2.5 × 4000 = 15000, tidak boleh jadi 16000
	b := Calculate(setting(), 2.5, at("10:00"))
	if b.Total != 15000 {
		t.Fatalf("total = %v, want 15000", b.Total)
	}
}

func TestCalculateNightSurcharge(t *testing.T) {
	s := setting()
	s.NightSurchargePercent = 20
	// (5000 + 4×4000) × 1.2 = 25200 → 26000
	b := Calculate(s, 4, at("23:30"))
	if !b.IsNight || b.Total != 26000 {
		t.Fatalf("got %+v", b)
	}
	day := Calculate(s, 4, at("12:00"))
	if day.IsNight || day.Total != 21000 {
		t.Fatalf("day got %+v", day)
	}
}

func TestIsNightWrapsMidnight(t *testing.T) {
	s := setting()
	cases := map[string]bool{"21:59": false, "22:00": true, "02:00": true, "04:59": true, "05:00": false}
	for hhmm, want := range cases {
		if got := IsNight(s, at(hhmm)); got != want {
			t.Errorf("IsNight(%s) = %v, want %v", hhmm, got, want)
		}
	}
	// Konversi dari UTC: 16:00 UTC = 23:00 WIB
	utc := time.Date(2026, 11, 11, 16, 0, 0, 0, time.UTC)
	if !IsNight(s, utc) {
		t.Error("16:00 UTC should be night in WIB")
	}
}

func TestRoundUpNoStep(t *testing.T) {
	if RoundUp(17800.4, 0) != 17801 {
		t.Fatal("RoundUp without step should ceil to rupiah")
	}
}

func TestValidateSetting(t *testing.T) {
	if err := ValidateSetting(setting()); err != nil {
		t.Fatalf("default should be valid: %v", err)
	}
	bad := setting()
	bad.MaxDistanceKm = 0
	if ValidateSetting(bad) == nil {
		t.Error("max distance 0 should be invalid")
	}
	bad = setting()
	bad.NightStart = "25:00"
	if ValidateSetting(bad) == nil {
		t.Error("invalid night start should fail")
	}
}

func TestHaversine(t *testing.T) {
	// Tugu Jogja → Kraton kira-kira 2,4 km garis lurus
	km := HaversineKm(-7.7829, 110.3671, -7.8053, 110.3644)
	if km < 2.3 || km > 2.6 {
		t.Fatalf("haversine = %v", km)
	}
}
