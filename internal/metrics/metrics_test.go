package metrics

import (
	"testing"
	"github.com/ZulferDev/backtest-go/internal/backtest"
)

func TestCalculateTotalPnL(t *testing.T) {
	trades := []backtest.Trade{
		{PnL: 100, Fee: 5},
		{PnL: -50, Fee: 5},
		{PnL: 200, Fee: 10},
	}

	expected := 100 - 5 - 50 - 5 + 200 - 10 // 230
	result := CalculateTotalPnL(trades)

	if result != float64(expected) {
		t.Errorf("Expected %f, got %f", float64(expected), result)
	}
}

func TestCalculateWinRate(t *testing.T) {
	trades := []backtest.Trade{
		{PnL: 100},
		{PnL: -50},
		{PnL: 200},
		{PnL: -30},
	}

	expected := 50.0 // 2 wins out of 4 trades
	result := CalculateWinRate(trades)

	if result != expected {
		t.Errorf("Expected %f, got %f", expected, result)
	}
}

func TestCalculateMaxDrawdown(t *testing.T) {
	curve := EquityCurve{
		Timestamps: []int64{1000, 2000, 3000, 4000, 5000},
		Equity:     []float64{10000, 12000, 9000, 11000, 13000},
	}

	// Peak at 12000, trough at 9000 -> DD = (12000-9000)/12000 * 100 = 25%
	expected := 25.0
	result := CalculateMaxDrawdown(curve)

	if result != expected {
		t.Errorf("Expected max drawdown %f%%, got %f%%", expected, result)
	}
}

func TestCalculateSharpeRatio(t *testing.T) {
	returns := []float64{0.01, 0.02, -0.01, 0.03, 0.01}
	riskFreeRate := 0.02 // 2% annual
	periodsPerYear := 252.0 // daily

	result := CalculateSharpeRatio(returns, riskFreeRate, periodsPerYear)

	// Just check it's a reasonable value (should be positive with positive avg return)
	if result < 0 {
		t.Errorf("Expected positive Sharpe ratio, got %f", result)
	}
}
