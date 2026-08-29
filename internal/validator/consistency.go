package validator

import (
	"math"
	"github.com/ZulferDev/backtest-go/pkg/data"
)

// DetectOutliers checks for sudden abnormal price movement (e.g. > threshold % in 1 bar)
func DetectOutliers(series []data.OHLCV, thresholdPct float64) []int {
	var outlierIndices []int
	if len(series) < 2 { return outlierIndices }

	for i := 1; i < len(series); i++ {
		prevClose := series[i-1].Close
		if prevClose == 0 { continue }

		change := math.Abs(series[i].Close - prevClose) / prevClose
		if change > thresholdPct {
			outlierIndices = append(outlierIndices, i)
		}
	}
	return outlierIndices
}
