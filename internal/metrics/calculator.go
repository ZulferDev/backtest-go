package metrics

import (
	"github.com/ZulferDev/backtest-go/internal/backtest"
)

// Calculator provides unified metrics calculation interface
type Calculator struct {
	trades      []backtest.Trade
	initialCash float64
}

// NewCalculator creates a new metrics calculator
func NewCalculator(trades []backtest.Trade, initialCash float64) *Calculator {
	return &Calculator{
		trades:      trades,
		initialCash: initialCash,
	}
}

// TotalReturn calculates total return percentage
func (c *Calculator) TotalReturn(finalEquity float64) float64 {
	if c.initialCash == 0 {
		return 0
	}
	return (finalEquity - c.initialCash) / c.initialCash
}

// SharpeRatio calculates Sharpe ratio from trades
func (c *Calculator) SharpeRatio() float64 {
	if len(c.trades) == 0 {
		return 0
	}

	// Calculate returns per trade
	returns := make([]float64, len(c.trades))
	for i, t := range c.trades {
		returns[i] = t.PnL / c.initialCash
	}

	// Annualize assuming daily trades (252 trading days)
	return CalculateSharpeRatio(returns, 0.0, 252.0)
}

// MaxDrawdown calculates maximum drawdown from trades
func (c *Calculator) MaxDrawdown() float64 {
	if len(c.trades) == 0 {
		return 0
	}

	// Build equity curve
	equity := c.initialCash
	equityValues := []float64{equity}
	timestamps := []int64{0}

	for _, t := range c.trades {
		equity += t.PnL
		equityValues = append(equityValues, equity)
		timestamps = append(timestamps, t.ExitTime)
	}

	curve := EquityCurve{
		Timestamps: timestamps,
		Equity:     equityValues,
	}

	return CalculateMaxDrawdown(curve)
}

// ProfitFactor calculates profit factor (gross profit / gross loss)
func (c *Calculator) ProfitFactor() float64 {
	if len(c.trades) == 0 {
		return 0
	}

	var grossProfit, grossLoss float64
	for _, t := range c.trades {
		if t.PnL > 0 {
			grossProfit += t.PnL
		} else {
			grossLoss += -t.PnL
		}
	}

	if grossLoss == 0 {
		if grossProfit > 0 {
			return 999.99 // Infinite profit factor, cap it
		}
		return 0
	}

	return grossProfit / grossLoss
}

// WinRate calculates win rate percentage
func (c *Calculator) WinRate() float64 {
	if len(c.trades) == 0 {
		return 0
	}

	wins := 0
	for _, t := range c.trades {
		if t.PnL > 0 {
			wins++
		}
	}

	return float64(wins) / float64(len(c.trades)) * 100.0
}

// SortinoRatio calculates Sortino ratio (uses downside deviation)
func (c *Calculator) SortinoRatio() float64 {
	if len(c.trades) == 0 {
		return 0
	}

	// Calculate returns per trade
	returns := make([]float64, len(c.trades))
	for i, t := range c.trades {
		returns[i] = t.PnL / c.initialCash
	}

	// Annualize assuming daily trades
	return CalculateSortinoRatio(returns, 0.0, 252.0)
}
