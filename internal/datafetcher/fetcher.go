package datafetcher

import (
	"context"
	"github.com/ZulferDev/backtest-go/pkg/data"
)

// Fetcher defines the interface for exchange clients
type Fetcher interface {
	FetchKlines(ctx context.Context, symbol string, interval string, limit int, startTime, endTime int64) ([]data.OHLCV, error)
}
