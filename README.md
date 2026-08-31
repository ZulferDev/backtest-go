# backtest-go

**AI-Driven Quantitative Trading Research Infrastructure**

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Status](https://img.shields.io/badge/Status-Production%20Ready-success)]()

---

## Overview

**backtest-go** is a production-ready backtesting framework where AI acts as an autonomous quantitative researcher and code creator. Unlike traditional backtesting tools, AI writes complete trading strategy code in Go, not just parameters.

### Key Features

🤖 **AI-First Design**
- AI writes complete strategy code in Go
- Sandboxed Strategy SDK prevents unsafe operations
- Automated code validation and testing
- Self-improving through research memory

⚡ **High Performance**
- Parallel backtest execution (8+ concurrent workers)
- Zero-allocation hot paths for indicators
- Grid search with exhaustive parameter combinations
- 6,600+ lines of optimized Go code

🔒 **Safety & Robustness**
- AST-based code validation (no unsafe imports/goroutines)
- Walk-forward testing for overfitting detection
- Comprehensive test coverage (45%+)
- CircleCI automated validation

📊 **Rich Analytics**
- 15+ performance metrics (Sharpe, Sortino, Drawdown, etc.)
- Multi-criteria result ranking
- Research memory for pattern tracking
- HTML reports with equity curves

---

## Quick Start

### Installation

```bash
# Clone repository
git clone https://github.com/ZulferDev/backtest-go.git
cd backtest-go

# Install dependencies
go mod download

# Verify installation
go test ./...
go build ./...
```

### Write Your First Strategy

Create `my_strategy.go`:

```go
package strategies

import (
    "github.com/ZulferDev/backtest-go/pkg/sdk"
    "github.com/ZulferDev/backtest-go/internal/indicators"
)

type SimpleMA struct {
    shortPeriod int
    longPeriod  int
}

func (s *SimpleMA) Init(ctx sdk.InitContext) error {
    s.shortPeriod = 20
    s.longPeriod = 50
    return nil
}

func (s *SimpleMA) OnBar(ctx sdk.BarContext, bar sdk.OHLCV) error {
    history := ctx.History(s.longPeriod + 1)
    if len(history) < s.longPeriod+1 {
        return nil
    }
    
    closes := extractCloses(history)
    shortMA, _ := indicators.SMALast(closes, s.shortPeriod)
    longMA, _ := indicators.SMALast(closes, s.longPeriod)
    
    if !ctx.HasOpenPosition() && shortMA > longMA {
        ctx.MarketBuy(1.0)
    } else if ctx.HasOpenPosition() && shortMA < longMA {
        ctx.CloseAll()
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

### Run Backtest

```bash
go run cmd/backtest/main.go \
  -strategy my_strategy.go \
  -data data/BTCUSDT_1h.json \
  -capital 10000
```

**Output:**
```
Backtest Results:
  Total Return:     23.45%
  Sharpe Ratio:     1.82
  Max Drawdown:     12.34%
  Win Rate:         55.67%
  
Results saved to: results.json
Report saved to: report.html
```

---

## Documentation

📚 **Complete documentation available in [`docs/`](docs/)**

### Essential Guides

- **[Documentation Index](docs/README.md)** - Start here for navigation
- **[User Guide](docs/USER_GUIDE.md)** - Complete guide for developers (801 lines)
- **[AI Agent Guide](docs/AI_AGENT_GUIDE.md)** - Guide for AI researchers (1,001 lines)
- **[Development Summary](docs/DEVELOPMENT_SUMMARY.md)** - Full project history (714 lines)
- **[Testing Guide](docs/TESTING_GUIDE.md)** - Testing procedures and best practices

### Core Concepts

- **[Methodology](docs/methodology.md)** - Research methodology and AI paradigm
- **[Architecture](docs/architecture.md)** - System architecture design
- **[Coding Standards](docs/coding-standards.md)** - Safety rules and constraints

### For Developers

```bash
# Read user guide for tutorials
cat docs/USER_GUIDE.md

# Run tests
go test ./... -v

# Run with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### For AI Agents

```bash
# Read AI agent guide
cat docs/AI_AGENT_GUIDE.md

# Study anti-hallucination rules
cat docs/context_window_management.md

# Initialize research
# Follow 7-phase lifecycle: CONCEIVE → WRITE → LINT → TEST → BACKTEST → ANALYZE → REFINE
```

---

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                      AI AGENT                            │
│  Hypothesis → Code → Analysis → Learning → Iteration    │
└─────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│                  Strategy SDK Context                    │
│     Safe API: Buy(), Sell(), CloseAll(), History()      │
└─────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│                  Backtest Engine (Go)                    │
│   Event Loop → Order Execution → PnL Calculation        │
└─────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────┐
│                    Data Pipeline                         │
│      Binance/Bybit → Validation → Normalization         │
└─────────────────────────────────────────────────────────┘
```

---

## Features

### Core Framework

- ✅ Event-driven backtest engine
- ✅ Strategy SDK with sandboxed execution
- ✅ 6 technical indicators (SMA, EMA, RSI, MACD, ATR, Bollinger)
- ✅ Position sizing and risk management
- ✅ AST-based code validation
- ✅ 15+ performance metrics

### AI Integration

- ✅ Code generation pipeline
- ✅ Analytical feedback loop
- ✅ Walk-forward overfitting detection
- ✅ Research memory database
- ✅ Context window management
- ✅ Anti-hallucination system

### Optimization & Trading

- ✅ Parallel execution (8+ workers)
- ✅ Grid search parameter exploration
- ✅ Multi-criteria result ranking
- 📋 Paper trading (documented, ready for implementation)
- 📋 Live deployment (documented, ready for implementation)

---

## Project Statistics

```
Production Code:     3,768 lines
Test Code:           1,655 lines
Documentation:       7,546 lines
Total:              12,969 lines

Test Coverage:      45%+ across packages
CircleCI Status:    ✅ All pipelines GREEN
```

### Development Status

| Phase | Status | Description |
|-------|--------|-------------|
| **Phase 0** | ✅ Complete | Foundation & Research |
| **Phase 1** | ✅ Complete | Core Backtest Engine |
| **Phase 2** | ✅ Complete | Rich Strategy Framework |
| **Phase 3** | ✅ Complete | AI Researcher Integration |
| **Phase 4.1** | ✅ Complete | Mass Optimization |
| **Phase 4.2** | 📋 Documented | Real-time Simulation |
| **Phase 4.3** | 📋 Documented | Deployment Automation |

---

## Available Indicators

```go
// Moving Averages
indicators.SMA(closes, period)      // Simple Moving Average
indicators.EMA(closes, period)      // Exponential Moving Average

// Momentum
indicators.RSI(closes, period)      // Relative Strength Index
indicators.MACD(closes, 12, 26, 9)  // MACD with signal line

// Volatility
indicators.ATR(highs, lows, closes, period)           // Average True Range
indicators.BollingerBands(closes, period, multiplier) // Bollinger Bands
```

## Risk Management

```go
// Position Sizing
risk.NewFixedFractional(0.02)           // Risk 2% per trade
risk.NewKellyCriterion(...)             // Kelly Criterion

// Stop Loss
risk.PercentStopLoss(entry, 5.0, "long")     // 5% stop
risk.ATRStopLoss(entry, atr, 2.0, "long")    // ATR-based
risk.NewTrailingStop(entry, 5.0, "long")     // Trailing stop
```

---

## Example: Mass Optimization

```go
// Define parameter ranges
ranges := []optimizer.ParameterRange{
    {Name: "short_period", Type: "int", Min: 10, Max: 30, Step: 5},
    {Name: "long_period", Type: "int", Min: 40, Max: 100, Step: 10},
}

// Generate combinations
grid := optimizer.NewGridSearch(ranges)
combinations, _ := grid.Generate()

// Execute in parallel
executor := optimizer.NewParallelExecutor(8)
executor.Start()

// Submit tasks
for _, params := range combinations {
    task := optimizer.BacktestTask{...}
    executor.Submit(task)
}

// Get top results
aggregator := optimizer.NewResultAggregator(criteria)
top10 := aggregator.GetTopN(10)
```

---

## Development Principles

1. **Accuracy First** - Backtest results are source of truth
2. **AI as Researcher** - Not just parameter optimizer, but code creator
3. **Safety Boundaries** - Strict code validation prevents unsafe operations
4. **Continuous Learning** - Research memory system tracks patterns
5. **CI/CD Strict** - All tests must pass in CircleCI before merge

---

## Testing

```bash
# Run all tests
go test ./...

# Run with race detector
go test -race ./...

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run benchmarks
go test -bench=. -benchmem ./...
```

---

## CI/CD

CircleCI automatically runs on every push:
- ✅ Linting (golangci-lint v1.60.3)
- ✅ Tests with race detector
- ✅ Benchmarks
- ✅ Coverage reporting

---

## Contributing

See [AGENTS.md](AGENTS.md) for development guidelines and AI agent integration protocols.

### Code Quality Standards

- Follow [coding standards](docs/coding-standards.md)
- Write tests for new features
- Maintain 40%+ test coverage
- Pass all CircleCI checks
- Document public APIs

---

## Documentation

**Total: 7,546 lines across 12 files**

Quick links:
- [Documentation Index](docs/README.md) - Complete navigation
- [User Guide](docs/USER_GUIDE.md) - For human developers
- [AI Agent Guide](docs/AI_AGENT_GUIDE.md) - For AI researchers
- [Development Summary](docs/DEVELOPMENT_SUMMARY.md) - Project history
- [Testing Guide](docs/TESTING_GUIDE.md) - Testing procedures

---

## License

MIT License - see [LICENSE](LICENSE) file for details

---

## Links

- **Repository:** https://github.com/ZulferDev/backtest-go
- **Documentation:** [docs/](docs/)
- **Issues:** https://github.com/ZulferDev/backtest-go/issues

---

## Status

**Current:** Phase 4.1 Complete  
**Next:** Phase 4.2 Implementation (Real-time Simulation)  
**Version:** 1.0  
**Last Updated:** 2026-08-31

---

**Ready to start?** Read the [User Guide](docs/USER_GUIDE.md) or [AI Agent Guide](docs/AI_AGENT_GUIDE.md) to begin! 🚀
