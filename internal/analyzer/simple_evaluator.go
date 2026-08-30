package analyzer

// ResultParser provides parsing and formatting of backtest results
type ResultParser struct{}

// NewResultParser creates a new result parser
func NewResultParser() *ResultParser {
	return &ResultParser{}
}

// ParseBacktestResult parses backtest result map into summary
func (p *ResultParser) ParseBacktestResult(result map[string]interface{}) *SummaryMetrics {
	if result == nil {
		return nil
	}

	summary := &SummaryMetrics{}

	if v, ok := result["total_return"].(float64); ok {
		summary.TotalReturn = v
	}
	if v, ok := result["sharpe_ratio"].(float64); ok {
		summary.SharpeRatio = v
	}
	if v, ok := result["max_drawdown"].(float64); ok {
		summary.MaxDrawdown = v
	}
	if v, ok := result["num_trades"].(int); ok {
		summary.TotalTrades = v
	}
	if v, ok := result["final_equity"].(float64); ok {
		_ = v // Use final_equity if needed in future calculations
		// Calculate win rate if we have trades data
	}
	if v, ok := result["profit_factor"].(float64); ok {
		summary.ProfitFactor = v
	}

	return summary
}

// Evaluator evaluates strategy hypotheses
type Evaluator struct{}

// NewEvaluator creates a new evaluator
func NewEvaluator() *Evaluator {
	return &Evaluator{}
}

// EvaluateHypothesis evaluates a hypothesis against backtest summary
func (e *Evaluator) EvaluateHypothesis(hypothesis string, summary *SummaryMetrics) string {
	if summary == nil {
		return "Cannot evaluate: no summary data"
	}

	evaluation := "Hypothesis Evaluation:\n\n"
	
	// Performance assessment
	if summary.TotalReturn > 0 {
		evaluation += "✓ Positive return achieved\n"
	} else {
		evaluation += "✗ Negative return\n"
	}

	if summary.SharpeRatio > 1.0 {
		evaluation += "✓ Good risk-adjusted returns (Sharpe > 1.0)\n"
	} else if summary.SharpeRatio > 0 {
		evaluation += "○ Moderate risk-adjusted returns\n"
	} else {
		evaluation += "✗ Poor risk-adjusted returns\n"
	}

	if summary.MaxDrawdown < 0.15 {
		evaluation += "✓ Acceptable drawdown (< 15%)\n"
	} else if summary.MaxDrawdown < 0.30 {
		evaluation += "○ Moderate drawdown (15-30%)\n"
	} else {
		evaluation += "✗ High drawdown (> 30%)\n"
	}

	if summary.TotalTrades > 30 {
		evaluation += "✓ Sufficient trade sample\n"
	} else {
		evaluation += "○ Limited trade sample (statistical significance uncertain)\n"
	}

	// Overall verdict
	score := 0
	if summary.TotalReturn > 0 {
		score++
	}
	if summary.SharpeRatio > 1.0 {
		score++
	}
	if summary.MaxDrawdown < 0.15 {
		score++
	}
	if summary.TotalTrades > 30 {
		score++
	}

	evaluation += "\nVerdict: "
	if score >= 3 {
		evaluation += "Hypothesis SUPPORTED - Strategy shows promise"
	} else if score >= 2 {
		evaluation += "Hypothesis PARTIALLY SUPPORTED - Needs refinement"
	} else {
		evaluation += "Hypothesis NOT SUPPORTED - Significant revision needed"
	}

	return evaluation
}
