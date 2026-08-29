package normalizer

import (
	"github.com/ZulferDev/backtest-go/pkg/data"
)

// ToChronological ensures the data is sorted by Timestamp ascending
func ToChronological(series []data.OHLCV) []data.OHLCV {
	if len(series) < 2 { return series }
	
	if series[0].Timestamp > series[len(series)-1].Timestamp {
		// Reverse array
		for i, j := 0, len(series)-1; i < j; i, j = i+1, j-1 {
			series[i], series[j] = series[j], series[i]
		}
	}
	return series
}
