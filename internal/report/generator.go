package report

import (
	"encoding/json"
	"fmt"
	"os"
	"github.com/ZulferDev/backtest-go/internal/backtest"
	"github.com/ZulferDev/backtest-go/internal/metrics"
)

// Report contains all backtest results and metrics
type Report struct {
	Summary Summary                `json:"summary"`
	Trades  []backtest.Trade       `json:"trades"`
	Equity  metrics.EquityCurve    `json:"equity_curve"`
}

// Summary contains aggregated metrics
type Summary struct {
	TotalTrades    int     `json:"total_trades"`
	WinRate        float64 `json:"win_rate_percent"`
	TotalPnL       float64 `json:"total_pnl"`
	TotalReturn    float64 `json:"total_return_percent"`
	ProfitFactor   float64 `json:"profit_factor"`
	SharpeRatio    float64 `json:"sharpe_ratio"`
	SortinoRatio   float64 `json:"sortino_ratio"`
	MaxDrawdown    float64 `json:"max_drawdown_percent"`
	AverageWin     float64 `json:"average_win"`
	AverageLoss    float64 `json:"average_loss"`
}

// GenerateReport creates a complete report from backtest state
func GenerateReport(state *backtest.State, equityCurve metrics.EquityCurve) Report {
	trades := state.Trades()
	initialCash := state.InitialCash()
	finalEquity := state.Equity()

	returns := metrics.CalculateReturns(equityCurve)

	summary := Summary{
		TotalTrades:  len(trades),
		WinRate:      metrics.CalculateWinRate(trades),
		TotalPnL:     metrics.CalculateTotalPnL(trades),
		TotalReturn:  metrics.CalculateTotalReturn(initialCash, finalEquity),
		ProfitFactor: metrics.CalculateProfitFactor(trades),
		SharpeRatio:  metrics.CalculateSharpeRatio(returns, 0.02, 252), // Assume daily, 2% risk-free
		SortinoRatio: metrics.CalculateSortinoRatio(returns, 0, 252),
		MaxDrawdown:  metrics.CalculateMaxDrawdown(equityCurve),
		AverageWin:   metrics.CalculateAverageWin(trades),
		AverageLoss:  metrics.CalculateAverageLoss(trades),
	}

	return Report{
		Summary: summary,
		Trades:  trades,
		Equity:  equityCurve,
	}
}

// SaveJSON saves report as JSON file
func (r *Report) SaveJSON(filepath string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(r)
}
