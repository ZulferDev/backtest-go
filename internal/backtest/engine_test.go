package backtest

import (
	"testing"
	"github.com/ZulferDev/backtest-go/pkg/data"
	"github.com/ZulferDev/backtest-go/pkg/sdk"
)

// Mock strategy for testing
type mockStrategy struct {
	actionOnBar func(ctx sdk.BarContext, bar sdk.OHLCV) error
}

func (m *mockStrategy) Init(ctx sdk.InitContext) error {
	return nil
}

func (m *mockStrategy) OnBar(ctx sdk.BarContext, bar sdk.OHLCV) error {
	if m.actionOnBar != nil {
		return m.actionOnBar(ctx, bar)
	}
	return nil
}

func TestEngineBasicExecution(t *testing.T) {
	// Create mock historical data
	historicalData := []data.OHLCV{
		{Timestamp: 1000, Open: 100, High: 105, Low: 95, Close: 102, Volume: 1000},
		{Timestamp: 2000, Open: 102, High: 110, Low: 100, Close: 108, Volume: 1200},
		{Timestamp: 3000, Open: 108, High: 115, Low: 105, Close: 112, Volume: 1500},
	}

	// Create a simple buy-and-hold strategy
	buyExecuted := false
	strategy := &mockStrategy{
		actionOnBar: func(ctx sdk.BarContext, bar sdk.OHLCV) error {
			if !buyExecuted && !ctx.HasOpenPosition() {
				if err := ctx.MarketBuy(1.0); err != nil {
					return err
				}
				buyExecuted = true
			}
			return nil
		},
	}

	engine := NewEngine(strategy, historicalData, 10000.0)
	err := engine.Run()
	if err != nil {
		t.Fatalf("Engine run failed: %v", err)
	}

	state := engine.GetState()
	if state.position == nil {
		t.Fatal("Expected position to be open")
	}

	if state.position.Side() != "long" {
		t.Errorf("Expected long position, got %s", state.position.Side())
	}
}

func TestEngineTradeExecution(t *testing.T) {
	historicalData := []data.OHLCV{
		{Timestamp: 1000, Open: 100, High: 105, Low: 95, Close: 102, Volume: 1000},
		{Timestamp: 2000, Open: 102, High: 110, Low: 100, Close: 108, Volume: 1200},
		{Timestamp: 3000, Open: 108, High: 115, Low: 105, Close: 112, Volume: 1500},
	}

	buyExecuted := false
	strategy := &mockStrategy{
		actionOnBar: func(ctx sdk.BarContext, bar sdk.OHLCV) error {
			if !buyExecuted && !ctx.HasOpenPosition() {
				if err := ctx.MarketBuy(1.0); err != nil {
					return err
				}
				buyExecuted = true
			} else if buyExecuted && ctx.HasOpenPosition() && bar.Timestamp == 3000 {
				// Close position on last bar
				if err := ctx.CloseAll(); err != nil {
					return err
				}
			}
			return nil
		},
	}

	engine := NewEngine(strategy, historicalData, 10000.0)
	err := engine.Run()
	if err != nil {
		t.Fatalf("Engine run failed: %v", err)
	}

	state := engine.GetState()
	if len(state.trades) != 1 {
		t.Fatalf("Expected 1 trade, got %d", len(state.trades))
	}

	trade := state.trades[0]
	if trade.Side != "long" {
		t.Errorf("Expected long trade, got %s", trade.Side)
	}

	// Entry at 102, exit at 112 -> PnL should be 10
	expectedPnL := 10.0
	if trade.PnL != expectedPnL {
		t.Errorf("Expected PnL %f, got %f", expectedPnL, trade.PnL)
	}
}
