package indicators

import "fmt"

// RSI calculates Relative Strength Index
// Formula: RSI = 100 - (100 / (1 + RS))
// where RS = Average Gain / Average Loss over period
func RSI(data []float64, period int) ([]float64, error) {
	if period <= 0 {
		return nil, fmt.Errorf("period must be positive, got %d", period)
	}
	if len(data) < period+1 {
		// Need at least period+1 points to calculate RSI
		return make([]float64, len(data)), nil
	}

	result := make([]float64, len(data))

	// Calculate price changes
	gains := make([]float64, len(data)-1)
	losses := make([]float64, len(data)-1)

	for i := 1; i < len(data); i++ {
		change := data[i] - data[i-1]
		if change > 0 {
			gains[i-1] = change
			losses[i-1] = 0
		} else {
			gains[i-1] = 0
			losses[i-1] = -change
		}
	}

	// Calculate initial average gain and loss (SMA)
	avgGain := 0.0
	avgLoss := 0.0
	for i := 0; i < period; i++ {
		avgGain += gains[i]
		avgLoss += losses[i]
	}
	avgGain /= float64(period)
	avgLoss /= float64(period)

	// Calculate RSI for first valid point
	if avgLoss == 0 {
		result[period] = 100
	} else {
		rs := avgGain / avgLoss
		result[period] = 100 - (100 / (1 + rs))
	}

	// Calculate RSI for remaining points using smoothed averages
	for i := period + 1; i < len(data); i++ {
		avgGain = (avgGain*float64(period-1) + gains[i-1]) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + losses[i-1]) / float64(period)

		if avgLoss == 0 {
			result[i] = 100
		} else {
			rs := avgGain / avgLoss
			result[i] = 100 - (100 / (1 + rs))
		}
	}

	return result, nil
}
