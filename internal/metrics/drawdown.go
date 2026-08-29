package metrics

import "math"

// Drawdown represents a drawdown period
type Drawdown struct {
	StartTime int64
	EndTime   int64
	Depth     float64 // Percentage
}

// CalculateMaxDrawdown calculates maximum drawdown from equity curve
func CalculateMaxDrawdown(curve EquityCurve) float64 {
	if len(curve.Equity) == 0 {
		return 0
	}

	var maxDD float64
	peak := curve.Equity[0]

	for _, equity := range curve.Equity {
		if equity > peak {
			peak = equity
		}

		dd := (peak - equity) / peak * 100
		if dd > maxDD {
			maxDD = dd
		}
	}

	return maxDD
}

// CalculateDrawdowns identifies all drawdown periods
func CalculateDrawdowns(curve EquityCurve) []Drawdown {
	if len(curve.Equity) == 0 {
		return []Drawdown{}
	}

	var drawdowns []Drawdown
	peak := curve.Equity[0]
	peakIdx := 0
	inDrawdown := false
	var currentDD Drawdown

	for i, equity := range curve.Equity {
		if equity > peak {
			if inDrawdown {
				// Drawdown recovered
				currentDD.EndTime = curve.Timestamps[i-1]
				currentDD.Depth = (peak - math.Min(peak, equity)) / peak * 100
				drawdowns = append(drawdowns, currentDD)
				inDrawdown = false
			}
			peak = equity
			peakIdx = i
		} else if equity < peak {
			if !inDrawdown {
				// New drawdown starts
				inDrawdown = true
				currentDD = Drawdown{
					StartTime: curve.Timestamps[peakIdx],
				}
			}
		}
	}

	// Handle ongoing drawdown at the end
	if inDrawdown {
		currentDD.EndTime = curve.Timestamps[len(curve.Timestamps)-1]
		currentDD.Depth = (peak - curve.Equity[len(curve.Equity)-1]) / peak * 100
		drawdowns = append(drawdowns, currentDD)
	}

	return drawdowns
}
