package backtest

import (
	"math"
	"testing"

	"github.com/ZulferDev/backtest-go/internal/execution"
	"github.com/ZulferDev/backtest-go/pkg/data"
	"github.com/ZulferDev/backtest-go/pkg/sdk"
)

const epsilon = 1e-6 // Tolerance for floating point comparison

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < epsilon
}

// Simple test strategy for fees/slippage testing
type SimpleTestStrategy struct {
	barCount int
}

func (s *SimpleTestStrategy) Init(ctx sdk.InitContext) error {
	s.barCount = 0
	return nil
}

func (s *SimpleTestStrategy) OnBar(ctx sdk.BarContext, bar sdk.OHLCV) error {
	s.barCount++
	
	// Buy on first bar
	if s.barCount == 1 && !ctx.HasOpenPosition() {
		return ctx.MarketBuy(1.0)
	}
	
	// Sell on second bar
	if s.barCount == 2 && ctx.HasOpenPosition() {
		return ctx.CloseAll()
	}
	
	return nil
}

func TestEngineWithFeesAndSlippage(t *testing.T) {
	// Create test data
	testData := []data.OHLCV{
		{Timestamp: 1000, Open: 50000, High: 50100, Low: 49900, Close: 50000, Volume: 100},
		{Timestamp: 2000, Open: 50000, High: 51100, Low: 49900, Close: 51000, Volume: 100},
		{Timestamp: 3000, Open: 51000, High: 51100, Low: 50900, Close: 51000, Volume: 100},
	}

	strategy := &SimpleTestStrategy{}
	
	// Test 1: No fees/slippage (default)
	t.Run("No Fees/Slippage", func(t *testing.T) {
		engine := NewEngine(strategy, testData, 10000.0)
		if err := engine.Run(); err != nil {
			t.Fatalf("Engine run failed: %v", err)
		}
		
		state := engine.GetState()
		trades := state.Trades()
		
		if len(trades) != 1 {
			t.Fatalf("Expected 1 trade, got %d", len(trades))
		}
		
		trade := trades[0]
		// Buy at 50000, sell at 51000, size 1.0
		expectedPnL := (51000.0 - 50000.0) * 1.0 // = 1000
		
		if trade.PnL != expectedPnL {
			t.Errorf("Expected PnL %.2f, got %.2f", expectedPnL, trade.PnL)
		}
		
		if trade.Fee != 0 {
			t.Errorf("Expected no fee, got %.2f", trade.Fee)
		}
	})
	
	// Test 2: With realistic fees and slippage
	t.Run("With Fees and Slippage", func(t *testing.T) {
		config := EngineConfig{
			FeeModel:       execution.NewFixedPercentageFee(0.001), // 0.1%
			SlippageModel:  execution.NewFixedSlippage(10.0),       // 10 bps
			EnableFees:     true,
			EnableSlippage: true,
		}
		
		engine := NewEngineWithConfig(strategy, testData, 10000.0, config)
		if err := engine.Run(); err != nil {
			t.Fatalf("Engine run failed: %v", err)
		}
		
		state := engine.GetState()
		trades := state.Trades()
		
		if len(trades) != 1 {
			t.Fatalf("Expected 1 trade, got %d", len(trades))
		}
		
		trade := trades[0]
		
		// Entry: 50000 * 1.001 (slippage) = 50050
		// Entry fee: 50050 * 0.001 = 50.05
		// Exit: 51000 * 0.999 (slippage) = 50949
		// Exit fee: 50949 * 0.001 = 50.949
		// PnL = (50949 - 50050) - (50.05 + 50.949) = 899 - 100.999 = 798.001
		
		expectedEntryPrice := 50050.0
		expectedExitPrice := 50949.0
		expectedTotalFee := 100.999
		expectedPnL := (expectedExitPrice - expectedEntryPrice) - expectedTotalFee
		
		if !almostEqual(trade.EntryPrice, expectedEntryPrice) {
			t.Errorf("Expected entry price %.2f, got %.2f", expectedEntryPrice, trade.EntryPrice)
		}
		
		if !almostEqual(trade.ExitPrice, expectedExitPrice) {
			t.Errorf("Expected exit price %.2f, got %.2f", expectedExitPrice, trade.ExitPrice)
		}
		
		if !almostEqual(trade.Fee, expectedTotalFee) {
			t.Errorf("Expected total fee %.3f, got %.3f", expectedTotalFee, trade.Fee)
		}
		
		if !almostEqual(trade.PnL, expectedPnL) {
			t.Errorf("Expected PnL %.3f, got %.3f", expectedPnL, trade.PnL)
		}
		
		t.Logf("Trade details:")
		t.Logf("  Entry: %.2f (with slippage)", trade.EntryPrice)
		t.Logf("  Exit: %.2f (with slippage)", trade.ExitPrice)
		t.Logf("  Fee: %.3f", trade.Fee)
		t.Logf("  PnL: %.3f", trade.PnL)
	})
	
	// Test 3: Binance preset
	t.Run("Binance Preset", func(t *testing.T) {
		config := EngineConfig{
			FeeModel:       execution.BinanceSpotFeeModel(),
			SlippageModel:  execution.DefaultSlippageModel(),
			EnableFees:     true,
			EnableSlippage: true,
		}
		
		engine := NewEngineWithConfig(strategy, testData, 10000.0, config)
		if err := engine.Run(); err != nil {
			t.Fatalf("Engine run failed: %v", err)
		}
		
		state := engine.GetState()
		trades := state.Trades()
		
		if len(trades) != 1 {
			t.Fatalf("Expected 1 trade, got %d", len(trades))
		}
		
		trade := trades[0]
		
		t.Logf("Binance preset trade:")
		t.Logf("  Entry: %.2f", trade.EntryPrice)
		t.Logf("  Exit: %.2f", trade.ExitPrice)
		t.Logf("  Fee: %.2f", trade.Fee)
		t.Logf("  PnL: %.2f", trade.PnL)
		
		// Verify fees are applied
		if trade.Fee <= 0 {
			t.Error("Expected positive fee with Binance preset")
		}
		
		// Verify slippage is applied
		if trade.EntryPrice == 50000.0 {
			t.Error("Expected entry price to differ from bar close due to slippage")
		}
	})
}

func TestVolumeBasedSlippage(t *testing.T) {
	testData := []data.OHLCV{
		{Timestamp: 1000, Open: 50000, High: 50100, Low: 49900, Close: 50000, Volume: 100},
		{Timestamp: 2000, Open: 50000, High: 51100, Low: 49900, Close: 51000, Volume: 100},
		{Timestamp: 3000, Open: 51000, High: 51100, Low: 50900, Close: 51000, Volume: 100},
	}
	
	strategy := &SimpleTestStrategy{}
	
	// Volume-based slippage: larger orders pay more slippage
	config := EngineConfig{
		FeeModel:       execution.NewFixedPercentageFee(0.001),
		SlippageModel:  execution.NewVolumeBasedSlippage(5.0, 2.0, 1.0),
		EnableFees:     true,
		EnableSlippage: true,
	}
	
	engine := NewEngineWithConfig(strategy, testData, 10000.0, config)
	if err := engine.Run(); err != nil {
		t.Fatalf("Engine run failed: %v", err)
	}
	
	state := engine.GetState()
	trades := state.Trades()
	
	if len(trades) != 1 {
		t.Fatalf("Expected 1 trade, got %d", len(trades))
	}
	
	trade := trades[0]
	
	t.Logf("Volume-based slippage trade:")
	t.Logf("  Entry: %.2f", trade.EntryPrice)
	t.Logf("  Exit: %.2f", trade.ExitPrice)
	t.Logf("  Fee: %.2f", trade.Fee)
	t.Logf("  PnL: %.2f", trade.PnL)
	
	// Verify slippage increases with volume
	// Base 5 bps + (2 bps * 1.0/1.0) = 7 bps total
	expectedEntryPrice := 50000.0 * 1.0007 // 50035
	
	if !almostEqual(trade.EntryPrice, expectedEntryPrice) {
		t.Errorf("Expected entry price %.2f with volume slippage, got %.2f", expectedEntryPrice, trade.EntryPrice)
	}
}
