package backtest

import (
	"fmt"
	"github.com/ZulferDev/backtest-go/pkg/data"
	"github.com/ZulferDev/backtest-go/pkg/sdk"
)

// Engine is the core backtest execution engine
type Engine struct {
	strategy sdk.Strategy
	data     []data.OHLCV
	state    *State
}

// State holds the current backtest state
type State struct {
	currentIndex int
	position     *Position
	equity       float64
	initialCash  float64
	trades       []Trade
}

// Position represents an open position
type Position struct {
	side       string  // "long" or "short"
	size       float64
	entryPrice float64
	entryTime  int64
}

func (p *Position) Size() float64 {
	return p.size
}

func (p *Position) EntryPrice() float64 {
	return p.entryPrice
}

func (p *Position) Side() string {
	return p.side
}

func (p *Position) UnrealizedPnL(currentPrice float64) float64 {
	if p.side == "long" {
		return (currentPrice - p.entryPrice) * p.size
	}
	return (p.entryPrice - currentPrice) * p.size
}

// Trade represents a completed trade
type Trade struct {
	Side       string
	EntryPrice float64
	ExitPrice  float64
	Size       float64
	EntryTime  int64
	ExitTime   int64
	PnL        float64
	Fee        float64
}

// NewEngine creates a new backtest engine
func NewEngine(strategy sdk.Strategy, historicalData []data.OHLCV, initialCash float64) *Engine {
	return &Engine{
		strategy: strategy,
		data:     historicalData,
		state: &State{
			currentIndex: 0,
			position:     nil,
			equity:       initialCash,
			initialCash:  initialCash,
			trades:       []Trade{},
		},
	}
}

// Run executes the backtest
func (e *Engine) Run() error {
	// Initialize strategy
	initCtx := &initContext{}
	if err := e.strategy.Init(initCtx); err != nil {
		return fmt.Errorf("strategy init failed: %w", err)
	}

	// Iterate through all bars
	for i := 0; i < len(e.data); i++ {
		e.state.currentIndex = i
		bar := e.data[i]

		// Convert to SDK OHLCV
		sdkBar := sdk.OHLCV{
			Timestamp: bar.Timestamp,
			Open:      bar.Open,
			High:      bar.High,
			Low:       bar.Low,
			Close:     bar.Close,
			Volume:    bar.Volume,
		}

		// Create bar context
		barCtx := &barContext{
			engine:      e,
			currentBar:  sdkBar,
			historyData: e.data,
			currentIdx:  i,
		}

		// Call strategy
		if err := e.strategy.OnBar(barCtx, sdkBar); err != nil {
			return fmt.Errorf("strategy OnBar failed at index %d: %w", i, err)
		}

		// Update position PnL if open
		if e.state.position != nil {
			unrealizedPnL := e.state.position.UnrealizedPnL(bar.Close)
			e.state.equity = e.state.initialCash + unrealizedPnL
		}
	}

	return nil
}

// GetState returns current backtest state
func (e *Engine) GetState() *State {
	return e.state
}

// Accessor methods for State (for report generation)
func (s *State) Trades() []Trade {
	return s.trades
}

func (s *State) InitialCash() float64 {
	return s.initialCash
}

func (s *State) Equity() float64 {
	return s.equity
}
