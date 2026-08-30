package db

import (
	"os"
	"testing"
)

func TestNewResearchDB(t *testing.T) {
	dbPath := ":memory:"
	db, err := NewResearchDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	if db.db == nil {
		t.Error("Database connection should not be nil")
	}
}

func TestCreateStrategy(t *testing.T) {
	db, err := NewResearchDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	err = db.CreateStrategy("rsi_strategy", "RSI Mean Reversion", "ranging")
	if err != nil {
		t.Errorf("Failed to create strategy: %v", err)
	}

	// Duplicate should fail
	err = db.CreateStrategy("rsi_strategy", "Duplicate", "trending")
	if err == nil {
		t.Error("Expected error for duplicate strategy_id")
	}
}

func TestAddAndGetHypothesis(t *testing.T) {
	db, err := NewResearchDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	db.CreateStrategy("test_strategy", "Test", "any")

	err = db.AddHypothesis("test_strategy", "hyp_001", "RSI oversold signals work", "strategy_v1.go")
	if err != nil {
		t.Errorf("Failed to add hypothesis: %v", err)
	}

	hypotheses, err := db.GetStrategyHypotheses("test_strategy")
	if err != nil {
		t.Errorf("Failed to get hypotheses: %v", err)
	}

	if len(hypotheses) != 1 {
		t.Fatalf("Expected 1 hypothesis, got %d", len(hypotheses))
	}

	if hypotheses[0].Description != "RSI oversold signals work" {
		t.Errorf("Expected description 'RSI oversold signals work', got '%s'", hypotheses[0].Description)
	}
}

func TestUpdateHypothesisEvaluation(t *testing.T) {
	db, err := NewResearchDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	db.CreateStrategy("test_strategy", "Test", "any")
	db.AddHypothesis("test_strategy", "hyp_001", "Test hypothesis", "code.go")

	evaluation := map[string]interface{}{
		"supported": true,
		"confidence": "high",
	}
	lessons := []string{"Lesson 1", "Lesson 2"}

	err = db.UpdateHypothesisEvaluation("hyp_001", evaluation, lessons, "confirmed")
	if err != nil {
		t.Errorf("Failed to update hypothesis: %v", err)
	}

	hypotheses, _ := db.GetStrategyHypotheses("test_strategy")
	if hypotheses[0].Status != "confirmed" {
		t.Errorf("Expected status 'confirmed', got '%s'", hypotheses[0].Status)
	}

	if len(hypotheses[0].LessonsLearned) != 2 {
		t.Errorf("Expected 2 lessons, got %d", len(hypotheses[0].LessonsLearned))
	}
}

func TestAddAndGetIterations(t *testing.T) {
	db, err := NewResearchDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	db.CreateStrategy("test_strategy", "Test", "any")

	metrics := map[string]interface{}{
		"sharpe_ratio": 1.5,
		"total_return": 25.5,
	}
	changes := []string{"Added ADX filter", "Adjusted position sizing"}

	err = db.AddIteration("test_strategy", "iter_001", "v1", metrics, changes, "Initial version", 0.0)
	if err != nil {
		t.Errorf("Failed to add iteration: %v", err)
	}

	err = db.AddIteration("test_strategy", "iter_002", "v2", metrics, changes, "Refined version", 10.5)
	if err != nil {
		t.Errorf("Failed to add second iteration: %v", err)
	}

	iterations, err := db.GetStrategyIterations("test_strategy", 10)
	if err != nil {
		t.Errorf("Failed to get iterations: %v", err)
	}

	if len(iterations) != 2 {
		t.Fatalf("Expected 2 iterations, got %d", len(iterations))
	}

	// Should be in DESC order (newest first)
	if iterations[0].IterationID != "iter_002" {
		t.Errorf("Expected first iteration to be 'iter_002', got '%s'", iterations[0].IterationID)
	}

	if iterations[1].Improvement != 0.0 {
		t.Errorf("Expected improvement 0.0 for first iteration, got %.2f", iterations[1].Improvement)
	}
}

func TestAddAndGetInsights(t *testing.T) {
	db, err := NewResearchDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	db.CreateStrategy("test_strategy", "Test", "any")

	err = db.AddInsight("test_strategy", "RSI works better in ranging markets", "performance", 0.8)
	if err != nil {
		t.Errorf("Failed to add insight: %v", err)
	}

	err = db.AddInsight("test_strategy", "ATR-based stops are effective", "risk", 0.9)
	if err != nil {
		t.Errorf("Failed to add second insight: %v", err)
	}

	insights, err := db.GetStrategyInsights("test_strategy", 10)
	if err != nil {
		t.Errorf("Failed to get insights: %v", err)
	}

	if len(insights) != 2 {
		t.Fatalf("Expected 2 insights, got %d", len(insights))
	}

	if insights[0].Confidence != 0.9 {
		t.Errorf("Expected confidence 0.9, got %.2f", insights[0].Confidence)
	}
}

func TestObservePattern(t *testing.T) {
	db, err := NewResearchDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	db.CreateStrategy("test_strategy", "Test", "any")

	// First observation
	err = db.ObservePattern("test_strategy", "Consecutive losses during trends", "Strong uptrend", 0.7, true)
	if err != nil {
		t.Errorf("Failed to observe pattern: %v", err)
	}

	// Second observation of same pattern
	err = db.ObservePattern("test_strategy", "Consecutive losses during trends", "Strong downtrend", 0.8, true)
	if err != nil {
		t.Errorf("Failed to observe pattern again: %v", err)
	}

	patterns, err := db.GetStrategyPatterns("test_strategy")
	if err != nil {
		t.Errorf("Failed to get patterns: %v", err)
	}

	if len(patterns) != 1 {
		t.Fatalf("Expected 1 pattern (merged), got %d", len(patterns))
	}

	if patterns[0].Frequency != 2 {
		t.Errorf("Expected frequency 2, got %d", patterns[0].Frequency)
	}

	// Confidence should be averaged
	expectedConfidence := (0.7 + 0.8) / 2.0
	if patterns[0].Confidence != expectedConfidence {
		t.Errorf("Expected confidence %.2f, got %.2f", expectedConfidence, patterns[0].Confidence)
	}
}

func TestSaveFeedback(t *testing.T) {
	db, err := NewResearchDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	db.CreateStrategy("test_strategy", "Test", "any")

	feedback := map[string]interface{}{
		"status": "success",
		"metrics": map[string]float64{
			"sharpe_ratio": 1.8,
		},
	}

	err = db.SaveFeedback("test_strategy", "iter_001", "BACKTEST", feedback)
	if err != nil {
		t.Errorf("Failed to save feedback: %v", err)
	}
}

func TestGetResearchSummary(t *testing.T) {
	db, err := NewResearchDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	db.CreateStrategy("test_strategy", "Test", "any")

	// Add some data
	db.AddHypothesis("test_strategy", "hyp_001", "Hypothesis 1", "code.go")
	db.UpdateHypothesisEvaluation("hyp_001", nil, nil, "confirmed")

	db.AddHypothesis("test_strategy", "hyp_002", "Hypothesis 2", "code.go")
	db.UpdateHypothesisEvaluation("hyp_002", nil, nil, "rejected")

	db.AddHypothesis("test_strategy", "hyp_003", "Hypothesis 3", "code.go")

	db.AddIteration("test_strategy", "iter_001", "v1", map[string]interface{}{}, []string{}, "v1", 0.0)
	db.AddIteration("test_strategy", "iter_002", "v2", map[string]interface{}{}, []string{}, "v2", 10.5)
	db.AddIteration("test_strategy", "iter_003", "v3", map[string]interface{}{}, []string{}, "v3", 5.2)

	db.AddInsight("test_strategy", "Insight 1", "performance", 0.8)
	db.AddInsight("test_strategy", "Insight 2", "risk", 0.9)

	db.ObservePattern("test_strategy", "Pattern 1", "Context 1", 0.7, true)

	summary, err := db.GetResearchSummary("test_strategy")
	if err != nil {
		t.Errorf("Failed to get summary: %v", err)
	}

	if summary.TotalHypotheses != 3 {
		t.Errorf("Expected 3 total hypotheses, got %d", summary.TotalHypotheses)
	}

	if summary.Confirmed != 1 {
		t.Errorf("Expected 1 confirmed hypothesis, got %d", summary.Confirmed)
	}

	if summary.Rejected != 1 {
		t.Errorf("Expected 1 rejected hypothesis, got %d", summary.Rejected)
	}

	if summary.Active != 1 {
		t.Errorf("Expected 1 active hypothesis, got %d", summary.Active)
	}

	if summary.TotalIterations != 3 {
		t.Errorf("Expected 3 iterations, got %d", summary.TotalIterations)
	}

	expectedAvgImprovement := (10.5 + 5.2) / 2.0
	if summary.AverageImprovement != expectedAvgImprovement {
		t.Errorf("Expected average improvement %.2f, got %.2f", expectedAvgImprovement, summary.AverageImprovement)
	}

	if summary.TotalInsights != 2 {
		t.Errorf("Expected 2 insights, got %d", summary.TotalInsights)
	}

	if summary.PatternsFound != 1 {
		t.Errorf("Expected 1 pattern, got %d", summary.PatternsFound)
	}
}

func TestUpdateStrategyStatus(t *testing.T) {
	db, err := NewResearchDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	db.CreateStrategy("test_strategy", "Test", "any")

	err = db.UpdateStrategyStatus("test_strategy", "archived")
	if err != nil {
		t.Errorf("Failed to update status: %v", err)
	}
}

func TestDatabasePersistence(t *testing.T) {
	tmpFile := "test_research.db"
	defer os.Remove(tmpFile)

	// Create and populate database
	db1, err := NewResearchDB(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	db1.CreateStrategy("persist_test", "Persistence Test", "any")
	db1.AddHypothesis("persist_test", "hyp_persist", "Test persistence", "code.go")
	db1.Close()

	// Reopen database
	db2, err := NewResearchDB(tmpFile)
	if err != nil {
		t.Fatalf("Failed to reopen database: %v", err)
	}
	defer db2.Close()

	hypotheses, err := db2.GetStrategyHypotheses("persist_test")
	if err != nil {
		t.Errorf("Failed to retrieve data after reopen: %v", err)
	}

	if len(hypotheses) != 1 {
		t.Errorf("Expected 1 hypothesis after reopen, got %d", len(hypotheses))
	}

	if hypotheses[0].Description != "Test persistence" {
		t.Errorf("Data not persisted correctly")
	}
}
