package metrics

import (
	"math"
	"github.com/ZulferDev/backtest-go/internal/backtest"
)

// CalculateTotalPnL calculates total profit/loss from trades
func CalculateTotalPnL(trades []backtest.Trade) float64 {
	var total float64
	for _, trade := range trades {
		total += trade.PnL - trade.Fee
	}
	return total
}

// CalculateWinRate calculates percentage of winning trades
func CalculateWinRate(trades []backtest.Trade) float64 {
	if len(trades) == 0 {
		return 0
	}

	wins := 0
	for _, trade := range trades {
		if trade.PnL > 0 {
			wins++
		}
	}
	return float64(wins) / float64(len(trades)) * 100
}

// CalculateProfitFactor calculates gross profit / gross loss
func CalculateProfitFactor(trades []backtest.Trade) float64 {
	var grossProfit, grossLoss float64

	for _, trade := range trades {
		if trade.PnL > 0 {
			grossProfit += trade.PnL
		} else {
			grossLoss += math.Abs(trade.PnL)
		}
	}

	if grossLoss == 0 {
		if grossProfit > 0 {
			return math.Inf(1) // Perfect strategy
		}
		return 0
	}

	return grossProfit / grossLoss
}

// CalculateAverageWin calculates average profit from winning trades
func CalculateAverageWin(trades []backtest.Trade) float64 {
	var total float64
	count := 0

	for _, trade := range trades {
		if trade.PnL > 0 {
			total += trade.PnL
			count++
		}
	}

	if count == 0 {
		return 0
	}
	return total / float64(count)
}

// CalculateAverageLoss calculates average loss from losing trades
func CalculateAverageLoss(trades []backtest.Trade) float64 {
	var total float64
	count := 0

	for _, trade := range trades {
		if trade.PnL < 0 {
			total += math.Abs(trade.PnL)
			count++
		}
	}

	if count == 0 {
		return 0
	}
	return total / float64(count)
}
