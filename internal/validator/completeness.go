package validator

import "backtest-go/pkg/data"

// CheckGaps checks for timestamp gaps in ms based on timeframe duration in ms
func CheckGaps(series []data.OHLCV, expectedIntervalMs int64) []int {
	var gapIndices []int
	if len(series) < 2 { return gapIndices }

	for i := 1; i < len(series); i++ {
		diff := series[i].Timestamp - series[i-1].Timestamp
		if diff != expectedIntervalMs {
			gapIndices = append(gapIndices, i)
		}
	}
	return gapIndices
}
