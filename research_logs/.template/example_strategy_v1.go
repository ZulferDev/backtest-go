package strategies

import (
	"github.com/ZulferDev/backtest-go/internal/indicators"
	"github.com/ZulferDev/backtest-go/pkg/sdk"
)

// RSIMeanReversion implements RSI-based mean reversion strategy
// RATIONALE: Enter at RSI extremes, exit at neutral levels
type RSIMeanReversion struct {
	rsiPeriod    int
	oversold     float64
	overbought   float64
	neutral      float64
	stopLossPct  float64
	entryPrice   float64
	stopLossPrice float64
}

// Init initializes the strategy parameters
// RATIONALE: Setup RSI parameters for mean reversion detection
func (s *RSIMeanReversion) Init(ctx sdk.InitContext) error {
	s.rsiPeriod = 14      // Standard RSI period
	s.oversold = 30.0     // Oversold threshold
	s.overbought = 70.0   // Overbought threshold
	s.neutral = 50.0      // Neutral exit level
	s.stopLossPct = 0.05  // 5% stop loss
	s.entryPrice = 0.0
	s.stopLossPrice = 0.0
	return nil
}

// OnBar executes trading logic for each bar
// RATIONALE: Buy oversold, sell overbought, exit at neutral or stop loss
func (s *RSIMeanReversion) OnBar(ctx sdk.BarContext, bar sdk.OHLCV) error {
	// Need enough history for RSI calculation
	history := ctx.History(s.rsiPeriod + 10)
	if len(history) < s.rsiPeriod+1 {
		return nil // Not enough data yet
	}

	// Extract close prices
	closes := make([]float64, len(history))
	for i, h := range history {
		closes[i] = h.Close
	}

	// Calculate RSI
	rsi, err := indicators.RSILast(closes, s.rsiPeriod)
	if err != nil {
		return err
	}

	currentPrice := bar.Close

	// If we have no position, look for entry signals
	if !ctx.HasOpenPosition() {
		// ENTRY LOGIC: Buy when RSI crosses below oversold
		if rsi < s.oversold {
			// Calculate position size (95% of equity)
			size := (ctx.Equity() * 0.95) / currentPrice
			
			// Place market buy order
			ctx.MarketBuy(size)
			
			// Record entry price and calculate stop loss
			s.entryPrice = currentPrice
			s.stopLossPrice = currentPrice * (1.0 - s.stopLossPct)
		}
		
		// Note: We only trade long in this version
		// Short positions could be added for RSI > overbought
	} else {
		// If we have a position, check exit conditions
		
		// EXIT LOGIC 1: Stop loss hit
		if currentPrice <= s.stopLossPrice {
			ctx.CloseAll()
			return nil
		}
		
		// EXIT LOGIC 2: RSI returns to neutral (mean reversion complete)
		if rsi > s.neutral {
			ctx.CloseAll()
			return nil
		}
	}

	return nil
}
