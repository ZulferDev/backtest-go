# backtest-go

AI-Driven Quantitative Trading Research Infrastructure

## Overview

**backtest-go** is a production-ready backtesting framework where AI acts as an autonomous quantitative researcher and code creator. Unlike traditional backtesting tools, this framework enables AI to:

- Write complete trading strategy code in Go
- Validate code safety through AST analysis
- Execute parallel mass optimization across hundreds of parameter combinations
- Analyze results and iterate autonomously
- Detect overfitting through walk-forward analysis

## Key Features

### 🤖 AI-First Design
- AI writes strategy code, not just parameters
- Sandboxed Strategy SDK prevents unsafe operations
- Automated code validation and testing
- Self-improving through research memory

### ⚡ High Performance
- Parallel backtest execution (configurable workers)
- Efficient indicator calculations (zero-allocation hot paths)
- Grid search with exhaustive parameter combinations
- 5,000+ lines of optimized Go code

### 🔒 Safety & Robustness
- AST-based code validation (no unsafe imports/goroutines)
- Walk-forward testing for overfitting detection
- Comprehensive test coverage (45%+)
- CircleCI automated validation

### 📊 Rich Analytics
- 15+ performance metrics (Sharpe, Sortino, Drawdown, etc.)
- Multi-criteria result ranking
- Research memory for pattern tracking
- Detailed completion reports per phase

## Architecture

```
backtest-go/
├── internal/
│   ├── backtest/       # Core backtest engine
│   ├── optimizer/      # Parallel execution & grid search
│   ├── analyzer/       # Results analysis & walk-forward
│   ├── codegen/        # AI code generation pipeline
│   ├── indicators/     # Technical indicators library
│   ├── validator/      # AST validation & code safety
│   └── metrics/        # Performance metrics
├── pkg/
│   ├── sdk/           # Strategy SDK interface
│   └── data/          # OHLCV data structures
└── docs/              # Phase completion reports
```

## Current Status

### ✅ Completed Phases

**Phase 0: Foundation & Research**
- Documentation & methodology
- Exchange API research (Binance, Bybit)
- Data quality framework

**Phase 1: Core Backtest Engine**
- Data pipeline with validation
- Event-driven backtest engine
- Strategy SDK context
- Comprehensive metrics & reporting

**Phase 2: Rich Strategy Framework**
- Technical indicators (SMA, EMA, RSI, MACD, ATR, Bollinger)
- Risk management primitives (position sizing, stop-loss)
- Multi-timeframe support
- AST-based code validation

**Phase 3: AI Researcher Integration**
- Code generation pipeline
- Analytical feedback loop
- Walk-forward overfitting prevention
- Research memory system

**Phase 4.1: Mass Optimization**
- Parallel backtest executor (8+ workers)
- Grid search parameter exploration
- Multi-criteria result aggregation
- 45.7% test coverage

### 🚧 In Progress

**Phase 4.2: Real-time Simulation** (Next)
- WebSocket market data listener
- Paper trading execution state

**Phase 4.3: Deployment Automation** (Future)
- Live execution bridge
- Alerting & kill switches

## Quick Start

### Prerequisites
- Go 1.21+
- Git

### Installation

```bash
git clone https://github.com/ZulferDev/backtest-go.git
cd backtest-go
go mod download
```

### Run Tests

```bash
go test ./...
```

### Build

```bash
go build ./...
```

## Usage Example

### 1. Define Strategy

```go
package strategies

import (
    "github.com/ZulferDev/backtest-go/pkg/sdk"
    "github.com/ZulferDev/backtest-go/internal/indicators"
)

type SMACrossover struct {
    shortPeriod int
    longPeriod  int
}

func (s *SMACrossover) Init(ctx sdk.InitContext) error {
    s.shortPeriod = 20
    s.longPeriod = 50
    return nil
}

func (s *SMACrossover) OnBar(ctx sdk.BarContext, bar sdk.OHLCV) error {
    history := ctx.History(s.longPeriod + 1)
    if len(history) < s.longPeriod+1 {
        return nil
    }
    
    closes := extractCloses(history)
    shortSMA, _ := indicators.SMALast(closes, s.shortPeriod)
    longSMA, _ := indicators.SMALast(closes, s.longPeriod)
    
    if !ctx.HasOpenPosition() && shortSMA > longSMA {
        ctx.MarketBuy(1.0)
    } else if ctx.HasOpenPosition() && shortSMA < longSMA {
        ctx.CloseAll()
    }
    
    return nil
}
```

### 2. Run Backtest

```go
import (
    "github.com/ZulferDev/backtest-go/internal/backtest"
    "github.com/ZulferDev/backtest-go/pkg/data"
)

// Load historical data
data := loadOHLCV("BTCUSDT", "1h")

// Create strategy
strategy := &SMACrossover{}

// Run backtest
engine := backtest.NewEngine(strategy, data, 10000.0)
engine.Run()

// Get results
state := engine.GetState()
fmt.Printf("Total Return: %.2f%%\n", 
    (state.Equity()-state.InitialCash())/state.InitialCash()*100)
```

### 3. Mass Optimization

```go
import "github.com/ZulferDev/backtest-go/internal/optimizer"

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
    task := optimizer.BacktestTask{
        ID: fmt.Sprintf("task-%d", i),
        Config: optimizer.StrategyConfig{
            Strategy: strategy,
            Parameters: params,
        },
        Data: data,
        InitialCap: 10000.0,
    }
    executor.Submit(task)
}

// Collect results
aggregator := optimizer.NewResultAggregator(criteria)
for result := range executor.GetResults() {
    aggregator.Add(result)
}

// Get top strategies
top10 := aggregator.GetTopN(10)
```

## Project Statistics

- **Total Code:** 5,053 lines (production)
- **Files:** 54 Go files
- **Test Coverage:** 45.7% (optimizer package)
- **Commits:** 30+ (tracked via git)
- **Documentation:** 12 completion reports

## Development Principles

1. **Accuracy First** - Backtest results are source of truth
2. **AI as Researcher** - Not just parameter optimizer
3. **Safety Boundaries** - Strict code validation
4. **Continuous Learning** - Research memory system
5. **CI/CD Strict** - All tests must pass in CircleCI

## Testing

```bash
# Run all tests
go test -v ./...

# With coverage
go test -v -coverprofile=coverage.out ./...

# View coverage
go tool cover -html=coverage.out

# Run benchmarks
go test -bench=. -benchmem ./...
```

## Contributing

This is a research project. See `AGENTS.md` for development guidelines and AI agent integration protocols.

## CI/CD

CircleCI automatically runs on every push:
- Linting (golangci-lint v1.60.3)
- Tests with race detector
- Benchmarks
- Coverage reporting

## Documentation

Comprehensive documentation available in `docs/`:
- Architecture & methodology
- Exchange API specifications
- Phase completion reports
- Coding standards

## License

MIT

## Contact

GitHub: [ZulferDev/backtest-go](https://github.com/ZulferDev/backtest-go)

---

**Status:** Active Development | **Phase:** 4.1 Complete | **Next:** 4.2 Real-time Simulation
