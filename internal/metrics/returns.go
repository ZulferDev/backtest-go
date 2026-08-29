package metrics

import "math"

// EquityCurve represents equity progression over time
type EquityCurve struct {
	Timestamps []int64
	Equity     []float64
}

// CalculateReturns calculates periodic returns from equity curve
func CalculateReturns(curve EquityCurve) []float64 {
	if len(curve.Equity) < 2 {
		return []float64{}
	}

	returns := make([]float64, len(curve.Equity)-1)
	for i := 1; i < len(curve.Equity); i++ {
		if curve.Equity[i-1] != 0 {
			returns[i-1] = (curve.Equity[i] - curve.Equity[i-1]) / curve.Equity[i-1]
		}
	}
	return returns
}

// CalculateTotalReturn calculates overall return percentage
func CalculateTotalReturn(initialEquity, finalEquity float64) float64 {
	if initialEquity == 0 {
		return 0
	}
	return (finalEquity - initialEquity) / initialEquity * 100
}

// CalculateCAGR calculates Compound Annual Growth Rate
func CalculateCAGR(initialEquity, finalEquity float64, years float64) float64 {
	if initialEquity == 0 || years == 0 {
		return 0
	}
	return (math.Pow(finalEquity/initialEquity, 1/years) - 1) * 100
}
