package indicators

import (
	"math"
	"testing"
)

func TestSMA(t *testing.T) {
	tests := []struct {
		name     string
		data     []float64
		period   int
		wantErr  bool
		wantLast float64
	}{
		{
			name:     "simple case",
			data:     []float64{1, 2, 3, 4, 5},
			period:   3,
			wantErr:  false,
			wantLast: 4.0, // (3+4+5)/3
		},
		{
			name:     "period equals length",
			data:     []float64{10, 20, 30},
			period:   3,
			wantErr:  false,
			wantLast: 20.0,
		},
		{
			name:    "invalid period",
			data:    []float64{1, 2, 3},
			period:  0,
			wantErr: true,
		},
		{
			name:     "empty data",
			data:     []float64{},
			period:   5,
			wantErr:  false,
			wantLast: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := SMA(tt.data, tt.period)
			if (err != nil) != tt.wantErr {
				t.Errorf("SMA() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(result) > 0 {
				last := result[len(result)-1]
				if math.Abs(last-tt.wantLast) > 0.001 {
					t.Errorf("SMA() last value = %v, want %v", last, tt.wantLast)
				}
			}
		})
	}
}

func TestEMA(t *testing.T) {
	data := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	period := 5

	result, err := EMA(data, period)
	if err != nil {
		t.Fatalf("EMA() error = %v", err)
	}

	// First EMA value should be SMA of first 5 points
	expectedFirst := (1.0 + 2 + 3 + 4 + 5) / 5.0
	if math.Abs(result[4]-expectedFirst) > 0.001 {
		t.Errorf("EMA() first value = %v, want %v", result[4], expectedFirst)
	}

	// Last value should be non-zero
	if result[len(result)-1] == 0 {
		t.Error("EMA() last value should not be zero")
	}
}

func TestRSI(t *testing.T) {
	// Test with trending data
	data := []float64{44, 44.34, 44.09, 43.61, 44.33, 44.83, 45.10, 45.42, 45.84, 46.08, 45.89, 46.03, 45.61, 46.28, 46.28, 46.00, 46.03, 46.41, 46.22, 45.64}
	period := 14

	result, err := RSI(data, period)
	if err != nil {
		t.Fatalf("RSI() error = %v", err)
	}

	// RSI should be between 0 and 100
	for i, v := range result {
		if i < period {
			continue // Skip initial values
		}
		if v < 0 || v > 100 {
			t.Errorf("RSI() value at %d = %v, should be between 0 and 100", i, v)
		}
	}
}

func TestMACD(t *testing.T) {
	data := make([]float64, 100)
	for i := range data {
		data[i] = float64(i) + 50 // Trending data
	}

	result, err := MACD(data, 12, 26, 9)
	if err != nil {
		t.Fatalf("MACD() error = %v", err)
	}

	if len(result.MACD) != len(data) {
		t.Errorf("MACD() length = %d, want %d", len(result.MACD), len(data))
	}

	// Check histogram calculation
	for i := 26; i < len(data); i++ {
		expectedHist := result.MACD[i] - result.Signal[i]
		if math.Abs(result.Histogram[i]-expectedHist) > 0.001 {
			t.Errorf("MACD() histogram at %d = %v, want %v", i, result.Histogram[i], expectedHist)
			break
		}
	}
}

func TestATR(t *testing.T) {
	high := []float64{50, 52, 51, 53, 54}
	low := []float64{48, 49, 48, 50, 51}
	close := []float64{49, 51, 49, 52, 53}
	period := 3

	result, err := ATR(high, low, close, period)
	if err != nil {
		t.Fatalf("ATR() error = %v", err)
	}

	// ATR should be positive
	for i, v := range result {
		if i < period && v > 0 {
			continue
		}
		if v < 0 {
			t.Errorf("ATR() value at %d = %v, should be non-negative", i, v)
		}
	}
}

func TestATRMismatchedLengths(t *testing.T) {
	high := []float64{50, 52, 51}
	low := []float64{48, 49}
	close := []float64{49, 51, 49}

	_, err := ATR(high, low, close, 3)
	if err == nil {
		t.Error("ATR() should error on mismatched lengths")
	}
}

func TestBollinger(t *testing.T) {
	data := []float64{20, 21, 22, 23, 24, 25, 24, 23, 22, 21, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30}
	period := 20
	multiplier := 2.0

	result, err := Bollinger(data, period, multiplier)
	if err != nil {
		t.Fatalf("Bollinger() error = %v", err)
	}

	// Check that upper > middle > lower
	for i := period - 1; i < len(data); i++ {
		if result.Upper[i] <= result.Middle[i] {
			t.Errorf("Bollinger() at %d: upper (%v) should be > middle (%v)", i, result.Upper[i], result.Middle[i])
		}
		if result.Middle[i] <= result.Lower[i] {
			t.Errorf("Bollinger() at %d: middle (%v) should be > lower (%v)", i, result.Middle[i], result.Lower[i])
		}
	}
}

func BenchmarkSMA(b *testing.B) {
	data := make([]float64, 1000)
	for i := range data {
		data[i] = float64(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = SMA(data, 50)
	}
}

func BenchmarkEMA(b *testing.B) {
	data := make([]float64, 1000)
	for i := range data {
		data[i] = float64(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = EMA(data, 50)
	}
}

func BenchmarkRSI(b *testing.B) {
	data := make([]float64, 1000)
	for i := range data {
		data[i] = float64(i % 100)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = RSI(data, 14)
	}
}
