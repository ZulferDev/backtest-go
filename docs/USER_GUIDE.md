# User Guide - backtest-go

**Version:** 1.0  
**Last Updated:** 2026-08-31  
**Audience:** Human Developers & Quantitative Researchers

---

## Table of Contents

1. [Introduction](#introduction)
2. [Installation](#installation)
3. [Quick Start](#quick-start)
4. [Core Concepts](#core-concepts)
5. [Writing Strategies](#writing-strategies)
6. [Running Backtests](#running-backtests)
7. [Mass Optimization](#mass-optimization)
8. [Paper Trading](#paper-trading)
9. [Advanced Features](#advanced-features)
10. [Troubleshooting](#troubleshooting)

---

## Introduction

backtest-go adalah framework backtesting production-ready yang memungkinkan Anda untuk:

- Menulis strategi trading dalam Go dengan safety guarantees
- Menjalankan backtest historis dengan metrik lengkap
- Melakukan mass optimization dengan parallel execution
- Mendeteksi overfitting dengan walk-forward analysis
- Menjalankan paper trading real-time
- Integrasikan AI sebagai autonomous researcher

### Key Features

✅ **Event-Driven Engine** - Simulasi realistic order execution  
✅ **Rich Indicators** - SMA, EMA, RSI, MACD, ATR, Bollinger Bands  
✅ **Risk Management** - Position sizing, stop-loss, trailing stops  
✅ **Safety First** - AST validation mencegah unsafe code  
✅ **Parallel Execution** - Mass optimization dengan 8+ workers  
✅ **Overfitting Detection** - Walk-forward analysis  

---

## Installation

### Prerequisites

- Go 1.21 atau lebih tinggi
- Git
- (Optional) CircleCI account untuk CI/CD

### Clone Repository

```bash
git clone https://github.com/ZulferDev/backtest-go.git
cd backtest-go
```

### Install Dependencies

```bash
go mod download
```

### Verify Installation

```bash
# Run tests
go test ./...

# Build all packages
go build ./...
```

Expected output: All tests pass, all packages compile successfully.

---

## Quick Start

### 1. Fetch Historical Data

```bash
go run cmd/fetch-data/main.go \
  -exchange binance \
  -symbol BTCUSDT \
  -interval 1h \
  -start 2023-01-01 \
  -end 2023-12-31
```

Data akan disimpan di `data/BTCUSDT_1h.json`.

### 2. Create Simple Strategy

Create file `strategies/my_first_strategy.go`:

```go
package strategies

import (
    "github.com/ZulferDev/backtest-go/pkg/sdk"
    "github.com/ZulferDev/backtest-go/internal/indicators"
)

type SimpleMAStrategy struct {
    shortPeriod int
    longPeriod  int
}

func (s *SimpleMAStrategy) Init(ctx sdk.InitContext) error {
    s.shortPeriod = 20
    s.longPeriod = 50
    return nil
}

func (s *SimpleMAStrategy) OnBar(ctx sdk.BarContext, bar sdk.OHLCV) error {
    // Get historical data
    history := ctx.History(s.longPeriod + 1)
    if len(history) < s.longPeriod+1 {
        return nil // Not enough data yet
    }
    
    // Extract close prices
    closes := make([]float64, len(history))
    for i, h := range history {
        closes[i] = h.Close
    }
    
    // Calculate moving averages
    shortMA, _ := indicators.SMALast(closes, s.shortPeriod)
    longMA, _ := indicators.SMALast(closes, s.longPeriod)
    
    // Trading logic
    if !ctx.HasOpenPosition() {
        // Buy when short MA crosses above long MA
        if shortMA > longMA {
            ctx.MarketBuy(1.0) // Buy 1 unit
        }
    } else {
        // Sell when short MA crosses below long MA
        if shortMA < longMA {
            ctx.CloseAll()
        }
    }
    
    return nil
}
```

### 3. Run Backtest

```bash
go run cmd/backtest/main.go \
  -strategy strategies/my_first_strategy.go \
  -data data/BTCUSDT_1h.json \
  -capital 10000
```

Output:
```
Backtest Results:
  Total Return:     23.45%
  Sharpe Ratio:     1.82
  Max Drawdown:     12.34%
  Win Rate:         55.67%
  Total Trades:     45
  
Results saved to: results.json
Report saved to: report.html
```

### 4. View HTML Report

```bash
open report.html
```

Report berisi:
- Summary metrics
- Equity curve
- Trade log
- Drawdown analysis

---

## Core Concepts

### Strategy Interface

Semua strategi harus implement interface ini:

```go
type Strategy interface {
    Init(ctx InitContext) error
    OnBar(ctx BarContext, bar OHLCV) error
}
```

- `Init()` - Dipanggil sekali saat initialization
- `OnBar()` - Dipanggil setiap bar baru

### BarContext Methods

Methods yang tersedia dalam `OnBar()`:

```go
// Market data
CurrentBar() OHLCV              // Bar saat ini
History(lookback int) []OHLCV   // Historical bars

// Position queries
HasOpenPosition() bool          // Apakah ada position terbuka
CurrentPosition() Position      // Detail position saat ini

// Trading actions
MarketBuy(quantity float64) error    // Open long position
MarketSell(quantity float64) error   // Open short position
CloseAll() error                     // Close position

// Metrics
LogCustomMetric(key string, value float64)  // Log custom metric
```

### OHLCV Data Structure

```go
type OHLCV struct {
    Timestamp int64     // Unix timestamp (milliseconds)
    Open      float64   // Opening price
    High      float64   // Highest price
    Low       float64   // Lowest price
    Close     float64   // Closing price
    Volume    float64   // Volume
}
```

---

## Writing Strategies

### Using Technical Indicators

```go
import "github.com/ZulferDev/backtest-go/internal/indicators"

// Simple Moving Average
sma, err := indicators.SMA(closes, 20)
smaLast, err := indicators.SMALast(closes, 20) // Only last value

// Exponential Moving Average
ema, err := indicators.EMA(closes, 20)

// Relative Strength Index
rsi, err := indicators.RSI(closes, 14)

// MACD
macd, err := indicators.MACD(closes, 12, 26, 9)
// Access: macd.MACD, macd.Signal, macd.Histogram

// Average True Range
atr, err := indicators.ATR(highs, lows, closes, 14)

// Bollinger Bands
bb, err := indicators.BollingerBands(closes, 20, 2.0)
// Access: bb.Upper, bb.Middle, bb.Lower
```

### Position Sizing

```go
import "github.com/ZulferDev/backtest-go/internal/risk"

// Fixed Fractional: risk 2% per trade
sizer, _ := risk.NewFixedFractional(0.02)
size, _ := sizer.CalculateSize(equity, entryPrice, stopPrice)

// Kelly Criterion
kelly, _ := risk.NewKellyCriterion(winRate, avgWin, avgLoss, 0.5, 0.1)
size, _ := kelly.CalculateSize(equity, price)

// Percent of equity
percent, _ := risk.NewPercentPositionSize(0.1) // 10% of equity
size, _ := percent.CalculateSize(equity, price)
```

### Stop-Loss Management

```go
import "github.com/ZulferDev/backtest-go/internal/risk"

// Fixed price stop
stopPrice, _ := risk.FixedStopLoss(entryPrice, 100.0, "long")

// Percent-based stop
stopPrice, _ := risk.PercentStopLoss(entryPrice, 5.0, "long") // 5% below entry

// ATR-based stop (volatility adaptive)
stopPrice, _ := risk.ATRStopLoss(entryPrice, atr, 2.0, "long") // 2x ATR

// Trailing stop
trail, _ := risk.NewTrailingStop(entryPrice, 5.0, "long")
trail.Update(currentPrice)
if trail.IsTriggered(currentPrice) {
    ctx.CloseAll()
}
```

### Multi-Timeframe Analysis

```go
import "github.com/ZulferDev/backtest-go/internal/signal"

// Aggregate 1h bars to 4h
hourlyBars := ctx.History(100)
fourHourBars, _ := signal.AggregateToHigherTimeframe(hourlyBars, signal.TF4h)

// Get last completed bar (avoid lookahead bias)
lastBar, _ := signal.GetLastCompletedBar(bars, currentTime, signal.TF4h)
```

### Example: RSI Mean Reversion Strategy

```go
type RSIMeanReversion struct {
    rsiPeriod   int
    oversold    float64
    overbought  float64
    sizer       *risk.FixedFractional
}

func (s *RSIMeanReversion) Init(ctx sdk.InitContext) error {
    s.rsiPeriod = 14
    s.oversold = 30.0
    s.overbought = 70.0
    s.sizer, _ = risk.NewFixedFractional(0.02)
    return nil
}

func (s *RSIMeanReversion) OnBar(ctx sdk.BarContext, bar sdk.OHLCV) error {
    history := ctx.History(s.rsiPeriod + 1)
    if len(history) < s.rsiPeriod+1 {
        return nil
    }
    
    closes := extractCloses(history)
    rsi, _ := indicators.RSI(closes, s.rsiPeriod)
    currentRSI := rsi[len(rsi)-1]
    
    if ctx.HasOpenPosition() {
        // Exit when RSI reaches overbought
        if currentRSI > s.overbought {
            ctx.CloseAll()
        }
    } else {
        // Enter when RSI reaches oversold
        if currentRSI < s.oversold {
            stopPrice := bar.Close * 0.95 // 5% stop
            size, _ := s.sizer.CalculateSize(ctx.Equity(), bar.Close, stopPrice)
            ctx.MarketBuy(size)
        }
    }
    
    return nil
}

func extractCloses(bars []sdk.OHLCV) []float64 {
    closes := make([]float64, len(bars))
    for i, b := range bars {
        closes[i] = b.Close
    }
    return closes
}
```

---

## Running Backtests

### Basic Backtest

```bash
go run cmd/backtest/main.go \
  -strategy strategies/my_strategy.go \
  -data data/BTCUSDT_1h.json \
  -capital 10000
```

### With Custom Fees & Slippage

```bash
go run cmd/backtest/main.go \
  -strategy strategies/my_strategy.go \
  -data data/BTCUSDT_1h.json \
  -capital 10000 \
  -fee 0.001 \      # 0.1% fee per trade
  -slippage 0.0005  # 0.05% slippage
```

### Output Files

1. **results.json** - Raw metrics data
```json
{
  "summary": {
    "total_return": 0.2345,
    "sharpe_ratio": 1.82,
    "max_drawdown": 0.1234,
    "win_rate": 0.5567,
    "total_trades": 45
  },
  "trades": [...],
  "equity_curve": [...]
}
```

2. **report.html** - Visual report dengan charts

### Interpreting Results

**Profitability Metrics:**
- `Total Return` - Overall profit/loss percentage
- `Net Profit` - Absolute profit in currency
- `CAGR` - Compound Annual Growth Rate

**Risk Metrics:**
- `Sharpe Ratio` - Risk-adjusted return (> 1.5 good)
- `Sortino Ratio` - Downside risk-adjusted return
- `Max Drawdown` - Largest peak-to-trough decline (< 20% good)
- `Calmar Ratio` - Return / Max Drawdown

**Consistency Metrics:**
- `Win Rate` - Percentage of winning trades (> 50% good)
- `Profit Factor` - Gross profit / Gross loss (> 1.5 good)
- `Average Win/Loss` - Average profit vs average loss

---

## Mass Optimization

### Grid Search

Explore parameter combinations untuk find optimal settings.

```go
package main

import (
    "github.com/ZulferDev/backtest-go/internal/optimizer"
    "github.com/ZulferDev/backtest-go/pkg/data"
)

func main() {
    // Define parameter ranges
    ranges := []optimizer.ParameterRange{
        {
            Name: "short_period",
            Type: "int",
            Min:  10,
            Max:  30,
            Step: 5,
        },
        {
            Name: "long_period",
            Type: "int",
            Min:  40,
            Max:  100,
            Step: 10,
        },
        {
            Name: "rsi_oversold",
            Type: "float",
            Min:  20.0,
            Max:  35.0,
            Step: 5.0,
        },
    }
    
    // Generate combinations
    grid := optimizer.NewGridSearch(ranges)
    combinations, _ := grid.Generate()
    fmt.Printf("Total combinations: %d\n", len(combinations))
    
    // Load data
    data := loadData("data/BTCUSDT_1h.json")
    
    // Create parallel executor
    executor := optimizer.NewParallelExecutor(8) // 8 workers
    executor.Start()
    defer executor.Stop()
    
    // Submit tasks
    for i, params := range combinations {
        task := optimizer.BacktestTask{
            ID: fmt.Sprintf("task-%d", i),
            Config: optimizer.StrategyConfig{
                Strategy:   createStrategy(params),
                Parameters: params,
            },
            Data:       data,
            InitialCap: 10000.0,
        }
        executor.Submit(task)
    }
    
    // Collect results
    aggregator := optimizer.NewResultAggregator(map[string]float64{
        "sharpe_ratio": 0.4,
        "total_return": 0.3,
        "profit_factor": 0.2,
        "win_rate": 0.1,
    })
    
    for result := range executor.GetResults() {
        aggregator.Add(result)
    }
    
    // Get top strategies
    top10 := aggregator.GetTopN(10)
    for i, result := range top10 {
        fmt.Printf("%d. Sharpe: %.2f, Return: %.2f%%, Params: %v\n",
            i+1, result.Metrics.SharpeRatio, result.Metrics.TotalReturn*100, result.Parameters)
    }
}
```

### Run Optimization

```bash
go run cmd/optimize/main.go \
  -strategy strategies/my_strategy.go \
  -data data/BTCUSDT_1h.json \
  -workers 8 \
  -output optimization_results.json
```

### Analyzing Results

Best practices:
1. **Don't overfit** - Too many parameters = curve fitting
2. **Check robustness** - Top strategies should have similar parameters
3. **Validate out-of-sample** - Test best params on new data
4. **Look for stability** - Small parameter changes shouldn't drastically change results

---

## Paper Trading

Run strategies against live market data without risking capital.

### Start Paper Trading

```bash
go run cmd/paper-trading/main.go \
  -symbol btcusdt \
  -interval 1m \
  -cash 10000.0 \
  -strategy strategies/my_strategy.go
```

### Output

```
[2026-08-31 00:15:00] Connected to Binance WebSocket
[2026-08-31 00:15:01] Bar #1: BTCUSDT @ 50000.00, Equity: 10000.00
[2026-08-31 00:16:01] Bar #2: BTCUSDT @ 50100.00, Equity: 10000.00
[2026-08-31 00:16:01] TRADE: BUY 0.2 BTC @ 50100.00
[2026-08-31 00:17:01] Bar #3: BTCUSDT @ 50200.00, Equity: 10020.00 [Position: LONG 0.2 BTC]
...

=== 30-Second Summary ===
Bars Processed: 30
Total PnL: +234.56 (2.35%)
Trades: 3 (2 wins, 1 loss)
Win Rate: 66.67%
```

### Features

- Real-time WebSocket connection to Binance
- Automatic reconnection on failures
- Position tracking dengan unrealized PnL
- Trade logging dengan fees
- Periodic performance summaries
- Graceful shutdown (Ctrl+C)

### Safety

Paper trading simulates:
- Order execution dengan slippage (0.05%)
- Trading fees (0.1% per side)
- Position management
- PnL calculation

No real money at risk!

---

## Advanced Features

### Walk-Forward Analysis

Detect overfitting dengan testing strategy across multiple time windows.

```go
import "github.com/ZulferDev/backtest-go/internal/analyzer"

config := analyzer.WalkForwardConfig{
    InSampleDays:  90,  // 90 days training
    OutSampleDays: 30,  // 30 days testing
    StepDays:      30,  // Roll forward 30 days
}

orchestrator := analyzer.NewWalkForwardOrchestrator(config)
result, _ := orchestrator.Run(strategy, data, 10000.0)

fmt.Printf("Overfitting Score: %.2f\n", result.OverfittingScore)
fmt.Printf("Risk Level: %s\n", result.RiskLevel)
fmt.Printf("IS Sharpe: %.2f, OOS Sharpe: %.2f\n", 
    result.InSampleAvg.Sharpe, result.OutSampleAvg.Sharpe)
```

**Risk Levels:**
- **Low** (< 0.3): Strategy is robust
- **Medium** (0.3-0.6): Needs more validation
- **High** (> 0.6): Likely curve-fitted, revise strategy

### Code Validation

Validate strategy code before running:

```bash
go run cmd/validate/main.go -strategy strategies/my_strategy.go
```

AST validator checks for:
- ❌ Unsafe imports (`os`, `net`, `syscall`)
- ❌ Goroutine usage
- ❌ Syscalls
- ✅ Safe SDK usage only

### Research Memory

Track strategy iterations and learnings:

```go
import "github.com/ZulferDev/backtest-go/internal/analyzer"

memory := analyzer.NewResearchMemory("my_strategy", "BTCUSDT", "ranging")

// Record hypothesis
memory.AddHypothesis(analyzer.HypothesisRecord{
    ID:          "hyp_001",
    Description: "RSI oversold signals work in ranging markets",
    Status:      "testing",
})

// Record iteration
memory.AddIteration(analyzer.IterationRecord{
    Version:   "v1",
    Timestamp: time.Now(),
    Metrics: map[string]float64{
        "sharpe_ratio": 1.8,
        "max_drawdown": 0.12,
    },
    Changes: []string{"Initial implementation"},
})

// Save to disk
memory.Save("research_logs/my_strategy/memory.json")
```

---

## Troubleshooting

### Common Issues

#### 1. "Not enough data for indicator calculation"

**Problem:** Strategy requests more historical bars than available.

**Solution:**
```go
history := ctx.History(100)
if len(history) < 100 {
    return nil // Wait for more data
}
```

#### 2. "Divide by zero in indicator calculation"

**Problem:** Data contains zero or invalid values.

**Solution:**
```go
if bar.Close == 0 || bar.Volume == 0 {
    return nil // Skip invalid bar
}
```

#### 3. "Strategy validation failed: unsafe import"

**Problem:** Strategy imports disallowed packages.

**Solution:** Remove imports like `os`, `net`, `syscall`. Use only:
- `github.com/ZulferDev/backtest-go/pkg/sdk`
- `github.com/ZulferDev/backtest-go/internal/indicators`
- `github.com/ZulferDev/backtest-go/internal/risk`

#### 4. "WebSocket connection failed"

**Problem:** Network issue or rate limit.

**Solution:**
- Check internet connection
- Verify exchange API is accessible
- Wait a few minutes if rate limited
- Framework akan auto-reconnect

#### 5. "Backtest runs but no trades executed"

**Problem:** Entry conditions never met.

**Solution:**
- Add logging: `fmt.Printf("RSI: %.2f\n", currentRSI)`
- Check if indicators calculated correctly
- Verify data timeframe matches strategy logic
- Lower entry thresholds for testing

### Debug Mode

Enable verbose logging:

```bash
go run cmd/backtest/main.go \
  -strategy strategies/my_strategy.go \
  -data data/BTCUSDT_1h.json \
  -capital 10000 \
  -verbose
```

### Getting Help

1. Check documentation: `docs/`
2. Review examples: `examples/`
3. Run tests: `go test ./... -v`
4. Check CircleCI: Verify code passes CI
5. GitHub Issues: Report bugs or ask questions

---

## Best Practices

### Strategy Development

1. **Start simple** - Implement basic logic first, then add complexity
2. **Test incrementally** - Verify each component works before combining
3. **Use proper risk management** - Always implement stop-loss
4. **Avoid lookahead bias** - Only use data available at decision time
5. **Log key decisions** - Use `LogCustomMetric()` for debugging

### Backtesting

1. **Use realistic fees** - Include trading costs in simulation
2. **Account for slippage** - Don't assume perfect fills
3. **Test multiple timeframes** - Verify strategy works across different periods
4. **Check drawdowns** - Ensure you can handle worst-case losses
5. **Validate out-of-sample** - Always test on unseen data

### Optimization

1. **Limit parameters** - More parameters = more overfitting risk
2. **Use walk-forward** - Always validate with walk-forward analysis
3. **Check robustness** - Best results should cluster around similar params
4. **Be skeptical** - If results look too good, they probably are
5. **Document everything** - Keep research logs for future reference

### Going Live

1. **Paper trade first** - Minimum 1 week real-time validation
2. **Start small** - Use small position sizes initially
3. **Monitor closely** - Watch for unexpected behavior
4. **Have kill switches** - Set max drawdown limits
5. **Keep learning** - Continuously improve based on results

---

## Next Steps

1. ✅ Follow [Quick Start](#quick-start) to run your first backtest
2. ✅ Study [example strategies](../examples/)
3. ✅ Read [API Reference](API_REFERENCE.md) for detailed documentation
4. ✅ Explore [AI Integration Guide](AI_AGENT_GUIDE.md) for autonomous research
5. ✅ Join community discussions on GitHub

---

**Happy Trading! 📈**
