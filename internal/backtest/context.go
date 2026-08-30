package backtest

import (
	"fmt"
	"github.com/ZulferDev/backtest-go/pkg/data"
	"github.com/ZulferDev/backtest-go/pkg/sdk"
)

// initContext implements sdk.InitContext
type initContext struct{}

// barContext implements sdk.BarContext
type barContext struct {
	engine      *Engine
	currentBar  sdk.OHLCV
	historyData []data.OHLCV
	currentIdx  int
}

func (c *barContext) CurrentBar() sdk.OHLCV {
	return c.currentBar
}

func (c *barContext) History(lookback int) []sdk.OHLCV {
	if lookback <= 0 || c.currentIdx < lookback {
		return []sdk.OHLCV{}
	}

	start := c.currentIdx - lookback
	var result []sdk.OHLCV
	for i := start; i <= c.currentIdx; i++ {
		bar := c.historyData[i]
		result = append(result, sdk.OHLCV{
			Timestamp: bar.Timestamp,
			Open:      bar.Open,
			High:      bar.High,
			Low:       bar.Low,
			Close:     bar.Close,
			Volume:    bar.Volume,
		})
	}
	return result
}

func (c *barContext) HasOpenPosition() bool {
	return c.engine.state.position != nil
}

func (c *barContext) CurrentPosition() sdk.Position {
	return c.engine.state.position
}

func (c *barContext) MarketBuy(quantity float64) error {
	if quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}
	if c.HasOpenPosition() {
		return fmt.Errorf("position already open")
	}

	// Use current close as base price
	basePrice := c.currentBar.Close
	
	// Apply fees and slippage if configured
	fillPrice := basePrice
	fee := 0.0
	
	if c.engine.executionSim != nil {
		var err error
		fillPrice, _, fee, err = c.engine.executionSim.SimulateExecution(basePrice, quantity, "buy")
		if err != nil {
			return fmt.Errorf("execution simulation failed: %w", err)
		}
	}

	c.engine.state.position = &Position{
		side:       "long",
		size:       quantity,
		entryPrice: fillPrice,
		entryTime:  c.currentBar.Timestamp,
		entryFee:   fee,
	}

	return nil
}

func (c *barContext) MarketSell(quantity float64) error {
	if quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}
	if c.HasOpenPosition() {
		return fmt.Errorf("position already open")
	}

	// Use current close as base price
	basePrice := c.currentBar.Close
	
	// Apply fees and slippage if configured
	fillPrice := basePrice
	fee := 0.0
	
	if c.engine.executionSim != nil {
		var err error
		fillPrice, _, fee, err = c.engine.executionSim.SimulateExecution(basePrice, quantity, "sell")
		if err != nil {
			return fmt.Errorf("execution simulation failed: %w", err)
		}
	}

	c.engine.state.position = &Position{
		side:       "short",
		size:       quantity,
		entryPrice: fillPrice,
		entryTime:  c.currentBar.Timestamp,
		entryFee:   fee,
	}

	return nil
}

func (c *barContext) CloseAll() error {
	if !c.HasOpenPosition() {
		return fmt.Errorf("no position to close")
	}

	pos := c.engine.state.position
	basePrice := c.currentBar.Close
	
	// Apply fees and slippage for exit
	exitPrice := basePrice
	exitFee := 0.0
	
	if c.engine.executionSim != nil {
		var err error
		side := "sell"
		if pos.side == "short" {
			side = "buy" // Close short = buy
		}
		exitPrice, _, exitFee, err = c.engine.executionSim.SimulateExecution(basePrice, pos.size, side)
		if err != nil {
			return fmt.Errorf("execution simulation failed: %w", err)
		}
	}
	
	pnl := pos.UnrealizedPnL(exitPrice)
	totalFee := pos.entryFee + exitFee
	pnl -= totalFee // Subtract fees from PnL

	// Record trade
	trade := Trade{
		Side:       pos.side,
		EntryPrice: pos.entryPrice,
		ExitPrice:  exitPrice,
		Size:       pos.size,
		EntryTime:  pos.entryTime,
		ExitTime:   c.currentBar.Timestamp,
		PnL:        pnl,
		Fee:        totalFee,
	}

	c.engine.state.trades = append(c.engine.state.trades, trade)
	c.engine.state.equity = c.engine.state.initialCash + pnl
	c.engine.state.position = nil

	return nil
}

func (c *barContext) LogCustomMetric(key string, value float64) {
	// TODO: implement custom metric logging
}
