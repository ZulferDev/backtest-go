package paper

import (
	"context"
	"testing"
	"time"

	"github.com/ZulferDev/backtest-go/pkg/sdk"
)

// TestWebSocketClient tests WebSocket connection and message parsing
func TestWebSocketClient(t *testing.T) {
	t.Skip("Integration test - requires live Binance WebSocket")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	ws, err := NewWebSocketClient(ctx, "btcusdt", "1m")
	if err != nil {
		t.Fatalf("Failed to create WebSocket client: %v", err)
	}
	defer ws.Close()

	ch := ws.Subscribe()

	select {
	case ohlcv := <-ch:
		if ohlcv.Timestamp == 0 {
			t.Error("Received OHLCV with zero timestamp")
		}
		if ohlcv.Close == 0 {
			t.Error("Received OHLCV with zero close price")
		}
		t.Logf("Received OHLCV: %+v", ohlcv)
	case <-time.After(20 * time.Second):
		t.Fatal("Timeout waiting for OHLCV data")
	}
}

// TestExecutorLifecycle tests paper trading executor lifecycle
func TestExecutorLifecycle(t *testing.T) {
	t.Skip("Integration test - requires live Binance WebSocket")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	strategy := &testStrategy{t: t}
	executor, err := NewExecutor(ctx, strategy, "btcusdt", "1m", 10000.0)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	if err := executor.Start(); err != nil {
		t.Fatalf("Failed to start executor: %v", err)
	}

	if !executor.IsRunning() {
		t.Error("Executor should be running after Start()")
	}

	// Wait for at least one bar
	time.Sleep(10 * time.Second)

	state := executor.GetState()
	if state.BarCount == 0 {
		t.Error("No bars processed")
	}

	executor.Stop()

	if executor.IsRunning() {
		t.Error("Executor should not be running after Stop()")
	}
}

// TestExecutorTrading tests paper trading execution
func TestExecutorTrading(t *testing.T) {
	t.Skip("Integration test - requires live Binance WebSocket")

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	tradeChan := make(chan Trade, 10)
	strategy := &tradingStrategy{t: t, tradeCount: 0}
	
	executor, err := NewExecutor(ctx, strategy, "btcusdt", "1m", 10000.0)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	executor.SetTradeCallback(func(trade Trade) {
		tradeChan <- trade
	})

	if err := executor.Start(); err != nil {
		t.Fatalf("Failed to start executor: %v", err)
	}

	// Wait for trades
	select {
	case trade := <-tradeChan:
		t.Logf("Trade executed: %+v", trade)
		if trade.Size == 0 {
			t.Error("Trade size is zero")
		}
	case <-time.After(90 * time.Second):
		t.Log("No trades executed within timeout (might be expected)")
	}

	executor.Stop()
}

// TestPositionManagement tests position opening and closing
func TestPositionManagement(t *testing.T) {
	executor := &Executor{
		state: &TradingState{
			Position:    nil,
			Equity:      10000.0,
			InitialCash: 10000.0,
			Trades:      []Trade{},
		},
	}

	// Test opening position
	err := executor.OpenPosition("long", 0.1, 50000.0)
	if err != nil {
		t.Fatalf("Failed to open position: %v", err)
	}

	if executor.state.Position == nil {
		t.Fatal("Position should not be nil after opening")
	}

	if executor.state.Position.Side != "long" {
		t.Errorf("Expected side 'long', got '%s'", executor.state.Position.Side)
	}

	if executor.state.Position.Size != 0.1 {
		t.Errorf("Expected size 0.1, got %.4f", executor.state.Position.Size)
	}

	// Test closing position
	executor.state.CurrentBar.Timestamp = time.Now().Unix()
	err = executor.ClosePosition(51000.0)
	if err != nil {
		t.Fatalf("Failed to close position: %v", err)
	}

	if executor.state.Position != nil {
		t.Error("Position should be nil after closing")
	}

	if len(executor.state.Trades) != 1 {
		t.Errorf("Expected 1 trade, got %d", len(executor.state.Trades))
	}

	trade := executor.state.Trades[0]
	// Account for fees
	if trade.PnL <= 0 {
		t.Errorf("Expected positive PnL, got %.2f", trade.PnL)
	}
}

// TestUnrealizedPnL tests unrealized PnL calculation
func TestUnrealizedPnL(t *testing.T) {
	executor := &Executor{
		state: &TradingState{
			Position: &Position{
				Side:       "long",
				Size:       0.1,
				EntryPrice: 50000.0,
			},
		},
	}

	unrealizedPnL := executor.calculateUnrealizedPnL(51000.0)
	expectedPnL := (51000.0 - 50000.0) * 0.1
	if unrealizedPnL != expectedPnL {
		t.Errorf("Expected unrealized PnL %.2f, got %.2f", expectedPnL, unrealizedPnL)
	}

	// Test short position
	executor.state.Position.Side = "short"
	unrealizedPnL = executor.calculateUnrealizedPnL(51000.0)
	expectedPnL = (50000.0 - 51000.0) * 0.1
	if unrealizedPnL != expectedPnL {
		t.Errorf("Expected unrealized PnL %.2f, got %.2f", expectedPnL, unrealizedPnL)
	}
}

// testStrategy is a minimal test strategy
type testStrategy struct {
	t *testing.T
}

func (s *testStrategy) Init(ctx sdk.InitContext) error {
	s.t.Log("Test strategy initialized")
	return nil
}

func (s *testStrategy) OnBar(ctx sdk.BarContext, bar sdk.OHLCV) error {
	s.t.Logf("OnBar called: Close=%.2f", bar.Close)
	return nil
}

// tradingStrategy executes simple trades for testing
type tradingStrategy struct {
	t          *testing.T
	tradeCount int
}

func (s *tradingStrategy) Init(ctx sdk.InitContext) error {
	s.t.Log("Trading strategy initialized")
	return nil
}

func (s *tradingStrategy) OnBar(ctx sdk.BarContext, bar sdk.OHLCV) error {
	// Simple strategy: buy on first bar, sell on second
	if s.tradeCount == 0 && !ctx.HasOpenPosition() {
		s.t.Log("Opening long position")
		if err := ctx.MarketBuy(0.01); err != nil {
			s.t.Logf("MarketBuy error: %v", err)
		}
		s.tradeCount++
	} else if s.tradeCount == 1 && ctx.HasOpenPosition() {
		s.t.Log("Closing position")
		if err := ctx.CloseAll(); err != nil {
			s.t.Logf("CloseAll error: %v", err)
		}
		s.tradeCount++
	}
	return nil
}
