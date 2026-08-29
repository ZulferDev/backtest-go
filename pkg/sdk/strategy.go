package sdk

// Strategy is the interface that all AI-generated strategies must implement
type Strategy interface {
	// Init is called once before backtest starts
	Init(ctx InitContext) error

	// OnBar is called for each new bar/candle
	OnBar(ctx BarContext, bar OHLCV) error
}
