package execution

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ZulferDev/backtest-go/pkg/data"
	"github.com/ZulferDev/backtest-go/pkg/sdk"
)

// OrderType represents the type of order
type OrderType string

const (
	OrderTypeMarket OrderType = "MARKET"
	OrderTypeLimit  OrderType = "LIMIT"
)

// OrderSide represents the side of an order
type OrderSide string

const (
	OrderSideBuy  OrderSide = "BUY"
	OrderSideSell OrderSide = "SELL"
)

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "PENDING"
	OrderStatusFilled    OrderStatus = "FILLED"
	OrderStatusCancelled OrderStatus = "CANCELLED"
	OrderStatusFailed    OrderStatus = "FAILED"
)

// Order represents a live order
type Order struct {
	ID            string
	Symbol        string
	Side          OrderSide
	Type          OrderType
	Quantity      float64
	Price         float64
	Status        OrderStatus
	FilledQty     float64
	AvgFillPrice  float64
	Timestamp     int64
	ExchangeID    string
	ErrorMessage  string
}

// ExchangeAdapter interface for exchange-specific implementations
type ExchangeAdapter interface {
	// Order execution
	PlaceOrder(ctx context.Context, symbol string, side OrderSide, orderType OrderType, quantity float64, price float64) (*Order, error)
	CancelOrder(ctx context.Context, orderID string) error
	GetOrder(ctx context.Context, orderID string) (*Order, error)
	
	// Account information
	GetBalance(ctx context.Context) (float64, error)
	GetPosition(ctx context.Context, symbol string) (*Position, error)
	
	// Market data
	GetCurrentPrice(ctx context.Context, symbol string) (float64, error)
	
	// Connection management
	Connect(ctx context.Context) error
	Disconnect() error
	IsConnected() bool
}

// Position represents a live position
type Position struct {
	Symbol     string
	Side       string
	Size       float64
	EntryPrice float64
	EntryTime  int64
	Leverage   float64
}

// Bridge connects paper trading executor to live exchange
type Bridge struct {
	adapter       ExchangeAdapter
	strategy      sdk.Strategy
	symbol        string
	initialCash   float64
	
	// State management
	currentPosition *Position
	equity          float64
	orders          map[string]*Order
	mu              sync.RWMutex
	
	// Context and control
	ctx           context.Context
	cancel        context.CancelFunc
	isRunning     bool
	
	// Callbacks
	onOrder       func(*Order)
	onPosition    func(*Position)
	onError       func(error)
	
	// Kill switch
	killSwitch    *KillSwitch
}

// BridgeConfig holds configuration for the bridge
type BridgeConfig struct {
	Adapter     ExchangeAdapter
	Strategy    sdk.Strategy
	Symbol      string
	InitialCash float64
	
	// Risk limits
	MaxPositionSize float64
	MaxDrawdown     float64
	MaxDailyLoss    float64
}

// NewBridge creates a new live execution bridge
func NewBridge(ctx context.Context, config BridgeConfig) (*Bridge, error) {
	if config.Adapter == nil {
		return nil, fmt.Errorf("exchange adapter is required")
	}
	if config.Strategy == nil {
		return nil, fmt.Errorf("strategy is required")
	}
	
	execCtx, cancel := context.WithCancel(ctx)
	
	bridge := &Bridge{
		adapter:     config.Adapter,
		strategy:    config.Strategy,
		symbol:      config.Symbol,
		initialCash: config.InitialCash,
		equity:      config.InitialCash,
		orders:      make(map[string]*Order),
		ctx:         execCtx,
		cancel:      cancel,
		isRunning:   false,
	}
	
	// Initialize kill switch with risk limits
	bridge.killSwitch = NewKillSwitch(KillSwitchConfig{
		MaxDrawdown:     config.MaxDrawdown,
		MaxDailyLoss:    config.MaxDailyLoss,
		MaxPositionSize: config.MaxPositionSize,
	})
	
	return bridge, nil
}

// Start begins live trading execution
func (b *Bridge) Start() error {
	b.mu.Lock()
	if b.isRunning {
		b.mu.Unlock()
		return fmt.Errorf("bridge already running")
	}
	b.isRunning = true
	b.mu.Unlock()
	
	// Connect to exchange
	if err := b.adapter.Connect(b.ctx); err != nil {
		return fmt.Errorf("failed to connect to exchange: %w", err)
	}
	
	// Initialize strategy
	initCtx := &liveInitContext{bridge: b}
	if err := b.strategy.Init(initCtx); err != nil {
		return fmt.Errorf("strategy init failed: %w", err)
	}
	
	// Sync initial state
	if err := b.syncState(); err != nil {
		return fmt.Errorf("failed to sync initial state: %w", err)
	}
	
	log.Println("Live execution bridge started")
	
	// Start monitoring goroutine
	go b.monitor()
	
	return nil
}

// syncState synchronizes bridge state with exchange
func (b *Bridge) syncState() error {
	// Get current balance
	balance, err := b.adapter.GetBalance(b.ctx)
	if err != nil {
		return fmt.Errorf("failed to get balance: %w", err)
	}
	
	b.mu.Lock()
	b.equity = balance
	b.mu.Unlock()
	
	// Get current position
	position, err := b.adapter.GetPosition(b.ctx, b.symbol)
	if err != nil {
		log.Printf("Warning: failed to get position: %v", err)
		// Non-fatal, position might not exist
	} else {
		b.mu.Lock()
		b.currentPosition = position
		b.mu.Unlock()
	}
	
	log.Printf("State synced - Equity: %.2f, Position: %v", balance, position)
	return nil
}

// monitor runs periodic checks and kill switch evaluation
func (b *Bridge) monitor() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-b.ctx.Done():
			return
		case <-ticker.C:
			if err := b.runHealthCheck(); err != nil {
				log.Printf("Health check failed: %v", err)
				if b.onError != nil {
					b.onError(err)
				}
			}
		}
	}
}

// runHealthCheck performs periodic health checks
func (b *Bridge) runHealthCheck() error {
	// Check exchange connection
	if !b.adapter.IsConnected() {
		return fmt.Errorf("exchange connection lost")
	}
	
	// Sync state
	if err := b.syncState(); err != nil {
		return fmt.Errorf("state sync failed: %w", err)
	}
	
	// Evaluate kill switch
	b.mu.RLock()
	equity := b.equity
	position := b.currentPosition
	b.mu.RUnlock()
	
	if shouldKill, reason := b.killSwitch.Evaluate(equity, b.initialCash, position); shouldKill {
		log.Printf("KILL SWITCH TRIGGERED: %s", reason)
		if err := b.emergencyStop(reason); err != nil {
			return fmt.Errorf("emergency stop failed: %w", err)
		}
		return fmt.Errorf("kill switch triggered: %s", reason)
	}
	
	return nil
}

// emergencyStop immediately closes all positions and stops trading
func (b *Bridge) emergencyStop(reason string) error {
	log.Printf("EMERGENCY STOP: %s", reason)
	
	// Close all positions
	b.mu.RLock()
	hasPosition := b.currentPosition != nil
	b.mu.RUnlock()
	
	if hasPosition {
		if err := b.closePositionImmediate(); err != nil {
			log.Printf("Failed to close position during emergency stop: %v", err)
			// Continue with stop even if close fails
		}
	}
	
	// Stop the bridge
	b.Stop()
	
	return nil
}

// closePositionImmediate closes position with market order immediately
func (b *Bridge) closePositionImmediate() error {
	b.mu.RLock()
	pos := b.currentPosition
	b.mu.RUnlock()
	
	if pos == nil {
		return nil
	}
	
	var side OrderSide
	if pos.Side == "long" {
		side = OrderSideSell
	} else {
		side = OrderSideBuy
	}
	
	order, err := b.adapter.PlaceOrder(b.ctx, b.symbol, side, OrderTypeMarket, pos.Size, 0)
	if err != nil {
		return fmt.Errorf("failed to place close order: %w", err)
	}
	
	log.Printf("Emergency close order placed: %s", order.ID)
	return nil
}

// ProcessBar processes a new bar for live trading
func (b *Bridge) ProcessBar(ohlcv data.OHLCV) error {
	b.mu.Lock()
	if !b.isRunning {
		b.mu.Unlock()
		return fmt.Errorf("bridge not running")
	}
	b.mu.Unlock()
	
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
	barCtx := &liveBarContext{
		bridge: b,
		bar:    sdkBar,
	}
	
	// Call strategy
	if err := b.strategy.OnBar(barCtx, sdkBar); err != nil {
		return fmt.Errorf("strategy OnBar failed: %w", err)
	}
	
	return nil
}

// PlaceMarketOrder places a market order on the exchange
func (b *Bridge) PlaceMarketOrder(side string, quantity float64) error {
	var orderSide OrderSide
	if side == "long" || side == "buy" {
		orderSide = OrderSideBuy
	} else {
		orderSide = OrderSideSell
	}
	
	order, err := b.adapter.PlaceOrder(b.ctx, b.symbol, orderSide, OrderTypeMarket, quantity, 0)
	if err != nil {
		return fmt.Errorf("failed to place order: %w", err)
	}
	
	b.mu.Lock()
	b.orders[order.ID] = order
	b.mu.Unlock()
	
	log.Printf("Market order placed: %s %s %.4f @ %s", order.Side, b.symbol, quantity, order.Status)
	
	if b.onOrder != nil {
		b.onOrder(order)
	}
	
	// Wait for order fill (with timeout)
	return b.waitForOrderFill(order.ID, 30*time.Second)
}

// waitForOrderFill waits for an order to be filled
func (b *Bridge) waitForOrderFill(orderID string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(b.ctx, timeout)
	defer cancel()
	
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for order fill")
		case <-ticker.C:
			order, err := b.adapter.GetOrder(ctx, orderID)
			if err != nil {
				return fmt.Errorf("failed to get order status: %w", err)
			}
			
			b.mu.Lock()
			b.orders[orderID] = order
			b.mu.Unlock()
			
			if order.Status == OrderStatusFilled {
				log.Printf("Order filled: %s - %.4f @ %.2f", orderID, order.FilledQty, order.AvgFillPrice)
				
				// Update position after fill
				if err := b.syncState(); err != nil {
					log.Printf("Warning: failed to sync state after fill: %v", err)
				}
				
				return nil
			}
			
			if order.Status == OrderStatusFailed || order.Status == OrderStatusCancelled {
				return fmt.Errorf("order %s: %s", order.Status, order.ErrorMessage)
			}
		}
	}
}

// GetEquity returns current equity
func (b *Bridge) GetEquity() float64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.equity
}

// GetPosition returns current position
func (b *Bridge) GetPosition() *Position {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if b.currentPosition == nil {
		return nil
	}
	// Return copy
	posCopy := *b.currentPosition
	return &posCopy
}

// SetOrderCallback sets callback for order events
func (b *Bridge) SetOrderCallback(callback func(*Order)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.onOrder = callback
}

// SetPositionCallback sets callback for position updates
func (b *Bridge) SetPositionCallback(callback func(*Position)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.onPosition = callback
}

// SetErrorCallback sets callback for errors
func (b *Bridge) SetErrorCallback(callback func(error)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.onError = callback
}

// Stop stops the live execution bridge
func (b *Bridge) Stop() {
	b.mu.Lock()
	if !b.isRunning {
		b.mu.Unlock()
		return
	}
	b.isRunning = false
	b.mu.Unlock()
	
	b.cancel()
	
	if err := b.adapter.Disconnect(); err != nil {
		log.Printf("Error disconnecting from exchange: %v", err)
	}
	
	log.Println("Live execution bridge stopped")
}

// IsRunning returns whether bridge is currently running
func (b *Bridge) IsRunning() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.isRunning
}
