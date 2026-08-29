package indicators

import "fmt"

// MACDResult holds MACD line, signal line, and histogram
type MACDResult struct {
	MACD      []float64
	Signal    []float64
	Histogram []float64
}

// MACD calculates Moving Average Convergence Divergence
// Standard periods: fast=12, slow=26, signal=9
// MACD Line = EMA(fast) - EMA(slow)
// Signal Line = EMA(MACD, signal)
// Histogram = MACD - Signal
func MACD(data []float64, fastPeriod, slowPeriod, signalPeriod int) (*MACDResult, error) {
	if fastPeriod <= 0 || slowPeriod <= 0 || signalPeriod <= 0 {
		return nil, fmt.Errorf("all periods must be positive")
	}
	if fastPeriod >= slowPeriod {
		return nil, fmt.Errorf("fast period must be less than slow period")
	}
	if len(data) < slowPeriod {
		return &MACDResult{
			MACD:      make([]float64, len(data)),
			Signal:    make([]float64, len(data)),
			Histogram: make([]float64, len(data)),
		}, nil
	}

	// Calculate fast and slow EMAs
	fastEMA, err := EMA(data, fastPeriod)
	if err != nil {
		return nil, err
	}
	slowEMA, err := EMA(data, slowPeriod)
	if err != nil {
		return nil, err
	}

	// Calculate MACD line
	macdLine := make([]float64, len(data))
	for i := 0; i < len(data); i++ {
		macdLine[i] = fastEMA[i] - slowEMA[i]
	}

	// Calculate signal line (EMA of MACD)
	signalLine, err := EMA(macdLine, signalPeriod)
	if err != nil {
		return nil, err
	}

	// Calculate histogram
	histogram := make([]float64, len(data))
	for i := 0; i < len(data); i++ {
		histogram[i] = macdLine[i] - signalLine[i]
	}

	return &MACDResult{
		MACD:      macdLine,
		Signal:    signalLine,
		Histogram: histogram,
	}, nil
}
