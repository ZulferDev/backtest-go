package indicators

import (
	"fmt"
	"math"
)

// ATR calculates Average True Range
// True Range = max(High - Low, |High - PrevClose|, |Low - PrevClose|)
// ATR = EMA of True Range
func ATR(high, low, close []float64, period int) ([]float64, error) {
	if period <= 0 {
		return nil, fmt.Errorf("period must be positive, got %d", period)
	}
	if len(high) != len(low) || len(high) != len(close) {
		return nil, fmt.Errorf("high, low, close must have same length")
	}
	if len(high) < 2 {
		return make([]float64, len(high)), nil
	}

	// Calculate True Range for each bar
	tr := make([]float64, len(high))
	tr[0] = high[0] - low[0] // First bar: just High - Low

	for i := 1; i < len(high); i++ {
		hl := high[i] - low[i]
		hc := math.Abs(high[i] - close[i-1])
		lc := math.Abs(low[i] - close[i-1])
		tr[i] = math.Max(hl, math.Max(hc, lc))
	}

	// Calculate ATR as EMA of True Range
	atr, err := EMA(tr, period)
	if err != nil {
		return nil, err
	}

	return atr, nil
}
