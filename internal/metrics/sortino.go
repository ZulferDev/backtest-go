package metrics

import "math"

// CalculateSortinoRatio calculates Sortino ratio (annualized)
// Only considers downside deviation
func CalculateSortinoRatio(returns []float64, targetReturn float64, periodsPerYear float64) float64 {
	if len(returns) == 0 {
		return 0
	}

	avg := average(returns)
	downsideStd := downsideDeviation(returns, targetReturn)

	if downsideStd == 0 {
		return 0
	}

	// Annualize
	excessReturn := (avg - targetReturn/periodsPerYear) * periodsPerYear
	annualizedDownsideStd := downsideStd * math.Sqrt(periodsPerYear)

	return excessReturn / annualizedDownsideStd
}

func downsideDeviation(values []float64, target float64) float64 {
	if len(values) == 0 {
		return 0
	}

	var sumSquares float64
	count := 0
	for _, v := range values {
		if v < target {
			diff := v - target
			sumSquares += diff * diff
			count++
		}
	}

	if count == 0 {
		return 0
	}

	return math.Sqrt(sumSquares / float64(count))
}
