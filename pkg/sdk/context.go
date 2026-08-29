package sdk

// OHLCV represents a candlestick bar
type OHLCV struct {
	Timestamp int64
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
}

// Position represents an open trading position
type Position interface {
	Size() float64
	EntryPrice() float64
	UnrealizedPnL(currentPrice float64) float64
	Side() string // "long" or "short"
}

// InitContext provides methods available during strategy initialization
type InitContext interface {
	// Future: RegisterSMA, RegisterRSI, etc.
}

// BarContext provides methods available during OnBar execution
type BarContext interface {
	// Market Data
	CurrentBar() OHLCV
	History(lookback int) []OHLCV

	// Position Info
	HasOpenPosition() bool
	CurrentPosition() Position

	// Order Execution
	MarketBuy(quantity float64) error
	MarketSell(quantity float64) error
	CloseAll() error

	// Metrics
	LogCustomMetric(key string, value float64)
}
