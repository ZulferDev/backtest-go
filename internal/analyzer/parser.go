package analyzer

import (
	"encoding/json"
	"fmt"
	"os"
)

// BacktestResult represents parsed backtest results from results.json
type BacktestResult struct {
	Summary SummaryMetrics   `json:"summary"`
	Trades  []Trade          `json:"trades"`
	Equity  EquityCurve      `json:"equity_curve"`
}

// SummaryMetrics contains aggregated performance metrics
type SummaryMetrics struct {
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

// Trade represents a single trade
type Trade struct {
	Side       string  `json:"side"`
	EntryPrice float64 `json:"entry_price"`
	ExitPrice  float64 `json:"exit_price"`
	Size       float64 `json:"size"`
	EntryTime  int64   `json:"entry_time"`
	ExitTime   int64   `json:"exit_time"`
	PnL        float64 `json:"pnl"`
	Fee        float64 `json:"fee"`
}

// EquityCurve represents equity progression over time
type EquityCurve struct {
	Timestamps []int64   `json:"timestamps"`
	Equity     []float64 `json:"equity"`
}

// ParseResultsFile loads and parses a results.json file
func ParseResultsFile(filepath string) (*BacktestResult, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read results file: %w", err)
	}

	var result BacktestResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to parse results JSON: %w", err)
	}

	return &result, nil
}

// ParseResultsString parses results from JSON string
func ParseResultsString(jsonData string) (*BacktestResult, error) {
	var result BacktestResult
	if err := json.Unmarshal([]byte(jsonData), &result); err != nil {
		return nil, fmt.Errorf("failed to parse results JSON: %w", err)
	}
	return &result, nil
}

// ToMarkdown converts results to AI-readable Markdown format
func (r *BacktestResult) ToMarkdown() string {
	md := "# Backtest Results Analysis\n\n"
	md += "## Performance Summary\n\n"
	md += fmt.Sprintf("| Metric | Value |\n")
	md += fmt.Sprintf("|--------|-------|\n")
	md += fmt.Sprintf("| Total Trades | %d |\n", r.Summary.TotalTrades)
	md += fmt.Sprintf("| Win Rate | %.2f%% |\n", r.Summary.WinRate)
	md += fmt.Sprintf("| Total PnL | $%.2f |\n", r.Summary.TotalPnL)
	md += fmt.Sprintf("| Total Return | %.2f%% |\n", r.Summary.TotalReturn)
	md += fmt.Sprintf("| Profit Factor | %.2f |\n", r.Summary.ProfitFactor)
	md += fmt.Sprintf("| Sharpe Ratio | %.3f |\n", r.Summary.SharpeRatio)
	md += fmt.Sprintf("| Sortino Ratio | %.3f |\n", r.Summary.SortinoRatio)
	md += fmt.Sprintf("| Max Drawdown | %.2f%% |\n", r.Summary.MaxDrawdown)
	md += fmt.Sprintf("| Average Win | $%.2f |\n", r.Summary.AverageWin)
	md += fmt.Sprintf("| Average Loss | $%.2f |\n", r.Summary.AverageLoss)

	if len(r.Trades) > 0 {
		md += "\n## Trade Statistics\n\n"
		winningTrades := 0
		losingTrades := 0
		var totalFees float64
		for _, t := range r.Trades {
			if t.PnL > 0 {
				winningTrades++
			} else {
				losingTrades++
			}
			totalFees += t.Fee
		}
		md += fmt.Sprintf("- Winning Trades: %d (%.1f%%)\n", winningTrades, float64(winningTrades)/float64(len(r.Trades))*100)
		md += fmt.Sprintf("- Losing Trades: %d (%.1f%%)\n", losingTrades, float64(losingTrades)/float64(len(r.Trades))*100)
		md += fmt.Sprintf("- Total Fees Paid: $%.2f\n", totalFees)
	}

	if len(r.Equity.Equity) > 0 {
		initialEquity := r.Equity.Equity[0]
		finalEquity := r.Equity.Equity[len(r.Equity.Equity)-1]
		md += fmt.Sprintf("\n## Equity Curve\n\n")
		md += fmt.Sprintf("- Starting Equity: $%.2f\n", initialEquity)
		md += fmt.Sprintf("- Ending Equity: $%.2f\n", finalEquity)
		md += fmt.Sprintf("- Data Points: %d\n", len(r.Equity.Equity))
	}

	return md
}

// ToAIReadableFormat converts results to structured AI analysis input
func (r *BacktestResult) ToAIReadableFormat() AIAnalysisInput {
	winLossRatio := r.Summary.AverageWin / maxFloat(r.Summary.AverageLoss, 0.001)
	
	return AIAnalysisInput{
		PerformanceMetrics: map[string]interface{}{
			"total_trades":     r.Summary.TotalTrades,
			"win_rate":         r.Summary.WinRate,
			"total_pnl":        r.Summary.TotalPnL,
			"total_return":     r.Summary.TotalReturn,
			"profit_factor":    r.Summary.ProfitFactor,
			"sharpe_ratio":     r.Summary.SharpeRatio,
			"sortino_ratio":    r.Summary.SortinoRatio,
			"max_drawdown":     r.Summary.MaxDrawdown,
			"average_win":      r.Summary.AverageWin,
			"average_loss":     r.Summary.AverageLoss,
			"win_loss_ratio":   winLossRatio,
		},
		TradeCount:       len(r.Trades),
		EquityPoints:     len(r.Equity.Equity),
		RawSummary:       r.Summary,
		RawTrades:        r.Trades,
		RawEquityCurve:   r.Equity,
	}
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// AIAnalysisInput is the structured input for AI analysis
type AIAnalysisInput struct {
	PerformanceMetrics map[string]interface{} `json:"performance_metrics"`
	TradeCount         int                    `json:"trade_count"`
	EquityPoints       int                    `json:"equity_points"`
	RawSummary         SummaryMetrics         `json:"raw_summary"`
	RawTrades          []Trade                `json:"raw_trades"`
	RawEquityCurve     EquityCurve            `json:"raw_equity_curve"`
}
