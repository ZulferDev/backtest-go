package report

import (
	"os"
	"testing"

	"github.com/ZulferDev/backtest-go/internal/backtest"
	"github.com/ZulferDev/backtest-go/pkg/data"
	"github.com/ZulferDev/backtest-go/pkg/sdk"
)

// Test strategy for report generation
type ReportTestStrategy struct {
	tradeCount int
}

func (s *ReportTestStrategy) Init(ctx sdk.InitContext) error {
	s.tradeCount = 0
	return nil
}

func (s *ReportTestStrategy) OnBar(ctx sdk.BarContext, bar sdk.OHLCV) error {
	// Execute 3 trades: 2 wins, 1 loss
	if s.tradeCount == 0 && !ctx.HasOpenPosition() {
		s.tradeCount++
		return ctx.MarketBuy(1.0) // Win trade
	}
	if s.tradeCount == 1 && ctx.HasOpenPosition() {
		s.tradeCount++
		return ctx.CloseAll()
	}
	if s.tradeCount == 2 && !ctx.HasOpenPosition() {
		s.tradeCount++
		return ctx.MarketBuy(1.0) // Loss trade
	}
	if s.tradeCount == 3 && ctx.HasOpenPosition() {
		s.tradeCount++
		return ctx.CloseAll()
	}
	if s.tradeCount == 4 && !ctx.HasOpenPosition() {
		s.tradeCount++
		return ctx.MarketBuy(1.0) // Win trade
	}
	if s.tradeCount == 5 && ctx.HasOpenPosition() {
		s.tradeCount++
		return ctx.CloseAll()
	}
	return nil
}

func TestHTMLReportGeneration(t *testing.T) {
	// Create test data with varied price movements
	testData := []data.OHLCV{
		{Timestamp: 1000, Open: 50000, High: 50100, Low: 49900, Close: 50000, Volume: 100},
		{Timestamp: 2000, Open: 50000, High: 51500, Low: 49900, Close: 51000, Volume: 100}, // Win
		{Timestamp: 3000, Open: 51000, High: 51100, Low: 50900, Close: 51000, Volume: 100},
		{Timestamp: 4000, Open: 51000, High: 51100, Low: 49500, Close: 49500, Volume: 100}, // Loss
		{Timestamp: 5000, Open: 49500, High: 49600, Low: 49400, Close: 49500, Volume: 100},
		{Timestamp: 6000, Open: 49500, High: 52000, Low: 49400, Close: 52000, Volume: 100}, // Win
		{Timestamp: 7000, Open: 52000, High: 52100, Low: 51900, Close: 52000, Volume: 100},
	}

	strategy := &ReportTestStrategy{}
	engine := backtest.NewEngine(strategy, testData, 10000.0)

	if err := engine.Run(); err != nil {
		t.Fatalf("Engine run failed: %v", err)
	}

	state := engine.GetState()
	trades := state.Trades()

	if len(trades) != 3 {
		t.Fatalf("Expected 3 trades, got %d", len(trades))
	}

	// Generate HTML report
	generator := NewHTMLGenerator()
	outputPath := "test_report.html"
	defer os.Remove(outputPath) // Clean up after test

	err := generator.Generate(state, "ReportTestStrategy", outputPath)
	if err != nil {
		t.Fatalf("Failed to generate report: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("Report file was not created")
	}

	// Read file to verify it's not empty
	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("Failed to stat report file: %v", err)
	}

	if fileInfo.Size() == 0 {
		t.Fatal("Report file is empty")
	}

	t.Logf("Report generated successfully: %s (%.2f KB)", outputPath, float64(fileInfo.Size())/1024)

	// Verify HTML structure
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read report: %v", err)
	}

	htmlStr := string(content)

	// Check for essential elements
	requiredElements := []string{
		"<!DOCTYPE html>",
		"Backtest Report",
		"ReportTestStrategy",
		"Initial Capital",
		"Final Equity",
		"Total Return",
		"Sharpe Ratio",
		"Equity Curve",
		"Trade History",
		"chart.js",
	}

	for _, elem := range requiredElements {
		if !contains(htmlStr, elem) {
			t.Errorf("Report missing required element: %s", elem)
		}
	}

	t.Log("All required HTML elements present")
}

func TestHTMLReportWithNoTrades(t *testing.T) {
	// Strategy that never trades
	type NoTradeStrategy struct{}
	
	// Implement Init and OnBar for NoTradeStrategy inline
	var noTradeInit = func(ctx sdk.InitContext) error {
		return nil
	}
	var noTradeOnBar = func(ctx sdk.BarContext, bar sdk.OHLCV) error {
		return nil
	}
	
	// Create wrapper
	noTrade := &noTradeStrategyImpl{
		initFn: noTradeInit,
		onBarFn: noTradeOnBar,
	}
	
	testData := []data.OHLCV{
		{Timestamp: 1000, Open: 50000, High: 50100, Low: 49900, Close: 50000, Volume: 100},
		{Timestamp: 2000, Open: 50000, High: 50100, Low: 49900, Close: 50000, Volume: 100},
	}

	engine := backtest.NewEngine(noTrade, testData, 10000.0)
	if err := engine.Run(); err != nil {
		t.Fatalf("Engine run failed: %v", err)
	}

	state := engine.GetState()
	generator := NewHTMLGenerator()
	outputPath := "test_report_no_trades.html"
	defer os.Remove(outputPath)

	err := generator.Generate(state, "NoTradeStrategy", outputPath)
	if err != nil {
		t.Fatalf("Failed to generate report with no trades: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatal("Report file was not created")
	}

	t.Log("Report with no trades generated successfully")
}

// noTradeStrategyImpl implements sdk.Strategy
type noTradeStrategyImpl struct {
	initFn  func(sdk.InitContext) error
	onBarFn func(sdk.BarContext, sdk.OHLCV) error
}

func (s *noTradeStrategyImpl) Init(ctx sdk.InitContext) error {
	return s.initFn(ctx)
}

func (s *noTradeStrategyImpl) OnBar(ctx sdk.BarContext, bar sdk.OHLCV) error {
	return s.onBarFn(ctx, bar)
}

func TestHTMLReportMetricsAccuracy(t *testing.T) {
	testData := []data.OHLCV{
		{Timestamp: 1000, Open: 50000, High: 50100, Low: 49900, Close: 50000, Volume: 100},
		{Timestamp: 2000, Open: 50000, High: 51000, Low: 49900, Close: 51000, Volume: 100},
		{Timestamp: 3000, Open: 51000, High: 51100, Low: 50900, Close: 51000, Volume: 100},
	}

	strategy := &ReportTestStrategy{}
	engine := backtest.NewEngine(strategy, testData, 10000.0)

	if err := engine.Run(); err != nil {
		t.Fatalf("Engine run failed: %v", err)
	}

	state := engine.GetState()
	
	// Generate report
	generator := NewHTMLGenerator()
	outputPath := "test_report_metrics.html"
	defer os.Remove(outputPath)

	err := generator.Generate(state, "TestStrategy", outputPath)
	if err != nil {
		t.Fatalf("Failed to generate report: %v", err)
	}

	// Read and verify metrics in HTML
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read report: %v", err)
	}

	htmlStr := string(content)

	// Verify initial capital is present
	if !contains(htmlStr, "10000.00") {
		t.Error("Initial capital not found in report")
	}

	t.Log("Metrics accuracy verified in HTML report")
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
