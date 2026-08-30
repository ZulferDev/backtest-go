package strategies

import (
	"github.com/ZulferDev/backtest-go/internal/indicators"
	"github.com/ZulferDev/backtest-go/pkg/sdk"
)

// RATIONALE: Simple SMA crossover strategy for testing integration
// Buy when short SMA crosses above long SMA, sell when opposite
type SMACrossover struct {
	shortPeriod int
	longPeriod  int
}

func NewSMACrossover() *SMACrossover {
	return &SMACrossover{
		shortPeriod: 20,
		longPeriod:  50,
	}
}

func (s *SMACrossover) Init(ctx sdk.InitContext) error {
	return nil
}

func (s *SMACrossover) OnBar(ctx sdk.BarContext, bar sdk.OHLCV) error {
	history := ctx.History(s.longPeriod + 1)
	if len(history) < s.longPeriod+1 {
		return nil
	}

	// Extract closes from history
	closes := make([]float64, len(history))
	for i, h := range history {
		closes[i] = h.Close
	}

	// Calculate SMAs
	shortSMA, err := indicators.SMALast(closes, s.shortPeriod)
	if err != nil {
		return err
	}

	longSMA, err := indicators.SMALast(closes, s.longPeriod)
	if err != nil {
		return err
	}

	// Trading logic
	if !ctx.HasOpenPosition() && shortSMA > longSMA {
		return ctx.MarketBuy(1.0)
	} else if ctx.HasOpenPosition() && shortSMA < longSMA {
		return ctx.CloseAll()
	}

	return nil
}
