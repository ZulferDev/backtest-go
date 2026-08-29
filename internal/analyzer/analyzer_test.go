package analyzer

import (
	"testing"
)

func TestParseResultsString(t *testing.T) {
	jsonData := `{
		"summary": {
			"total_trades": 10,
			"win_rate_percent": 60.0,
			"total_pnl": 500.0,
			"total_return_percent": 5.0,
			"profit_factor": 2.0,
			"sharpe_ratio": 1.5,
			"sortino_ratio": 2.0,
			"max_drawdown_percent": 10.0,
			"average_win": 100.0,
			"average_loss": 50.0
		},
		"trades": [
			{"side": "long", "entry_price": 100, "exit_price": 110, "size": 1, "entry_time": 1000, "exit_time": 2000, "pnl": 10, "fee": 1}
		],
		"equity_curve": {
			"timestamps": [1000, 2000, 3000],
			"equity": [10000, 10500, 11000]
		}
	}`

	result, err := ParseResultsString(jsonData)
	if err != nil {
		t.Fatalf("Failed to parse results: %v", err)
	}

	if result.Summary.TotalTrades != 10 {
		t.Errorf("Expected 10 trades, got %d", result.Summary.TotalTrades)
	}

	if result.Summary.WinRate != 60.0 {
		t.Errorf("Expected 60%% win rate, got %.2f", result.Summary.WinRate)
	}
}

func TestToMarkdown(t *testing.T) {
	jsonData := `{
		"summary": {
			"total_trades": 5,
			"win_rate_percent": 80.0,
			"total_pnl": 1000.0,
			"total_return_percent": 10.0,
			"profit_factor": 3.0,
			"sharpe_ratio": 2.0,
			"sortino_ratio": 2.5,
			"max_drawdown_percent": 5.0,
			"average_win": 250.0,
			"average_loss": 50.0
		},
		"trades": [
			{"side": "long", "entry_price": 100, "exit_price": 125, "size": 1, "entry_time": 1000, "exit_time": 2000, "pnl": 25, "fee": 1},
			{"side": "long", "entry_price": 100, "exit_price": 95, "size": 1, "entry_time": 2000, "exit_time": 3000, "pnl": -5, "fee": 1}
		],
		"equity_curve": {
			"timestamps": [1000, 2000, 3000],
			"equity": [10000, 10500, 11000]
		}
	}`

	result, _ := ParseResultsString(jsonData)
	md := result.ToMarkdown()

	if md == "" {
		t.Error("Expected non-empty markdown output")
	}

	// Check for key sections
	expectedSections := []string{
		"# Backtest Results Analysis",
		"Performance Summary",
		"Total Trades",
		"Win Rate",
	}

	for _, section := range expectedSections {
		if !contains(md, section) {
			t.Errorf("Expected markdown to contain '%s'", section)
		}
	}
}

func TestToAIReadableFormat(t *testing.T) {
	jsonData := `{
		"summary": {
			"total_trades": 10,
			"win_rate_percent": 60.0,
			"total_pnl": 500.0,
			"total_return_percent": 5.0,
			"profit_factor": 2.0,
			"sharpe_ratio": 1.5,
			"sortino_ratio": 2.0,
			"max_drawdown_percent": 10.0,
			"average_win": 100.0,
			"average_loss": 50.0
		},
		"trades": [],
		"equity_curve": {
			"timestamps": [1000],
			"equity": [10000]
		}
	}`

	result, _ := ParseResultsString(jsonData)
	aiInput := result.ToAIReadableFormat()

	if aiInput.TradeCount != 0 {
		t.Errorf("Expected 0 trades, got %d", aiInput.TradeCount)
	}

	if aiInput.EquityPoints != 1 {
		t.Errorf("Expected 1 equity point, got %d", aiInput.EquityPoints)
	}

	if aiInput.PerformanceMetrics["win_rate"] != 60.0 {
		t.Errorf("Expected win_rate 60.0, got %v", aiInput.PerformanceMetrics["win_rate"])
	}
}

func TestHypothesisEvaluator(t *testing.T) {
	jsonData := `{
		"summary": {
			"total_trades": 50,
			"win_rate_percent": 65.0,
			"total_pnl": 1500.0,
			"total_return_percent": 15.0,
			"profit_factor": 2.5,
			"sharpe_ratio": 1.8,
			"sortino_ratio": 2.2,
			"max_drawdown_percent": 12.0,
			"average_win": 150.0,
			"average_loss": 60.0
		},
		"trades": [],
		"equity_curve": {"timestamps": [], "equity": []}
	}`

	result, _ := ParseResultsString(jsonData)
	evaluator := NewHypothesisEvaluator(result)

	// Test trend-following hypothesis
	evalResult := evaluator.EvaluateHypothesis("Trend following momentum strategy")
	
	if !evalResult.Supported {
		t.Error("Expected hypothesis to be supported")
	}

	if len(evalResult.Evidence) == 0 {
		t.Error("Expected some evidence for supported hypothesis")
	}

	// Test mean-reversion hypothesis
	evalResult2 := evaluator.EvaluateHypothesis("Mean reversion oversold bounce")
	if !evalResult2.Supported {
		t.Error("Expected mean-reversion hypothesis to be supported")
	}
}

func TestGenerateInsights(t *testing.T) {
	jsonData := `{
		"summary": {
			"total_trades": 20,
			"win_rate_percent": 75.0,
			"total_pnl": 2000.0,
			"total_return_percent": 20.0,
			"profit_factor": 4.0,
			"sharpe_ratio": 2.5,
			"sortino_ratio": 3.0,
			"max_drawdown_percent": 8.0,
			"average_win": 200.0,
			"average_loss": 50.0
		},
		"trades": [],
		"equity_curve": {"timestamps": [], "equity": []}
	}`

	result, _ := ParseResultsString(jsonData)
	evaluator := NewHypothesisEvaluator(result)
	
	insights := evaluator.GenerateInsights()
	
	// Should have insight about small sample size
	hasSampleSizeWarning := false
	for _, insight := range insights {
		if contains(insight, "Sample size") {
			hasSampleSizeWarning = true
			break
		}
	}
	
	if !hasSampleSizeWarning {
		t.Error("Expected warning about small sample size")
	}
}

func TestSuggestImprovements(t *testing.T) {
	jsonData := `{
		"summary": {
			"total_trades": 100,
			"win_rate_percent": 40.0,
			"total_pnl": -100.0,
			"total_return_percent": -1.0,
			"profit_factor": 0.8,
			"sharpe_ratio": -0.5,
			"sortino_ratio": -0.3,
			"max_drawdown_percent": 25.0,
			"average_win": 80.0,
			"average_loss": 100.0
		},
		"trades": [],
		"equity_curve": {"timestamps": [], "equity": []}
	}`

	result, _ := ParseResultsString(jsonData)
	evaluator := NewHypothesisEvaluator(result)
	
	suggestions := evaluator.SuggestImprovements()
	
	if len(suggestions) == 0 {
		t.Error("Expected suggestions for poor performing strategy")
	}

	// Should suggest improving entry timing for low win rate
	hasEntrySuggestion := false
	for _, s := range suggestions {
		if contains(s, "filter") || contains(s, "entry") {
			hasEntrySuggestion = true
			break
		}
	}
	
	if !hasEntrySuggestion {
		t.Error("Expected suggestion about improving entry timing")
	}
}

func TestResearchMemory(t *testing.T) {
	mem := NewResearchMemory("TestStrategy")
	
	if mem.StrategyName != "TestStrategy" {
		t.Errorf("Expected strategy name 'TestStrategy', got '%s'", mem.StrategyName)
	}

	// Add hypothesis
	hypID := mem.AddHypothesis("Test hypothesis", "strategy_v1.go")
	if hypID == "" {
		t.Error("Expected non-empty hypothesis ID")
	}

	// Update evaluation
	evalResult := EvaluationResult{
		Supported:       true,
		ConfidenceLevel: "high",
		Evidence:        []string{"Good win rate"},
		Recommendation:  "Proceed with testing",
	}
	
	err := mem.UpdateHypothesisEvaluation(hypID, evalResult, []string{"Learned something"})
	if err != nil {
		t.Fatalf("Failed to update hypothesis: %v", err)
	}

	// Add iteration
	metrics := SummaryMetrics{
		TotalTrades: 50,
		WinRate:     60.0,
		TotalReturn: 10.0,
	}
	mem.AddIteration("strategy_v2.go", metrics, []string{"Added RSI filter"}, "Improve entry timing")

	// Add insight
	mem.AddInsight("RSI works well in ranging markets")

	// Get summary
	summary := mem.GetSummary()
	if summary.TotalHypotheses != 1 {
		t.Errorf("Expected 1 hypothesis, got %d", summary.TotalHypotheses)
	}
	if summary.TotalIterations != 1 {
		t.Errorf("Expected 1 iteration, got %d", summary.TotalIterations)
	}
}

func TestPatternObservation(t *testing.T) {
	mem := NewResearchMemory("PatternTest")
	
	// Observe pattern multiple times
	mem.ObservePattern("Volume spike before breakout", "BTC/USDT 1h", 0.7, true)
	mem.ObservePattern("Volume spike before breakout", "ETH/USDT 1h", 0.8, true)
	
	if len(mem.Patterns) != 1 {
		t.Errorf("Expected 1 pattern (merged), got %d", len(mem.Patterns))
	}
	
	if mem.Patterns[0].Frequency != 2 {
		t.Errorf("Expected frequency 2, got %d", mem.Patterns[0].Frequency)
	}
}

func TestBuildAIFeedback(t *testing.T) {
	mem := NewResearchMemory("FeedbackTest")
	
	mem.AddHypothesis("Test hypothesis", "v1.go")
	mem.AddInsight("First insight")
	mem.AddInsight("Second insight")
	
	feedback := mem.BuildAIFeedback()
	
	if feedback == "" {
		t.Error("Expected non-empty feedback")
	}
	
	if !contains(feedback, "Research Progress Summary") {
		t.Error("Expected feedback to contain summary header")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
