package main

import (
	"github.com/ZulferDev/backtest-go/pkg/sdk"
)

// TemplateStrategy is a template for AI-generated strategies
type TemplateStrategy struct {
	// Add your state variables here
	// Example:
	// shortPeriod int
	// longPeriod  int
	// rsiPeriod   int
}

// Init initializes the strategy with parameters
// RATIONALE: Setup phase - initialize indicators and parameters
func (s *TemplateStrategy) Init(ctx sdk.InitContext) error {
	// Initialize your parameters here
	// Example:
	// s.shortPeriod = 20
	// s.longPeriod = 50
	// s.rsiPeriod = 14
	
	return nil
}

// OnBar is called for each new bar/candle
// RATIONALE: Main trading logic - analyze bar and generate signals
func (s *TemplateStrategy) OnBar(ctx sdk.BarContext, bar sdk.OHLCV) error {
	// Get historical data
	// history := ctx.History(s.longPeriod + 1)
	// if len(history) < s.longPeriod+1 {
	//     return nil // Not enough data yet
	// }
	
	// Extract close prices
	// closes := make([]float64, len(history))
	// for i, h := range history {
	//     closes[i] = h.Close
	// }
	
	// Calculate indicators
	// Example: SMA crossover
	// shortSMA, _ := indicators.SMALast(closes, s.shortPeriod)
	// longSMA, _ := indicators.SMALast(closes, s.longPeriod)
	
	// Example: RSI
	// rsi, _ := indicators.RSILast(closes, s.rsiPeriod)
	
	// Trading logic
	// Example: Buy when short SMA crosses above long SMA
	// if !ctx.HasOpenPosition() && shortSMA > longSMA {
	//     size := ctx.Equity() * 0.95 / bar.Close // 95% of equity
	//     ctx.MarketBuy(size)
	// }
	
	// Example: Sell when short SMA crosses below long SMA
	// if ctx.HasOpenPosition() && shortSMA < longSMA {
	//     ctx.CloseAll()
	// }
	
	return nil
}
