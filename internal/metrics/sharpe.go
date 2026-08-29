package metrics

import "math"

// CalculateSharpeRatio calculates Sharpe ratio (annualized)
// Assumes returns are periodic (e.g., daily)
func CalculateSharpeRatio(returns []float64, riskFreeRate float64, periodsPerYear float64) float64 {
	if len(returns) == 0 {
		return 0
	}

	avg := average(returns)
	std := standardDeviation(returns)

	if std == 0 {
		return 0
	}

	// Annualize
	excessReturn := (avg - riskFreeRate/periodsPerYear) * periodsPerYear
	annualizedStd := std * math.Sqrt(periodsPerYear)

	return excessReturn / annualizedStd
}

func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func standardDeviation(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	avg := average(values)
	var sumSquares float64
	for _, v := range values {
		diff := v - avg
		sumSquares += diff * diff
	}
	return math.Sqrt(sumSquares / float64(len(values)))
}
