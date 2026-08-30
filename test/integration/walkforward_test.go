package integration

import (
	"testing"

	"github.com/ZulferDev/backtest-go/internal/analyzer"
	"github.com/ZulferDev/backtest-go/pkg/data"
	"github.com/ZulferDev/backtest-go/strategies"
)

// TestWalkForwardOrchestrator tests walk-forward analysis
func TestWalkForwardOrchestrator(t *testing.T) {
	// Generate test data (300 bars)
	testData := generateTestDataWalkForward(300)

	strategy := strategies.NewSMACrossover()

	// Walk-forward configuration:
	// - In-sample: 100 bars
	// - Out-of-sample: 50 bars
	// - Step: 50 bars (rolling window)
	wf := analyzer.NewWalkForward(100, 50, 50)

	results, err := wf.Execute(strategy, testData, 10000.0)
	if err != nil {
		t.Fatalf("Walk-forward execution failed: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("No walk-forward results generated")
	}

	t.Logf("Walk-forward windows executed: %d", len(results))

	// Verify each window has both in-sample and out-of-sample results
	for i, result := range results {
		if result.InSample == nil {
			t.Errorf("Window %d: in-sample result is nil", i)
		}
		if result.OutOfSample == nil {
			t.Errorf("Window %d: out-of-sample result is nil", i)
		}

		// Check for overfitting (large performance gap)
		if result.InSample != nil && result.OutOfSample != nil {
			inReturn := result.InSample.TotalReturn
			outReturn := result.OutOfSample.TotalReturn
			gap := inReturn - outReturn

			t.Logf("Window %d: In-sample=%.2f%%, Out-of-sample=%.2f%%, Gap=%.2f%%",
				i, inReturn*100, outReturn*100, gap*100)

			// Flag potential overfitting (gap > 10%)
			if gap > 0.10 {
				t.Logf("⚠️  Window %d shows potential overfitting (gap=%.2f%%)", i, gap*100)
			}
		}
	}

	// Aggregate analysis
	analyzer := analyzer.NewGapAnalyzer()
	overfitScore := analyzer.AnalyzeGap(results)
	
	t.Logf("Overfitting score: %.2f", overfitScore)

	if overfitScore > 0.5 {
		t.Logf("⚠️  High overfitting risk detected (score=%.2f)", overfitScore)
	}
}

// generateTestDataWalkForward creates synthetic OHLCV data
func generateTestDataWalkForward(bars int) []data.OHLCV {
	result := make([]data.OHLCV, bars)
	basePrice := 50000.0

	for i := 0; i < bars; i++ {
		// Trending data with cycles
		trend := float64(i) * 10.0
		cycle := 500.0 * (1.0 + 0.5*float64((i%50)-25)/25.0) // Cyclical pattern
		noise := float64((i*7)%20 - 10)

		close := basePrice + trend + cycle + noise
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
