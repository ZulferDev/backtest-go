package feedback

import (
	"testing"
	"time"
)

func TestNewStructuredFeedback(t *testing.T) {
	feedback := NewStructuredFeedback("test_strategy", "iter_001", "BACKTEST")

	if feedback.StrategyID != "test_strategy" {
		t.Errorf("Expected strategy_id 'test_strategy', got '%s'", feedback.StrategyID)
	}

	if feedback.IterationID != "iter_001" {
		t.Errorf("Expected iteration_id 'iter_001', got '%s'", feedback.IterationID)
	}

	if feedback.Phase != "BACKTEST" {
		t.Errorf("Expected phase 'BACKTEST', got '%s'", feedback.Phase)
	}

	if len(feedback.IdentifiedIssues) != 0 {
		t.Errorf("Expected empty issues slice, got %d items", len(feedback.IdentifiedIssues))
	}
}

func TestAddIssue(t *testing.T) {
	feedback := NewStructuredFeedback("test", "iter_001", "BACKTEST")
	
	feedback.AddIssue("high", "performance", "Low Sharpe ratio", "Sharpe = 0.5", "Increase position sizing")

	if len(feedback.IdentifiedIssues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(feedback.IdentifiedIssues))
	}

	issue := feedback.IdentifiedIssues[0]
	if issue.Severity != "high" {
		t.Errorf("Expected severity 'high', got '%s'", issue.Severity)
	}
	if issue.Category != "performance" {
		t.Errorf("Expected category 'performance', got '%s'", issue.Category)
	}
}

func TestSetBacktestResults(t *testing.T) {
	feedback := NewStructuredFeedback("test", "iter_001", "BACKTEST")
	
	metrics := map[string]float64{
		"total_return":  45.5,
		"sharpe_ratio":  1.8,
		"max_drawdown":  12.3,
		"win_rate":      58.5,
		"total_trades":  120,
		"profit_factor": 2.1,
		"sortino_ratio": 2.3,
		"average_win":   150.0,
		"average_loss":  80.0,
	}

	feedback.SetBacktestResults(metrics)

	if feedback.BacktestResults == nil {
		t.Fatal("BacktestResults should not be nil")
	}

	if feedback.BacktestResults.TotalReturn != 45.5 {
		t.Errorf("Expected total_return 45.5, got %.2f", feedback.BacktestResults.TotalReturn)
	}

	if feedback.BacktestResults.SharpeRatio != 1.8 {
		t.Errorf("Expected sharpe_ratio 1.8, got %.2f", feedback.BacktestResults.SharpeRatio)
	}

	if feedback.BacktestResults.TotalTrades != 120 {
		t.Errorf("Expected total_trades 120, got %d", feedback.BacktestResults.TotalTrades)
	}
}

func TestClassifyPerformance(t *testing.T) {
	tests := []struct {
		name     string
		metrics  map[string]float64
		expected string
	}{
		{
			name: "excellent performance",
			metrics: map[string]float64{
				"total_return": 60.0,
				"sharpe_ratio": 2.5,
				"max_drawdown": 10.0,
			},
			expected: "excellent",
		},
		{
			name: "good performance",
			metrics: map[string]float64{
				"total_return": 30.0,
				"sharpe_ratio": 1.7,
				"max_drawdown": 18.0,
			},
			expected: "good",
		},
		{
			name: "acceptable performance",
			metrics: map[string]float64{
				"total_return": 15.0,
				"sharpe_ratio": 1.2,
				"max_drawdown": 22.0,
			},
			expected: "acceptable",
		},
		{
			name: "poor performance",
			metrics: map[string]float64{
				"total_return": 5.0,
				"sharpe_ratio": 0.8,
				"max_drawdown": 28.0,
			},
			expected: "poor",
		},
		{
			name: "failing performance",
			metrics: map[string]float64{
				"total_return": -10.0,
				"sharpe_ratio": -0.5,
				"max_drawdown": 35.0,
			},
			expected: "failing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			feedback := NewStructuredFeedback("test", "iter_001", "BACKTEST")
			feedback.SetBacktestResults(tt.metrics)

			if feedback.BacktestResults.PerformanceTag != tt.expected {
				t.Errorf("Expected performance tag '%s', got '%s'", tt.expected, feedback.BacktestResults.PerformanceTag)
			}
		})
	}
}

func TestSetHypothesisValidation(t *testing.T) {
	feedback := NewStructuredFeedback("test", "iter_001", "ANALYZE")
	
	evidence := []string{"High Sharpe ratio", "Consistent profits"}
	contradictions := []string{"High drawdown"}

	feedback.SetHypothesisValidation(true, evidence, contradictions, "high")

	if feedback.HypothesisCheck == nil {
		t.Fatal("HypothesisCheck should not be nil")
	}

	if !feedback.HypothesisCheck.ThesisConfirmed {
		t.Error("Expected thesis to be confirmed")
	}

	if feedback.HypothesisCheck.ConfidenceLevel != "high" {
		t.Errorf("Expected confidence 'high', got '%s'", feedback.HypothesisCheck.ConfidenceLevel)
	}

	expectedScore := 1 // 2 evidence - 1 contradiction
	if feedback.HypothesisCheck.SupportScore != expectedScore {
		t.Errorf("Expected support score %d, got %d", expectedScore, feedback.HypothesisCheck.SupportScore)
	}
}

func TestSetNextAction(t *testing.T) {
	feedback := NewStructuredFeedback("test", "iter_001", "ANALYZE")
	
	tasks := []string{"Add ADX filter", "Adjust position sizing"}
	feedback.SetNextAction("refine", "Risk management", "High drawdown detected", tasks, "Reduce drawdown by 5%")

	if feedback.NextAction == nil {
		t.Fatal("NextAction should not be nil")
	}

	if feedback.NextAction.Action != "refine" {
		t.Errorf("Expected action 'refine', got '%s'", feedback.NextAction.Action)
	}

	if len(feedback.NextAction.SpecificTasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(feedback.NextAction.SpecificTasks))
	}
}

func TestToJSON(t *testing.T) {
	feedback := NewStructuredFeedback("test_strategy", "iter_001", "BACKTEST")
	feedback.Status = "success"
	
	metrics := map[string]float64{
		"total_return": 25.5,
		"sharpe_ratio": 1.5,
		"max_drawdown": 15.0,
	}
	feedback.SetBacktestResults(metrics)

	jsonData, err := feedback.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize to JSON: %v", err)
	}

	if len(jsonData) == 0 {
		t.Error("JSON data should not be empty")
	}

	// Deserialize back
	parsed, err := FromJSON(jsonData)
	if err != nil {
		t.Fatalf("Failed to deserialize JSON: %v", err)
	}

	if parsed.StrategyID != "test_strategy" {
		t.Errorf("Expected strategy_id 'test_strategy', got '%s'", parsed.StrategyID)
	}

	if parsed.BacktestResults.TotalReturn != 25.5 {
		t.Errorf("Expected total_return 25.5, got %.2f", parsed.BacktestResults.TotalReturn)
	}
}

func TestToAIPromptFormat(t *testing.T) {
	feedback := NewStructuredFeedback("rsi_mean_reversion", "iter_002", "ANALYZE")
	feedback.Status = "success"

	metrics := map[string]float64{
		"total_return":  32.5,
		"sharpe_ratio":  1.75,
		"max_drawdown":  14.2,
		"win_rate":      62.5,
		"total_trades":  85,
		"profit_factor": 2.1,
		"sortino_ratio": 2.0,
		"average_win":   180.0,
		"average_loss":  95.0,
	}
	feedback.SetBacktestResults(metrics)

	evidence := []string{"Strong Sharpe ratio", "Good win rate"}
	contradictions := []string{}
	feedback.SetHypothesisValidation(true, evidence, contradictions, "high")

	feedback.AddIssue("medium", "risk", "Drawdown slightly elevated", "Max DD = 14.2%", "Add trailing stop")

	tasks := []string{"Implement trailing stop", "Test on different timeframes"}
	feedback.SetNextAction("refine", "Risk optimization", "Further reduce drawdown", tasks, "Target max DD < 12%")

	prompt := feedback.ToAIPromptFormat()

	if len(prompt) == 0 {
		t.Error("Prompt should not be empty")
	}

	// Check key sections exist
	expectedSections := []string{
		"Strategy Analysis Feedback",
		"Context",
		"Backtest Performance",
		"Hypothesis Validation",
		"Identified Issues",
		"Next Action",
	}

	for _, section := range expectedSections {
		if !contains(prompt, section) {
			t.Errorf("Prompt should contain section: %s", section)
		}
	}
}

func TestLearningContext(t *testing.T) {
	feedback := NewStructuredFeedback("test", "iter_003", "ANALYZE")

	ctx := &LearningContext{
		PreviousIterations:   2,
		ImprovementTrend:     "improving",
		CurrentVersion:       "v3",
		LearnedPatterns:      []string{"RSI works best in ranging markets"},
		FailedApproaches:     []string{"Fixed stop loss too tight"},
		SuccessfulTechniques: []string{"ATR-based position sizing"},
		BestPerformance: map[string]float64{
			"sharpe_ratio": 2.1,
		},
		WorstPerformance: map[string]float64{
			"sharpe_ratio": 0.8,
		},
	}

	feedback.SetLearningContext(ctx)

	if feedback.LearningContext == nil {
		t.Fatal("LearningContext should not be nil")
	}

	if feedback.LearningContext.PreviousIterations != 2 {
		t.Errorf("Expected 2 previous iterations, got %d", feedback.LearningContext.PreviousIterations)
	}

	if len(feedback.LearningContext.LearnedPatterns) != 1 {
		t.Errorf("Expected 1 learned pattern, got %d", len(feedback.LearningContext.LearnedPatterns))
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr)))
}
