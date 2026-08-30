package optimizer

import (
	"testing"

	"github.com/ZulferDev/backtest-go/pkg/data"
	"github.com/ZulferDev/backtest-go/pkg/sdk"
)

func TestGridSearchGeneration(t *testing.T) {
	ranges := []ParameterRange{
		{Name: "period", Type: "int", Min: 10, Max: 20, Step: 5},
		{Name: "threshold", Type: "float", Min: 0.5, Max: 1.0, Step: 0.25},
		{Name: "enabled", Type: "bool", Values: []interface{}{true, false}},
	}

	grid := NewGridSearch(ranges)
	combinations, err := grid.Generate()

	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	expectedSize := 3 * 3 * 2 // 18 combinations
	if len(combinations) != expectedSize {
		t.Errorf("Expected %d combinations, got %d", expectedSize, len(combinations))
	}

	// Verify first combination has all parameters
	if len(combinations[0]) != 3 {
		t.Errorf("Expected 3 parameters in combination, got %d", len(combinations[0]))
	}
}

func TestGridSearchEstimateSize(t *testing.T) {
	ranges := []ParameterRange{
		{Name: "a", Type: "int", Min: 1, Max: 5, Step: 1},
		{Name: "b", Type: "int", Min: 10, Max: 30, Step: 10},
	}

	grid := NewGridSearch(ranges)
	size := grid.EstimateSize()

	expected := 5 * 3 // 15
	if size != expected {
		t.Errorf("Expected size %d, got %d", expected, size)
	}
}

func TestGridSearchInvalidRange(t *testing.T) {
	ranges := []ParameterRange{
		{Name: "invalid", Type: "int", Min: 10, Max: 5, Step: 1}, // Min > Max
	}

	grid := NewGridSearch(ranges)
	_, err := grid.Generate()

	if err == nil {
		t.Error("Expected error for invalid range, got nil")
	}
}

func TestResultAggregatorAdd(t *testing.T) {
	criteria := []RankingCriteria{
		{Metric: "return", Weight: 1.0},
	}

	agg := NewResultAggregator(criteria)

	result := BacktestResult{
		TaskID:      "test-1",
		TotalReturn: 15.5,
		SharpeRatio: 1.2,
	}

	agg.Add(result)

	all := agg.GetAll()
	if len(all) != 1 {
		t.Errorf("Expected 1 result, got %d", len(all))
	}
}

func TestResultAggregatorGetTopN(t *testing.T) {
	criteria := []RankingCriteria{
		{Metric: "return", Weight: 1.0},
	}

	agg := NewResultAggregator(criteria)

	// Add results with different returns
	agg.Add(BacktestResult{TaskID: "test-1", TotalReturn: 10.0})
	agg.Add(BacktestResult{TaskID: "test-2", TotalReturn: 20.0})
	agg.Add(BacktestResult{TaskID: "test-3", TotalReturn: 15.0})

	top2 := agg.GetTopN(2)

	if len(top2) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(top2))
	}

	// Check that results are sorted by return descending
	if top2[0].TotalReturn != 20.0 {
		t.Errorf("Expected top result return 20.0, got %.2f", top2[0].TotalReturn)
	}
	if top2[1].TotalReturn != 15.0 {
		t.Errorf("Expected second result return 15.0, got %.2f", top2[1].TotalReturn)
	}
}

func TestResultAggregatorStatistics(t *testing.T) {
	criteria := []RankingCriteria{
		{Metric: "return", Weight: 1.0},
	}

	agg := NewResultAggregator(criteria)

	agg.Add(BacktestResult{TotalReturn: 10.0, SharpeRatio: 1.0, WinRate: 60.0})
	agg.Add(BacktestResult{TotalReturn: 20.0, SharpeRatio: 1.5, WinRate: 70.0})
	agg.Add(BacktestResult{TotalReturn: -5.0, SharpeRatio: 0.5, WinRate: 40.0})

	stats := agg.GetStatistics()

	if stats.Total != 3 {
		t.Errorf("Expected total 3, got %d", stats.Total)
	}

	if stats.Successful != 3 {
		t.Errorf("Expected successful 3, got %d", stats.Successful)
	}

	if stats.Profitable != 2 {
		t.Errorf("Expected profitable 2, got %d", stats.Profitable)
	}

	expectedAvgReturn := (10.0 + 20.0 - 5.0) / 3.0
	if stats.AvgReturn < expectedAvgReturn-0.1 || stats.AvgReturn > expectedAvgReturn+0.1 {
		t.Errorf("Expected avg return ~%.2f, got %.2f", expectedAvgReturn, stats.AvgReturn)
	}
}

func TestResultAggregatorFilterResults(t *testing.T) {
	criteria := []RankingCriteria{
		{Metric: "return", Weight: 1.0},
	}

	agg := NewResultAggregator(criteria)

	agg.Add(BacktestResult{TaskID: "test-1", TotalReturn: 10.0})
	agg.Add(BacktestResult{TaskID: "test-2", TotalReturn: 20.0})
	agg.Add(BacktestResult{TaskID: "test-3", TotalReturn: -5.0})

	// Filter only profitable results
	profitable := agg.FilterResults(func(r BacktestResult) bool {
		return r.TotalReturn > 0
	})

	if len(profitable) != 2 {
		t.Errorf("Expected 2 profitable results, got %d", len(profitable))
	}
}

func TestParallelExecutorCreation(t *testing.T) {
	executor := NewParallelExecutor(4)

	if executor.workers != 4 {
		t.Errorf("Expected 4 workers, got %d", executor.workers)
	}

	if executor.taskQueue == nil {
		t.Error("Task queue should not be nil")
	}

	if executor.resultQueue == nil {
		t.Error("Result queue should not be nil")
	}
}

func TestParallelExecutorStatus(t *testing.T) {
	executor := NewParallelExecutor(2)

	total, running := executor.GetStatus()

	if total != 0 {
		t.Errorf("Expected total 0, got %d", total)
	}

	if running != 0 {
		t.Errorf("Expected running 0, got %d", running)
	}
}

// Mock strategy for testing
type mockStrategy struct{}

func (m *mockStrategy) Init(ctx sdk.InitContext) error {
	return nil
}

func (m *mockStrategy) OnBar(ctx sdk.BarContext, bar sdk.OHLCV) error {
	return nil
}

func TestBacktestTaskCreation(t *testing.T) {
	task := BacktestTask{
		ID: "task-1",
		Config: StrategyConfig{
			Name:     "TestStrategy",
			Strategy: &mockStrategy{},
		},
		Data:       []data.OHLCV{},
		InitialCap: 10000.0,
	}

	if task.ID != "task-1" {
		t.Errorf("Expected task ID 'task-1', got '%s'", task.ID)
	}

	if task.InitialCap != 10000.0 {
		t.Errorf("Expected initial capital 10000.0, got %.2f", task.InitialCap)
	}
}
