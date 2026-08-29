package analyzer

import (
	"fmt"
	"strings"
)

// HypothesisEvaluator provides framework for evaluating trading hypotheses
type HypothesisEvaluator struct {
	results *BacktestResult
}

// NewHypothesisEvaluator creates a new hypothesis evaluator
func NewHypothesisEvaluator(results *BacktestResult) *HypothesisEvaluator {
	return &HypothesisEvaluator{results: results}
}

// EvaluationResult contains the evaluation of a hypothesis
type EvaluationResult struct {
	Supported       bool     `json:"supported"`
	ConfidenceLevel string   `json:"confidence_level"` // "high", "medium", "low"
	Evidence        []string `json:"evidence"`
	Contradictions  []string `json:"contradictions"`
	Recommendation  string   `json:"recommendation"`
}

// EvaluateHypothesis evaluates a trading hypothesis against backtest results
func (h *HypothesisEvaluator) EvaluateHypothesis(hypothesis string) EvaluationResult {
	var result EvaluationResult
	
	// Parse hypothesis keywords
	hypothesisLower := strings.ToLower(hypothesis)
	
	// Check for trend-following hypothesis
	isTrendFollowing := strings.Contains(hypothesisLower, "trend") || 
		strings.Contains(hypothesisLower, "momentum") ||
		strings.Contains(hypothesisLower, "breakout")
	
	// Check for mean-reversion hypothesis
	isMeanReversion := strings.Contains(hypothesisLower, "mean") ||
		strings.Contains(hypothesisLower, "reversion") ||
		strings.Contains(hypothesisLower, "oversold") ||
		strings.Contains(hypothesisLower, "overbought")
	
	// Evaluate based on metrics
	summary := h.results.Summary
	
	// Evidence collection
	if summary.WinRate > 55 {
		result.Evidence = append(result.Evidence, fmt.Sprintf("Win rate of %.1f%% suggests edge in market", summary.WinRate))
	}
	if summary.ProfitFactor > 1.5 {
		result.Evidence = append(result.Evidence, fmt.Sprintf("Profit factor of %.2f indicates profitable strategy", summary.ProfitFactor))
	}
	if summary.SharpeRatio > 1.0 {
		result.Evidence = append(result.Evidence, fmt.Sprintf("Sharpe ratio of %.2f shows good risk-adjusted returns", summary.SharpeRatio))
	}
	if summary.TotalReturn > 0 {
		result.Evidence = append(result.Evidence, fmt.Sprintf("Total return of %.2f%% is positive", summary.TotalReturn))
	}
	
	// Contradictions collection
	if summary.MaxDrawdown > 20 {
		result.Contradictions = append(result.Contradictions, fmt.Sprintf("Max drawdown of %.2f%% exceeds acceptable risk threshold", summary.MaxDrawdown))
	}
	if summary.WinRate < 40 {
		result.Contradictions = append(result.Contradictions, fmt.Sprintf("Win rate of %.1f%% is below random expectation", summary.WinRate))
	}
	if summary.ProfitFactor < 1.0 {
		result.Contradictions = append(result.Contradictions, fmt.Sprintf("Profit factor of %.2f indicates losing strategy", summary.ProfitFactor))
	}
	
	// Determine support level
	supportScore := len(result.Evidence) - len(result.Contradictions)
	
	if supportScore >= 3 {
		result.Supported = true
		result.ConfidenceLevel = "high"
		result.Recommendation = "Hypothesis strongly supported. Consider increasing position size or expanding to similar markets."
	} else if supportScore >= 1 {
		result.Supported = true
		result.ConfidenceLevel = "medium"
		result.Recommendation = "Hypothesis moderately supported. Consider refining entry/exit rules to improve consistency."
	} else if supportScore >= 0 {
		result.Supported = false
		result.ConfidenceLevel = "low"
		result.Recommendation = "Insufficient evidence. Consider gathering more data or adjusting hypothesis parameters."
	} else {
		result.Supported = false
		result.ConfidenceLevel = "low"
		result.Recommendation = "Hypothesis contradicted by data. Consider rejecting or significantly revising the hypothesis."
	}
	
	// Strategy-specific insights
	if isTrendFollowing && summary.WinRate < 50 && summary.ProfitFactor > 1.5 {
		result.Evidence = append(result.Evidence, "Pattern typical of trend-following: lower win rate but high reward/risk")
	}
	
	if isMeanReversion && summary.WinRate > 60 && summary.AverageWin < summary.AverageLoss*0.8 {
		result.Evidence = append(result.Evidence, "Pattern typical of mean-reversion: high win rate but smaller wins vs losses")
	}
	
	return result
}

// GenerateInsights generates key insights from backtest results
func (h *HypothesisEvaluator) GenerateInsights() []string {
	var insights []string
	summary := h.results.Summary
	
	if summary.TotalTrades < 30 {
		insights = append(insights, "Sample size too small for statistical significance. Run more tests.")
	}
	
	if summary.WinRate > 70 && summary.TotalTrades > 50 {
		insights = append(insights, "Exceptionally high win rate - verify no lookahead bias or overfitting")
	}
	
	if summary.MaxDrawdown > 30 {
		insights = append(insights, "Critical: Drawdown exceeds 30%. Implement stricter risk management.")
	}
	
	if summary.SharpeRatio > 2.0 {
		insights = append(insights, "Outstanding risk-adjusted returns. Verify results with out-of-sample testing.")
	}
	
	if summary.ProfitFactor > 3.0 && summary.TotalTrades < 100 {
		insights = append(insights, "Very high profit factor with limited trades - likely overfitting")
	}
	
	if len(h.results.Trades) > 0 {
		consecutiveLosses := 0
		maxConsecutiveLosses := 0
		for _, t := range h.results.Trades {
			if t.PnL < 0 {
				consecutiveLosses++
				if consecutiveLosses > maxConsecutiveLosses {
					maxConsecutiveLosses = consecutiveLosses
				}
			} else {
				consecutiveLosses = 0
			}
		}
		if maxConsecutiveLosses > 5 {
			insights = append(insights, fmt.Sprintf("Strategy experienced %d consecutive losses. Ensure adequate capital buffer.", maxConsecutiveLosses))
		}
	}
	
	return insights
}

// SuggestImprovements suggests strategy improvements based on weaknesses
func (h *HypothesisEvaluator) SuggestImprovements() []string {
	var suggestions []string
	summary := h.results.Summary
	
	if summary.WinRate < 45 {
		suggestions = append(suggestions, "Add filter conditions to improve entry timing (e.g., volume confirmation, trend alignment)")
	}
	
	if summary.MaxDrawdown > 15 {
		suggestions = append(suggestions, "Implement tighter stop-loss or reduce position sizing to limit drawdown")
	}
	
	if summary.ProfitFactor < 1.2 && summary.WinRate > 50 {
		suggestions = append(suggestions, "Let winners run longer - consider trailing stops or wider profit targets")
	}
	
	if summary.ProfitFactor > 2.0 && summary.WinRate < 40 {
		suggestions = append(suggestions, "Consider pyramiding into winning positions to maximize strong trends")
	}
	
	if summary.AverageWin < summary.AverageLoss {
		suggestions = append(suggestions, "Review exit strategy - average winner should exceed average loser")
	}
	
	if summary.TotalTrades < 50 {
		suggestions = append(suggestions, "Test on longer time period or lower timeframe for more trade samples")
	}
	
	return suggestions
}
