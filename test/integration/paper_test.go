package integration

import (
	"testing"
	"time"

	"github.com/ZulferDev/backtest-go/internal/paper"
	"github.com/ZulferDev/backtest-go/pkg/data"
	"github.com/ZulferDev/backtest-go/pkg/sdk"
	"github.com/ZulferDev/backtest-go/strategies"
)

// TestPaperTradingExecution tests paper trading executor
func TestPaperTradingExecution(t *testing.T) {
	// Create paper trading executor
	executor := paper.NewSimpleExecutor(10000.0)

	// Verify initial state
	if executor.GetEquity() != 10000.0 {
		t.Errorf("Expected initial equity 10000, got %.2f", executor.GetEquity())
	}

	if executor.GetPosition() != nil {
		t.Error("Expected no initial position")
	}

	// Test market buy
	err := executor.PlaceMarketOrder("buy", 1.0, 50000.0)
	if err != nil {
		t.Fatalf("Failed to place buy order: %v", err)
	}

	pos := executor.GetPosition()
	if pos == nil {
		t.Fatal("Expected position after buy order")
	}

	if pos.Side != "long" {
		t.Errorf("Expected long position, got %s", pos.Side)
	}

	if pos.Size != 1.0 {
		t.Errorf("Expected size 1.0, got %.2f", pos.Size)
	}

	// Test close position
	err = executor.PlaceMarketOrder("sell", 1.0, 51000.0)
	if err != nil {
		t.Fatalf("Failed to close position: %v", err)
	}

	if executor.GetPosition() != nil {
		t.Error("Expected no position after close")
	}

	// Verify PnL (bought at 50000, sold at 51000, size=1.0)
	// Gross PnL = 1000, Fee = (50000 + 51000) * 1.0 * 0.001 = 101
	// Net PnL = 1000 - 101 = 899
	expectedEquity := 10000.0 + 899.0
	if executor.GetEquity() != expectedEquity {
		t.Errorf("Expected equity %.2f, got %.2f", expectedEquity, executor.GetEquity())
	}
}

// TestPaperTradingContext tests paper trading context with strategy
func TestPaperTradingContext(t *testing.T) {
	executor := paper.NewSimpleExecutor(10000.0)
	strategy := strategies.NewSMACrossover()

	// Initialize strategy
	initCtx := paper.NewInitContext(executor)
	if err := strategy.Init(initCtx); err != nil {
		t.Fatalf("Strategy init failed: %v", err)
	}

	// Simulate bars
	testData := generateTestDataPaper(100)

	for _, bar := range testData {
		ohlcv := sdk.OHLCV{
			Timestamp: bar.Timestamp,
			Open:      bar.Open,
			High:      bar.High,
			Low:       bar.Low,
			Close:     bar.Close,
			Volume:    bar.Volume,
		}

		barCtx := paper.NewBarContext(executor, ohlcv, []sdk.OHLCV{ohlcv})
		if err := strategy.OnBar(barCtx, ohlcv); err != nil {
			t.Errorf("OnBar failed at bar %d: %v", bar.Timestamp, err)
		}
	}

	t.Logf("Final equity: %.2f", executor.GetEquity())
	t.Logf("Total trades: %d", len(executor.GetTrades()))
}

// TestPaperTradingWebSocketSimulation tests WebSocket data handling
func TestPaperTradingWebSocketSimulation(t *testing.T) {
	// Note: This is a mock test since we can't connect to real WebSocket in tests
	// In production, this would test actual WebSocket connection

	executor := paper.NewSimpleExecutor(10000.0)
	strategy := strategies.NewSMACrossover()

	// Initialize
	initCtx := paper.NewInitContext(executor)
	if err := strategy.Init(initCtx); err != nil {
		t.Fatalf("Strategy init failed: %v", err)
	}

	// Simulate receiving WebSocket messages
	mockBars := []sdk.OHLCV{
		{Timestamp: time.Now().Unix(), Open: 50000, High: 50100, Low: 49900, Close: 50050, Volume: 100},
		{Timestamp: time.Now().Unix() + 60, Open: 50050, High: 50150, Low: 49950, Close: 50100, Volume: 110},
		{Timestamp: time.Now().Unix() + 120, Open: 50100, High: 50200, Low: 50000, Close: 50150, Volume: 120},
	}

	for _, bar := range mockBars {
		barCtx := paper.NewBarContext(executor, bar, []sdk.OHLCV{bar})
		if err := strategy.OnBar(barCtx, bar); err != nil {
			t.Errorf("Failed to process bar: %v", err)
		}
	}

	t.Logf("Processed %d bars successfully", len(mockBars))
}

// TestPaperTradingRiskManagement tests position sizing and risk controls
func TestPaperTradingRiskManagement(t *testing.T) {
	executor := paper.NewSimpleExecutor(10000.0)

	// Test position size limits
	tests := []struct {
		name      string
		side      string
		size      float64
		price     float64
		shouldErr bool
	}{
		{"valid_buy", "buy", 0.1, 50000.0, false},
		{"valid_sell", "sell", 0.1, 50000.0, false},
		{"zero_size", "buy", 0.0, 50000.0, true},
		{"negative_size", "buy", -0.1, 50000.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset executor state
			executor = paper.NewSimpleExecutor(10000.0)

			err := executor.PlaceMarketOrder(tt.side, tt.size, tt.price)
			hasError := err != nil

			if hasError != tt.shouldErr {
				t.Errorf("Expected error=%v, got error=%v (%v)", tt.shouldErr, hasError, err)
			}
		})
	}
}

// generateTestDataPaper creates synthetic OHLCV data
func generateTestDataPaper(bars int) []data.OHLCV {
	result := make([]data.OHLCV, bars)
	basePrice := 50000.0

	for i := 0; i < bars; i++ {
		trend := float64(i) * 10.0
		noise := float64((i*7)%20 - 10)

		close := basePrice + trend + noise
		high := close + 50
		low := close - 50
		open := close - float64((i*3)%10)

		result[i] = data.OHLCV{
			Timestamp: int64(1609459200 + i*3600),
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    1000.0 + float64(i*10),
		}
	}

	return result
}
