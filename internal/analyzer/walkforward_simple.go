package analyzer

import (
	"fmt"
	"github.com/ZulferDev/backtest-go/internal/backtest"
	"github.com/ZulferDev/backtest-go/internal/metrics"
	"github.com/ZulferDev/backtest-go/pkg/data"
	"github.com/ZulferDev/backtest-go/pkg/sdk"
)

// WalkForward provides a simpler interface for walk-forward testing
type WalkForward struct {
	inSampleBars  int
	outSampleBars int
	stepBars      int
}

// NewWalkForward creates a new walk-forward analyzer
func NewWalkForward(inSampleBars, outSampleBars, stepBars int) *WalkForward {
	return &WalkForward{
		inSampleBars:  inSampleBars,
		outSampleBars: outSampleBars,
		stepBars:      stepBars,
	}
}

// WalkForwardWindow represents one window of walk-forward analysis
type WalkForwardWindow struct {
	WindowID    int
	InSample    *WindowResult
	OutOfSample *WindowResult
}

// WindowResult contains metrics for one window segment
type WindowResult struct {
	TotalReturn  float64
	SharpeRatio  float64
	MaxDrawdown  float64
	NumTrades    int
	WinRate      float64
	ProfitFactor float64
}

// Execute runs walk-forward analysis on the given strategy and data
func (wf *WalkForward) Execute(strategy sdk.Strategy, data []data.OHLCV, initialCash float64) ([]WalkForwardWindow, error) {
	if len(data) < wf.inSampleBars+wf.outSampleBars {
		return nil, fmt.Errorf("insufficient data: need at least %d bars, got %d",
			wf.inSampleBars+wf.outSampleBars, len(data))
	}

	var windows []WalkForwardWindow
	windowID := 0
	pos := 0

	for pos+wf.inSampleBars+wf.outSampleBars <= len(data) {
		windowID++

		// In-sample data
		inSampleData := data[pos : pos+wf.inSampleBars]
		inSampleResult, err := wf.runBacktest(strategy, inSampleData, initialCash)
		if err != nil {
			return nil, fmt.Errorf("in-sample backtest failed for window %d: %w", windowID, err)
		}

		// Out-of-sample data
		outSampleData := data[pos+wf.inSampleBars : pos+wf.inSampleBars+wf.outSampleBars]
		outSampleResult, err := wf.runBacktest(strategy, outSampleData, initialCash)
		if err != nil {
			return nil, fmt.Errorf("out-of-sample backtest failed for window %d: %w", windowID, err)
		}

		windows = append(windows, WalkForwardWindow{
			WindowID:    windowID,
			InSample:    inSampleResult,
			OutOfSample: outSampleResult,
		})

		// Move window forward
		pos += wf.stepBars

		// Safety: limit to 20 windows
		if windowID >= 20 {
			break
		}
	}

	if len(windows) == 0 {
		return nil, fmt.Errorf("no windows could be generated")
	}

	return windows, nil
}

// runBacktest executes backtest on a data segment
func (wf *WalkForward) runBacktest(strategy sdk.Strategy, data []data.OHLCV, initialCash float64) (*WindowResult, error) {
	engine := backtest.NewEngine(strategy, data, initialCash)
	if err := engine.Run(); err != nil {
		return nil, err
	}

	state := engine.GetState()
	trades := state.Trades()

	calc := metrics.NewCalculator(trades, initialCash)
	
	totalReturn := calc.TotalReturn(state.Equity())
	sharpe := calc.SharpeRatio()
	maxDD := calc.MaxDrawdown()
	profitFactor := calc.ProfitFactor()

	winRate := 0.0
	if len(trades) > 0 {
		wins := 0
		for _, t := range trades {
			if t.PnL > 0 {
				wins++
			}
		}
		winRate = float64(wins) / float64(len(trades)) * 100.0
	}

	return &WindowResult{
		TotalReturn:  totalReturn,
		SharpeRatio:  sharpe,
		MaxDrawdown:  maxDD,
		NumTrades:    len(trades),
		WinRate:      winRate,
		ProfitFactor: profitFactor,
	}, nil
}

// GapAnalyzer analyzes performance gap between in-sample and out-of-sample
type GapAnalyzer struct{}

// NewGapAnalyzer creates a new gap analyzer
func NewGapAnalyzer() *GapAnalyzer {
	return &GapAnalyzer{}
}

// AnalyzeGap calculates overfitting score based on IS/OOS performance gap
func (ga *GapAnalyzer) AnalyzeGap(windows []WalkForwardWindow) float64 {
	if len(windows) == 0 {
		return 0.0
	}

	totalGap := 0.0
	validWindows := 0

	for _, w := range windows {
		if w.InSample != nil && w.OutOfSample != nil {
			// Calculate return degradation
			returnGap := w.InSample.TotalReturn - w.OutOfSample.TotalReturn
			
			// Calculate sharpe degradation
			sharpeGap := w.InSample.SharpeRatio - w.OutOfSample.SharpeRatio
			
			// Weighted score: 60% return gap, 40% sharpe gap
			windowGap := (returnGap * 0.6) + (sharpeGap * 0.04)
			totalGap += windowGap
			validWindows++
		}
	}

	if validWindows == 0 {
		return 0.0
	}

	avgGap := totalGap / float64(validWindows)
	
	// Normalize to 0-1 scale (gap of 0.5 = score of 0.5)
	score := avgGap
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}

	return score
}
