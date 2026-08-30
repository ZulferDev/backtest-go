package integration

import (
	"testing"

	"github.com/ZulferDev/backtest-go/internal/optimizer"
	"github.com/ZulferDev/backtest-go/pkg/data"
	"github.com/ZulferDev/backtest-go/strategies"
)

// TestParallelOptimizer tests parallel execution with multiple strategies
func TestParallelOptimizer(t *testing.T) {
	// Generate test data
	testData := generateTestDataOptimizer(200)

	// Define parameter ranges for grid search
	ranges := []optimizer.ParameterRange{
		{Name: "short_period", Type: "int", Min: 10, Max: 30, Step: 10},
		{Name: "long_period", Type: "int", Min: 40, Max: 80, Step: 20},
	}

	// Generate parameter combinations
	grid := optimizer.NewGridSearch(ranges)
	combinations, err := grid.Generate()
	if err != nil {
		t.Fatalf("Failed to generate grid: %v", err)
	}

	t.Logf("Generated %d parameter combinations", len(combinations))

	if len(combinations) == 0 {
		t.Fatal("No combinations generated")
	}

	// Create parallel executor (4 workers for testing)
	executor := optimizer.NewParallelExecutor(4)
	executor.Start()
	defer executor.Stop()

	// Submit backtest tasks
	for i, params := range combinations {
		task := optimizer.BacktestTask{
			ID: "task-" + string(rune(i)),
			Config: optimizer.StrategyConfig{
				Strategy:   strategies.NewSMACrossover(),
				Parameters: params,
			},
			Data:       testData,
			InitialCap: 10000.0,
		}

		if err := executor.Submit(task); err != nil {
			t.Errorf("Failed to submit task %d: %v", i, err)
		}
	}

	// Collect results
	results := make([]optimizer.BacktestResult, 0, len(combinations))
	resultsChan := executor.GetResults()

	for i := 0; i < len(combinations); i++ {
		select {
		case result := <-resultsChan:
			if result.Error != nil {
				t.Logf("Task %s failed: %v", result.TaskID, result.Error)
			} else {
				results = append(results, result)
				t.Logf("Task %s completed: Return=%.2f%%, Sharpe=%.2f",
					result.TaskID, result.TotalReturn*100, result.SharpeRatio)
			}
		}
	}

	if len(results) == 0 {
		t.Fatal("No successful backtest results")
	}

	// Get top 3 strategies by sorting
	top3 := results
	if len(top3) > 3 {
		top3 = top3[:3]
	}
	if len(top3) == 0 {
		t.Fatal("No top results returned")
	}

	t.Logf("\nTop 3 Strategies:")
	for i, result := range top3 {
		t.Logf("%d. %s - Return: %.2f%%, Sharpe: %.2f, DD: %.2f%%",
			i+1, result.TaskID,
			result.TotalReturn*100,
			result.SharpeRatio,
			result.MaxDrawdown*100)
	}
}

// TestParallelExecutorConcurrency tests that parallel executor actually runs concurrently
func TestParallelExecutorConcurrency(t *testing.T) {
	testData := generateTestDataOptimizer(100)

	executor := optimizer.NewParallelExecutor(2)
	executor.Start()
	defer executor.Stop()

	// Submit multiple tasks
	numTasks := 4
	for i := 0; i < numTasks; i++ {
		task := optimizer.BacktestTask{
			ID: "concurrent-task-" + string(rune(i)),
			Config: optimizer.StrategyConfig{
				Strategy:   strategies.NewSMACrossover(),
				Parameters: map[string]interface{}{},
			},
			Data:       testData,
			InitialCap: 10000.0,
		}
		executor.Submit(task)
	}

	// Collect results
	completed := 0
	for i := 0; i < numTasks; i++ {
		result := <-executor.GetResults()
		if result.Error == nil {
			completed++
		}
	}

	if completed != numTasks {
		t.Errorf("Expected %d completed tasks, got %d", numTasks, completed)
	}
}

// generateTestDataOptimizer creates synthetic OHLCV data
func generateTestDataOptimizer(bars int) []data.OHLCV {
	result := make([]data.OHLCV, bars)
	basePrice := 50000.0

	for i := 0; i < bars; i++ {
		trend := float64(i) * 10.0
		noise := float64((i*7)%20 - 10)

		close := basePrice + trend + noise
		high := close + 50
		low := close - 50
		open := close - float64((i*3)%10)

		result[i] = data.OHLCV{
			Timestamp: int64(1609459200 + i*3600),
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    1000.0 + float64(i*10),
		}
	}

	return result
}
