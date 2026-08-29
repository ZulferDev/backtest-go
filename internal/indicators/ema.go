package indicators

import "fmt"

// EMA calculates Exponential Moving Average
// Uses standard EMA formula: EMA = Price(t) * k + EMA(t-1) * (1-k)
// where k = 2 / (period + 1)
func EMA(data []float64, period int) ([]float64, error) {
	if period <= 0 {
		return nil, fmt.Errorf("period must be positive, got %d", period)
	}
	if len(data) == 0 {
		return []float64{}, nil
	}

	result := make([]float64, len(data))
	k := 2.0 / float64(period+1)

	// First EMA value is SMA of first 'period' points
	if len(data) < period {
		// Not enough data, return zeros
		return result, nil
	}

	// Calculate initial SMA
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += data[i]
	}
	result[period-1] = sum / float64(period)

	// Calculate EMA for remaining points
	for i := period; i < len(data); i++ {
		result[i] = data[i]*k + result[i-1]*(1-k)
	}

	return result, nil
}

// EMALast calculates EMA incrementally from previous EMA value
// Useful for real-time updates
func EMALast(price, prevEMA float64, period int) float64 {
	k := 2.0 / float64(period+1)
	return price*k + prevEMA*(1-k)
}
