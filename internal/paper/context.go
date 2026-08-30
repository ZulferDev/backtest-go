package paper

import (
	"github.com/ZulferDev/backtest-go/pkg/sdk"
)

// paperInitContext implements sdk.InitContext for paper trading
type paperInitContext struct{}

func (c *paperInitContext) Log(msg string) {
	// Implementation placeholder for logging during init
}

// paperBarContext implements sdk.BarContext for paper trading
type paperBarContext struct {
	executor *Executor
	bar      sdk.OHLCV
}

// Market order methods
func (c *paperBarContext) MarketOrder(side string, size float64) error {
	c.executor.mu.Lock()
	defer c.executor.mu.Unlock()

	// Close existing position if opposite side
	if c.executor.state.Position != nil {
		if c.executor.state.Position.Side != side {
			if err := c.executor.ClosePosition(c.bar.Close); err != nil {
				return err
			}
		}
	}

	// Open new position at current close price (simulated slippage)
	slippage := 0.0005 // 0.05% slippage
	executionPrice := c.bar.Close
	if side == "long" {
		executionPrice *= (1 + slippage)
	} else {
		executionPrice *= (1 - slippage)
	}

	return c.executor.OpenPosition(side, size, executionPrice)
}

func (c *paperBarContext) Buy(size float64) error {
	return c.MarketOrder("long", size)
}

func (c *paperBarContext) Sell(size float64) error {
	return c.MarketOrder("short", size)
}

func (c *paperBarContext) CloseAll() error {
	c.executor.mu.Lock()
	defer c.executor.mu.Unlock()

	if c.executor.state.Position == nil {
		return nil // No position to close
	}

	slippage := 0.0005
	executionPrice := c.bar.Close
	if c.executor.state.Position.Side == "long" {
		executionPrice *= (1 - slippage) // Worse price when selling
	} else {
		executionPrice *= (1 + slippage) // Worse price when covering short
	}

	return c.executor.ClosePosition(executionPrice)
}

// Position query methods
func (c *paperBarContext) HasPosition() bool {
	c.executor.mu.RLock()
	defer c.executor.mu.RUnlock()
	return c.executor.state.Position != nil
}

func (c *paperBarContext) PositionSize() float64 {
	c.executor.mu.RLock()
	defer c.executor.mu.RUnlock()
	
	if c.executor.state.Position == nil {
		return 0
	}
	return c.executor.state.Position.Size
}

func (c *paperBarContext) PositionSide() string {
	c.executor.mu.RLock()
	defer c.executor.mu.RUnlock()
	
	if c.executor.state.Position == nil {
		return ""
	}
	return c.executor.state.Position.Side
}

func (c *paperBarContext) Equity() float64 {
	c.executor.mu.RLock()
	defer c.executor.mu.RUnlock()
	return c.executor.state.Equity
}

// CurrentBar returns the current bar being processed
func (c *paperBarContext) CurrentBar() sdk.OHLCV {
	return c.bar
}

// History returns historical bars (limited lookback for paper trading)
func (c *paperBarContext) History(lookback int) []sdk.OHLCV {
	// For paper trading, we don't maintain full history buffer
	// This would need to be implemented with a rolling buffer if needed
	return []sdk.OHLCV{c.bar}
}

// Log prints a message (paper trading version)
func (c *paperBarContext) Log(msg string) {
	// Logging implementation
}
