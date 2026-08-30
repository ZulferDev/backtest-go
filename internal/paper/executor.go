package paper

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ZulferDev/backtest-go/pkg/data"
	"github.com/ZulferDev/backtest-go/pkg/sdk"
)

// Executor manages paper trading execution state
type Executor struct {
	strategy     sdk.Strategy
	wsClient     *WebSocketClient
	state        *TradingState
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
	dataChan     chan data.OHLCV
	onTrade      func(Trade)
	onStateUpdate func(*TradingState)
}

// TradingState represents current paper trading state
type TradingState struct {
	Position      *Position
	Equity        float64
	InitialCash   float64
	Trades        []Trade
	LastUpdate    time.Time
	CurrentBar    data.OHLCV
	BarCount      int
	IsRunning     bool
}

// Position represents an open position in paper trading
type Position struct {
	Side       string
	Size       float64
	EntryPrice float64
	EntryTime  int64
}

// Trade represents a completed paper trade
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

// NewExecutor creates a new paper trading executor
func NewExecutor(ctx context.Context, strategy sdk.Strategy, symbol, interval string, initialCash float64) (*Executor, error) {
	execCtx, cancel := context.WithCancel(ctx)

	wsClient, err := NewWebSocketClient(execCtx, symbol, interval)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create websocket client: %w", err)
	}

	executor := &Executor{
		strategy: strategy,
		wsClient: wsClient,
		state: &TradingState{
			Position:    nil,
			Equity:      initialCash,
			InitialCash: initialCash,
			Trades:      []Trade{},
			LastUpdate:  time.Now(),
			IsRunning:   false,
		},
		ctx:      execCtx,
		cancel:   cancel,
		dataChan: wsClient.Subscribe(),
	}

	return executor, nil
}

// Start begins paper trading execution
func (e *Executor) Start() error {
	e.mu.Lock()
	if e.state.IsRunning {
		e.mu.Unlock()
		return fmt.Errorf("executor already running")
	}
	e.state.IsRunning = true
	e.mu.Unlock()

	// Initialize strategy
	initCtx := &paperInitContext{}
	if err := e.strategy.Init(initCtx); err != nil {
		return fmt.Errorf("strategy init failed: %w", err)
	}

	log.Println("Paper trading executor started")
	go e.run()

	return nil
}

// run is the main execution loop
func (e *Executor) run() {
	defer e.Stop()

	for {
		select {
		case <-e.ctx.Done():
			log.Println("Paper trading executor stopped by context")
			return

		case ohlcv, ok := <-e.dataChan:
			if !ok {
				log.Println("Data channel closed, stopping executor")
				return
			}

			if err := e.processBar(ohlcv); err != nil {
				log.Printf("Error processing bar: %v", err)
			}
		}
	}
}

// processBar processes a new OHLCV bar
func (e *Executor) processBar(ohlcv data.OHLCV) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.state.CurrentBar = ohlcv
	e.state.BarCount++
	e.state.LastUpdate = time.Now()

	// Update unrealized PnL if position is open
	if e.state.Position != nil {
		unrealizedPnL := e.calculateUnrealizedPnL(ohlcv.Close)
		e.state.Equity = e.state.InitialCash + unrealizedPnL

		// Apply accumulated realized PnL from closed trades
		for _, trade := range e.state.Trades {
			e.state.Equity += trade.PnL
		}
	}

	// Convert to SDK OHLCV
	sdkBar := sdk.OHLCV{
		Timestamp: ohlcv.Timestamp,
		Open:      ohlcv.Open,
		High:      ohlcv.High,
		Low:       ohlcv.Low,
		Close:     ohlcv.Close,
		Volume:    ohlcv.Volume,
	}

	// Create bar context
	barCtx := &paperBarContext{
		executor: e,
		bar:      sdkBar,
	}

	// Call strategy
	if err := e.strategy.OnBar(barCtx, sdkBar); err != nil {
		return fmt.Errorf("strategy OnBar failed: %w", err)
	}

	// Trigger state update callback
	if e.onStateUpdate != nil {
		stateCopy := e.copyState()
		go e.onStateUpdate(stateCopy)
	}

	return nil
}

// calculateUnrealizedPnL calculates unrealized PnL for current position
func (e *Executor) calculateUnrealizedPnL(currentPrice float64) float64 {
	if e.state.Position == nil {
		return 0
	}

	pos := e.state.Position
	if pos.Side == "long" {
		return (currentPrice - pos.EntryPrice) * pos.Size
	}
	return (pos.EntryPrice - currentPrice) * pos.Size
}

// OpenPosition opens a new position (called by bar context)
func (e *Executor) OpenPosition(side string, size float64, price float64) error {
	if e.state.Position != nil {
		return fmt.Errorf("position already open")
	}

	e.state.Position = &Position{
		Side:       side,
		Size:       size,
		EntryPrice: price,
		EntryTime:  e.state.CurrentBar.Timestamp,
	}

	log.Printf("Position opened: %s %.4f @ %.2f", side, size, price)
	return nil
}

// ClosePosition closes the current position (called by bar context)
func (e *Executor) ClosePosition(exitPrice float64) error {
	if e.state.Position == nil {
		return fmt.Errorf("no position to close")
	}

	pos := e.state.Position
	var pnl float64

	if pos.Side == "long" {
		pnl = (exitPrice - pos.EntryPrice) * pos.Size
	} else {
		pnl = (pos.EntryPrice - exitPrice) * pos.Size
	}

	// Simple fee model: 0.1% per side
	fee := (pos.EntryPrice + exitPrice) * pos.Size * 0.001
	pnl -= fee

	trade := Trade{
		Side:       pos.Side,
		EntryPrice: pos.EntryPrice,
		ExitPrice:  exitPrice,
		Size:       pos.Size,
		EntryTime:  pos.EntryTime,
		ExitTime:   e.state.CurrentBar.Timestamp,
		PnL:        pnl,
		Fee:        fee,
	}

	e.state.Trades = append(e.state.Trades, trade)
	e.state.Position = nil
	e.state.Equity = e.state.InitialCash

	// Add all realized PnL
	for _, t := range e.state.Trades {
		e.state.Equity += t.PnL
	}

	log.Printf("Position closed: %.4f @ %.2f | PnL: %.2f", pos.Size, exitPrice, pnl)

	if e.onTrade != nil {
		go e.onTrade(trade)
	}

	return nil
}

// GetState returns a copy of the current trading state
func (e *Executor) GetState() *TradingState {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.copyState()
}

// copyState creates a deep copy of trading state (must hold lock)
func (e *Executor) copyState() *TradingState {
	stateCopy := &TradingState{
		Equity:      e.state.Equity,
		InitialCash: e.state.InitialCash,
		LastUpdate:  e.state.LastUpdate,
		CurrentBar:  e.state.CurrentBar,
		BarCount:    e.state.BarCount,
		IsRunning:   e.state.IsRunning,
	}

	if e.state.Position != nil {
		stateCopy.Position = &Position{
			Side:       e.state.Position.Side,
			Size:       e.state.Position.Size,
			EntryPrice: e.state.Position.EntryPrice,
			EntryTime:  e.state.Position.EntryTime,
		}
	}

	stateCopy.Trades = make([]Trade, len(e.state.Trades))
	copy(stateCopy.Trades, e.state.Trades)

	return stateCopy
}

// SetTradeCallback sets callback for trade events
func (e *Executor) SetTradeCallback(callback func(Trade)) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.onTrade = callback
}

// SetStateUpdateCallback sets callback for state updates
func (e *Executor) SetStateUpdateCallback(callback func(*TradingState)) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.onStateUpdate = callback
}

// Stop stops the paper trading executor
func (e *Executor) Stop() {
	e.mu.Lock()
	if !e.state.IsRunning {
		e.mu.Unlock()
		return
	}
	e.state.IsRunning = false
	e.mu.Unlock()

	e.cancel()
	if e.wsClient != nil {
		e.wsClient.Close()
	}

	log.Println("Paper trading executor stopped")
}

// IsRunning returns whether executor is currently running
func (e *Executor) IsRunning() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.state.IsRunning
}
