package indicators

import (
	"fmt"
	"math"
)

// BollingerBands result contains upper, middle, and lower bands
type BollingerBands struct {
	Upper  []float64
	Middle []float64
	Lower  []float64
}

// Bollinger calculates Bollinger Bands
// Middle Band = SMA(period)
// Upper Band = Middle + (stdDev * multiplier)
// Lower Band = Middle - (stdDev * multiplier)
// Standard: period=20, multiplier=2
func Bollinger(data []float64, period int, multiplier float64) (*BollingerBands, error) {
	if period <= 0 {
		return nil, fmt.Errorf("period must be positive, got %d", period)
	}
	if multiplier <= 0 {
		return nil, fmt.Errorf("multiplier must be positive, got %f", multiplier)
	}
	if len(data) < period {
		return &BollingerBands{
			Upper:  make([]float64, len(data)),
			Middle: make([]float64, len(data)),
			Lower:  make([]float64, len(data)),
		}, nil
	}

	// Calculate middle band (SMA)
	middle, err := SMA(data, period)
	if err != nil {
		return nil, err
	}

	// Calculate standard deviation for each window
	upper := make([]float64, len(data))
	lower := make([]float64, len(data))

	for i := period - 1; i < len(data); i++ {
		// Calculate standard deviation for window
		sum := 0.0
		for j := 0; j < period; j++ {
			diff := data[i-j] - middle[i]
			sum += diff * diff
		}
		stdDev := math.Sqrt(sum / float64(period))

		upper[i] = middle[i] + (stdDev * multiplier)
		lower[i] = middle[i] - (stdDev * multiplier)
	}

	return &BollingerBands{
		Upper:  upper,
		Middle: middle,
		Lower:  lower,
	}, nil
}
