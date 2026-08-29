package indicators

import "fmt"

// SMA calculates Simple Moving Average
// Returns NaN for periods where there's insufficient data
func SMA(data []float64, period int) ([]float64, error) {
	if period <= 0 {
		return nil, fmt.Errorf("period must be positive, got %d", period)
	}
	if len(data) == 0 {
		return []float64{}, nil
	}

	result := make([]float64, len(data))

	// Fill initial values with NaN
	for i := 0; i < period-1; i++ {
		result[i] = 0 // Will be treated as "no signal" in strategies
	}

	// Calculate SMA for each valid window
	for i := period - 1; i < len(data); i++ {
		sum := 0.0
		for j := 0; j < period; j++ {
			sum += data[i-j]
		}
		result[i] = sum / float64(period)
	}

	return result, nil
}

// SMALast calculates SMA for only the most recent period
// More efficient when you only need the latest value
func SMALast(data []float64, period int) (float64, error) {
	if period <= 0 {
		return 0, fmt.Errorf("period must be positive, got %d", period)
	}
	if len(data) < period {
		return 0, fmt.Errorf("insufficient data: need %d, have %d", period, len(data))
	}

	sum := 0.0
	for i := len(data) - period; i < len(data); i++ {
		sum += data[i]
	}
	return sum / float64(period), nil
}
