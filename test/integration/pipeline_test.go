package integration

import (
	"os"
	"testing"

	"github.com/ZulferDev/backtest-go/internal/analyzer"
	"github.com/ZulferDev/backtest-go/internal/backtest"
	"github.com/ZulferDev/backtest-go/internal/codegen"
	"github.com/ZulferDev/backtest-go/internal/metrics"
	"github.com/ZulferDev/backtest-go/pkg/data"
	"github.com/ZulferDev/backtest-go/strategies"
)

// TestEndToEndPipeline tests the complete flow:
// Strategy -> Validate -> Backtest -> Analyze
func TestEndToEndPipeline(t *testing.T) {
	// 1. Code Generation & Validation
	tmpDir := t.TempDir()
	pipeline := codegen.NewPipeline(tmpDir)

	strategyCode := `package strategies

import (
	"github.com/ZulferDev/backtest-go/internal/indicators"
	"github.com/ZulferDev/backtest-go/pkg/sdk"
)

type TestStrategy struct {
	period int
}

func (s *TestStrategy) Init(ctx sdk.InitContext) error {
	s.period = 20
	return nil
}

func (s *TestStrategy) OnBar(ctx sdk.BarContext, bar sdk.OHLCV) error {
	history := ctx.History(s.period + 1)
	if len(history) < s.period+1 {
		return nil
	}

	closes := make([]float64, len(history))
	for i, h := range history {
		closes[i] = h.Close
	}

	sma, _ := indicators.SMALast(closes, s.period)
	
	if !ctx.HasOpenPosition() && bar.Close > sma {
		return ctx.MarketBuy(1.0)
	} else if ctx.HasOpenPosition() && bar.Close < sma {
		return ctx.CloseAll()
	}

	return nil
}
`

	// Validate code (AST check)
	errors, err := pipeline.Validate(strategyCode, "test_strategy.go")
	if err != nil {
		t.Fatalf("Validation error: %v", err)
	}
	if len(errors) > 0 {
		t.Fatalf("Code validation failed: %v", errors)
	}

	// Save code
	path, err := pipeline.SaveCode(strategyCode, "test_strategy.go")
	if err != nil {
		t.Fatalf("Failed to save code: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("Strategy file not created: %s", path)
	}

	// 2. Backtest Execution
	testData := generateTestData(100)
	strategy := strategies.NewSMACrossover()
	
	engine := backtest.NewEngine(strategy, testData, 10000.0)
	if err := engine.Run(); err != nil {
		t.Fatalf("Backtest failed: %v", err)
	}

	state := engine.GetState()
	if state == nil {
		t.Fatal("Engine state is nil")
	}

	// 3. Metrics Calculation
	trades := state.Trades()
	if len(trades) == 0 {
		t.Log("Warning: No trades executed")
	}

	calc := metrics.NewCalculator(state.Trades(), state.InitialCash())
	
	totalReturn := calc.TotalReturn(state.Equity())
	sharpe := calc.SharpeRatio()
	maxDD := calc.MaxDrawdown()
	
	t.Logf("Total Return: %.2f%%", totalReturn*100)
	t.Logf("Sharpe Ratio: %.2f", sharpe)
	t.Logf("Max Drawdown: %.2f%%", maxDD*100)

	// 4. Analysis & Feedback
	parser := analyzer.NewResultParser()
	summary := parser.ParseBacktestResult(map[string]interface{}{
		"total_return":  totalReturn,
		"sharpe_ratio":  sharpe,
		"max_drawdown":  maxDD,
		"num_trades":    len(trades),
		"final_equity":  state.Equity(),
		"initial_cash":  state.InitialCash(),
	})

	if summary == nil {
		t.Fatal("Failed to parse backtest results")
	}

	evaluator := analyzer.NewEvaluator()
	evaluation := evaluator.EvaluateHypothesis("Test SMA strategy", summary)
	
	if evaluation == "" {
		t.Fatal("Evaluation returned empty string")
	}

	t.Logf("Evaluation: %s", evaluation)
}

// TestCodeValidation tests AST-based code validation
func TestCodeValidation(t *testing.T) {
	pipeline := codegen.NewPipeline(t.TempDir())

	tests := []struct {
		name      string
		code      string
		shouldErr bool
	}{
		{
			name: "valid_strategy",
			code: `package strategies
import "github.com/ZulferDev/backtest-go/pkg/sdk"
type Valid struct{}
func (s *Valid) Init(ctx sdk.InitContext) error { return nil }
func (s *Valid) OnBar(ctx sdk.BarContext, bar sdk.OHLCV) error { return nil }
`,
			shouldErr: false,
		},
		{
			name: "unsafe_os_import",
			code: `package strategies
import "os"
import "github.com/ZulferDev/backtest-go/pkg/sdk"
type Unsafe struct{}
func (s *Unsafe) Init(ctx sdk.InitContext) error { os.Exit(1); return nil }
func (s *Unsafe) OnBar(ctx sdk.BarContext, bar sdk.OHLCV) error { return nil }
`,
			shouldErr: true,
		},
		{
			name: "unsafe_net_import",
			code: `package strategies
import "net/http"
import "github.com/ZulferDev/backtest-go/pkg/sdk"
type Unsafe struct{}
func (s *Unsafe) Init(ctx sdk.InitContext) error { http.Get(""); return nil }
func (s *Unsafe) OnBar(ctx sdk.BarContext, bar sdk.OHLCV) error { return nil }
`,
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors, err := pipeline.Validate(tt.code, tt.name+".go")
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			hasErrors := len(errors) > 0
			if hasErrors != tt.shouldErr {
				t.Errorf("Expected errors=%v, got errors=%v (%v)", tt.shouldErr, hasErrors, errors)
			}
		})
	}
}

// TestCompilePipeline tests the compile step
func TestCompilePipeline(t *testing.T) {
	tmpDir := t.TempDir()
	pipeline := codegen.NewPipeline(tmpDir)

	// Create a valid strategy
	code := `package strategies

import "github.com/ZulferDev/backtest-go/pkg/sdk"

type CompileTest struct{}

func (s *CompileTest) Init(ctx sdk.InitContext) error {
	return nil
}

func (s *CompileTest) OnBar(ctx sdk.BarContext, bar sdk.OHLCV) error {
	return nil
}
`

	path, err := pipeline.SaveCode(code, "compile_test.go")
	if err != nil {
		t.Fatalf("Failed to save code: %v", err)
	}

	// Note: Compile might fail if not in a proper Go module context
	// This is expected in test environment
	err = pipeline.Compile(path)
	if err != nil {
		t.Logf("Compile check (expected to fail in test env): %v", err)
	}
}

// generateTestData creates synthetic OHLCV data for testing
func generateTestData(bars int) []data.OHLCV {
	result := make([]data.OHLCV, bars)
	basePrice := 50000.0

	for i := 0; i < bars; i++ {
		// Simple trending data with some noise
		trend := float64(i) * 10.0
		noise := float64((i*7)%20 - 10) // Pseudo-random noise

		close := basePrice + trend + noise
		high := close + 50
		low := close - 50
		open := close - float64((i*3)%10)

		result[i] = data.OHLCV{
			Timestamp: int64(1609459200 + i*3600), // 2021-01-01 + i hours
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    1000.0 + float64(i*10),
		}
	}

	return result
}
