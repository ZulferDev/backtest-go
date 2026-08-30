package paper

import (
	"github.com/ZulferDev/backtest-go/pkg/sdk"
)

// Simple paper trading executor for testing (standalone without WebSocket)
type SimpleExecutor struct {
	equity      float64
	initialCash float64
	position    *Position
	trades      []Trade
}

// NewSimpleExecutor creates a simple paper trading executor for testing
func NewSimpleExecutor(initialCash float64) *SimpleExecutor {
	return &SimpleExecutor{
		equity:      initialCash,
		initialCash: initialCash,
		position:    nil,
		trades:      []Trade{},
	}
}

// GetEquity returns current equity
func (e *SimpleExecutor) GetEquity() float64 {
	return e.equity
}

// GetPosition returns current position or nil
func (e *SimpleExecutor) GetPosition() *Position {
	return e.position
}

// GetTrades returns all completed trades
func (e *SimpleExecutor) GetTrades() []Trade {
	return e.trades
}

// PlaceMarketOrder places a market order
func (e *SimpleExecutor) PlaceMarketOrder(side string, size float64, price float64) error {
	if size <= 0 {
		return sdk.ErrInvalidQuantity
	}

	// If we have a position, close it first
	if e.position != nil {
		// Calculate PnL
		var pnl float64
		if e.position.Side == "long" {
			pnl = (price - e.position.EntryPrice) * e.position.Size
		} else {
			pnl = (e.position.EntryPrice - price) * e.position.Size
		}

		// Simple fee: 0.1% per side
		fee := (e.position.EntryPrice + price) * e.position.Size * 0.001
		pnl -= fee

		trade := Trade{
			Side:       e.position.Side,
			EntryPrice: e.position.EntryPrice,
			ExitPrice:  price,
			Size:       e.position.Size,
			EntryTime:  e.position.EntryTime,
			ExitTime:   0, // Current time
			PnL:        pnl,
			Fee:        fee,
		}

		e.trades = append(e.trades, trade)
		e.equity += pnl
		e.position = nil
	}

	// Open new position if buy or long
	if side == "buy" || side == "long" {
		e.position = &Position{
			Side:       "long",
			Size:       size,
			EntryPrice: price,
			EntryTime:  0,
		}
	} else if side == "sell" || side == "short" {
		e.position = &Position{
			Side:       "short",
			Size:       size,
			EntryPrice: price,
			EntryTime:  0,
		}
	}

	return nil
}

// NewInitContext creates initialization context
func NewInitContext(executor *SimpleExecutor) sdk.InitContext {
	return &simpleInitContext{executor: executor}
}

// NewBarContext creates bar context
func NewBarContext(executor *SimpleExecutor, bar sdk.OHLCV, history []sdk.OHLCV) sdk.BarContext {
	return &simpleBarContext{
		executor: executor,
		bar:      bar,
		history:  history,
	}
}

// simpleInitContext implements sdk.InitContext
type simpleInitContext struct {
	executor *SimpleExecutor
}

// simpleBarContext implements sdk.BarContext
type simpleBarContext struct {
	executor *SimpleExecutor
	bar      sdk.OHLCV
	history  []sdk.OHLCV
}

func (c *simpleBarContext) CurrentBar() sdk.OHLCV {
	return c.bar
}

func (c *simpleBarContext) History(lookback int) []sdk.OHLCV {
	if lookback > len(c.history) {
		return c.history
	}
	start := len(c.history) - lookback
	if start < 0 {
		start = 0
	}
	return c.history[start:]
}

func (c *simpleBarContext) HasOpenPosition() bool {
	return c.executor.position != nil
}

func (c *simpleBarContext) CurrentPosition() sdk.Position {
	if c.executor.position == nil {
		return nil
	}
	return &simplePosition{pos: c.executor.position}
}

func (c *simpleBarContext) MarketBuy(quantity float64) error {
	return c.executor.PlaceMarketOrder("buy", quantity, c.bar.Close)
}

func (c *simpleBarContext) MarketSell(quantity float64) error {
	return c.executor.PlaceMarketOrder("sell", quantity, c.bar.Close)
}

func (c *simpleBarContext) CloseAll() error {
	if c.executor.position == nil {
		return nil
	}
	
	// Close with opposite side
	var side string
	if c.executor.position.Side == "long" {
		side = "sell"
	} else {
		side = "buy"
	}
	
	return c.executor.PlaceMarketOrder(side, c.executor.position.Size, c.bar.Close)
}

func (c *simpleBarContext) LogCustomMetric(key string, value float64) {
	// No-op for simple executor
}

// simplePosition wraps Position to implement sdk.Position
type simplePosition struct {
	pos *Position
}

func (p *simplePosition) Size() float64 {
	return p.pos.Size
}

func (p *simplePosition) EntryPrice() float64 {
	return p.pos.EntryPrice
}

func (p *simplePosition) UnrealizedPnL(currentPrice float64) float64 {
	if p.pos.Side == "long" {
		return (currentPrice - p.pos.EntryPrice) * p.pos.Size
	}
	return (p.pos.EntryPrice - currentPrice) * p.pos.Size
}

func (p *simplePosition) Side() string {
	return p.pos.Side
}
