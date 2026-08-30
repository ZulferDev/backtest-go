package execution

import (
	"github.com/ZulferDev/backtest-go/pkg/sdk"
)

// liveInitContext implements sdk.InitContext for live trading
type liveInitContext struct {
	bridge *Bridge
}

func (c *liveInitContext) Log(msg string) {
	// Log during initialization
}

// liveBarContext implements sdk.BarContext for live trading
type liveBarContext struct {
	bridge *Bridge
	bar    sdk.OHLCV
}

// Market order methods
func (c *liveBarContext) MarketBuy(quantity float64) error {
	return c.bridge.PlaceMarketOrder("buy", quantity)
}

func (c *liveBarContext) MarketSell(quantity float64) error {
	return c.bridge.PlaceMarketOrder("sell", quantity)
}

func (c *liveBarContext) MarketOrder(side string, size float64) error {
	return c.bridge.PlaceMarketOrder(side, size)
}

func (c *liveBarContext) Buy(size float64) error {
	return c.bridge.PlaceMarketOrder("long", size)
}

func (c *liveBarContext) Sell(size float64) error {
	return c.bridge.PlaceMarketOrder("short", size)
}

func (c *liveBarContext) CloseAll() error {
	pos := c.bridge.GetPosition()
	if pos == nil {
		return nil
	}
	
	// Close by placing opposite order
	var side string
	if pos.Side == "long" {
		side = "sell"
	} else {
		side = "buy"
	}
	
	return c.bridge.PlaceMarketOrder(side, pos.Size)
}

// Position query methods
func (c *liveBarContext) HasOpenPosition() bool {
	return c.bridge.GetPosition() != nil
}

func (c *liveBarContext) HasPosition() bool {
	return c.HasOpenPosition()
}

func (c *liveBarContext) CurrentPosition() sdk.Position {
	pos := c.bridge.GetPosition()
	if pos == nil {
		return nil
	}
	return &livePosition{pos: pos}
}

func (c *liveBarContext) PositionSize() float64 {
	pos := c.bridge.GetPosition()
	if pos == nil {
		return 0
	}
	return pos.Size
}

func (c *liveBarContext) PositionSide() string {
	pos := c.bridge.GetPosition()
	if pos == nil {
		return ""
	}
	return pos.Side
}

func (c *liveBarContext) Equity() float64 {
	return c.bridge.GetEquity()
}

// CurrentBar returns the current bar being processed
func (c *liveBarContext) CurrentBar() sdk.OHLCV {
	return c.bar
}

// History returns historical bars (limited for live trading)
func (c *liveBarContext) History(lookback int) []sdk.OHLCV {
	// For live trading, history would need to be fetched from exchange
	// or maintained in a rolling buffer
	return []sdk.OHLCV{c.bar}
}

// Log prints a message
func (c *liveBarContext) Log(msg string) {
	// Logging implementation
}

// LogCustomMetric logs a custom metric
func (c *liveBarContext) LogCustomMetric(key string, value float64) {
	// Custom metric logging implementation
}

// livePosition wraps the execution Position to implement sdk.Position
type livePosition struct {
	pos *Position
}

func (p *livePosition) Size() float64 {
	return p.pos.Size
}

func (p *livePosition) EntryPrice() float64 {
	return p.pos.EntryPrice
}

func (p *livePosition) UnrealizedPnL(currentPrice float64) float64 {
	if p.pos.Side == "long" {
		return (currentPrice - p.pos.EntryPrice) * p.pos.Size
	}
	return (p.pos.EntryPrice - currentPrice) * p.pos.Size
}

func (p *livePosition) Side() string {
	return p.pos.Side
}
