package risk

import (
	"math"
	"testing"
)

func TestFixedFractional(t *testing.T) {
	sizer, err := NewFixedFractional(0.02)
	if err != nil {
		t.Fatalf("NewFixedFractional() error = %v", err)
	}
	size, err := sizer.CalculateSize(10000, 100, 98)
	if err != nil {
		t.Fatalf("CalculateSize() error = %v", err)
	}
	expected := 100.0
	if math.Abs(size-expected) > 0.01 {
		t.Errorf("CalculateSize() = %v, want %v", size, expected)
	}
}

func TestKellyCriterion(t *testing.T) {
	kelly, err := NewKellyCriterion(0.6, 100, 50, 0.25, 0.2)
	if err != nil {
		t.Fatalf("NewKellyCriterion() error = %v", err)
	}
	kellyPercent := kelly.CalculateKellyPercent()
	expected := 0.1
	if math.Abs(kellyPercent-expected) > 0.01 {
		t.Errorf("CalculateKellyPercent() = %v, want %v", kellyPercent, expected)
	}
}

func TestPercentStopLoss(t *testing.T) {
	stop, err := PercentStopLoss(100, 0.05, "long")
	if err != nil {
		t.Fatalf("PercentStopLoss() error = %v", err)
	}
	if stop.Price != 95 {
		t.Errorf("PercentStopLoss() price = %v, want 95", stop.Price)
	}
}

func TestTrailingStop(t *testing.T) {
	trail, err := NewTrailingStop(100, 5, "long")
	if err != nil {
		t.Fatalf("NewTrailingStop() error = %v", err)
	}
	if trail.CurrentPrice != 95 {
		t.Errorf("Initial stop = %v, want 95", trail.CurrentPrice)
	}
	trail.Update(110)
	if trail.CurrentPrice != 105 {
		t.Errorf("After 110, stop = %v, want 105", trail.CurrentPrice)
	}
}
