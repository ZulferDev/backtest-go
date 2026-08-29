# AGENT.md — Crypto Backtesting Framework

## Project Identity

**Name:** backtest-go  
**Language:** Golang  
**Purpose:** High-performance cryptocurrency trading strategy backtesting framework  
**Owner:** Fajar Hadi Tama  
**Created:** 2026-08-29  

---

## Vision

Build a zero-bloat, production-ready backtesting engine for crypto trading strategies with:
- **Speed**: Vectorized operations where possible, concurrent data processing
- **Accuracy**: Realistic order execution (slippage, fees, latency simulation)
- **Extensibility**: Plugin-based strategy system, custom indicator support
- **Observability**: Rich metrics, equity curves, trade-level analysis

---

## Architecture Principles

1. **Separation of Concerns**
   - Data layer: fetch, store, validate
   - Strategy layer: signal generation, indicator composition
   - Execution layer: order simulation, position management
   - Analytics layer: performance metrics, visualization

2. **Performance First**
   - Minimize allocations in hot paths
   - Use sync.Pool for frequently allocated objects
   - Benchmark-driven optimization

3. **Testability**
   - Every component has unit tests
   - Integration tests for end-to-end flows
   - Property-based testing for edge cases

4. **Production-Ready**
   - Configuration via YAML/env
   - Structured logging (zerolog)
   - Graceful error handling

---

## Tech Stack

| Component | Choice | Rationale |
|-----------|--------|----------|
| CLI | cobra | Standard, feature-rich |
| Config | viper | Flexible format support |
| HTTP | resty | Ergonomic, retry logic |
| Storage | parquet-go | Columnar, efficient |
| Math | gonum | Mature scientific computing |
| Charts | go-echarts | HTML output, portable |
| Logging | zerolog | Fast, structured |
| Testing | testify | Assertions, mocking |

---

## Project Structure

```
backtest-go/
├── cmd/
│   ├── backtest/         # Main CLI entrypoint
│   └── fetch/            # Data downloader utility
├── pkg/
│   ├── data/            # OHLCV fetchers, storage, validation
│   ├── strategy/        # Strategy interface & base implementations
│   ├── indicators/      # Technical indicators library
│   ├── backtest/        # Engine, portfolio, order execution
│   └── metrics/         # Performance analytics
├── internal/
│   └── util/            # Internal helpers
├── strategies/          # User-defined strategy plugins
├── testdata/            # Sample data for tests
├── results/             # Backtest outputs
├── .gitignore
├── go.mod
├── go.sum
├── Makefile
├── README.md
└── AGENT.md
```

---

## Development Phases

### Phase 1: Foundation (Week 1-2)
- [ ] Project scaffolding
- [ ] Data fetcher (Binance REST API)
- [ ] OHLCV storage (CSV, later Parquet)
- [ ] Basic indicators (SMA, EMA, RSI, MACD)
- [ ] Unit tests for data pipeline

### Phase 2: Core Engine (Week 3-4)
- [ ] Backtest engine (event-driven loop)
- [ ] Portfolio manager (position tracking, PnL)
- [ ] Order execution simulator (market orders)
- [ ] Commission & slippage modeling
- [ ] Integration tests

### Phase 3: Analytics (Week 5-6)
- [ ] Performance metrics (Sharpe, Sortino, drawdown)
- [ ] Equity curve generation
- [ ] Trade log export (JSON/CSV)
- [ ] HTML report with charts
- [ ] Walk-forward validation framework

### Phase 4: Advanced Features (Week 7+)
- [ ] Limit order simulation
- [ ] Multi-timeframe strategies
- [ ] Portfolio backtesting (multiple symbols)
- [ ] Parameter optimization (grid search, genetic algo)
- [ ] Live trading adapter (paper trading)

---

## Strategy Development Workflow

1. **Hypothesis**: Define edge/alpha source
2. **Indicator**: Implement required technical indicators
3. **Strategy**: Code entry/exit logic
4. **Backtest**: Run on historical data
5. **Analyze**: Review metrics, equity curve, trade distribution
6. **Iterate**: Refine parameters, risk management
7. **Validate**: Walk-forward test on unseen data
8. **Paper Trade**: Test in real-time (simulated)
9. **Deploy**: Live trading with capital allocation

---

## Example Strategy

```go
// strategies/sma_crossover.go
package strategies

import (
    "backtest-go/pkg/data"
    "backtest-go/pkg/indicators"
    "backtest-go/pkg/strategy"
)

type SMACrossover struct {
    fast *indicators.SMA
    slow *indicators.SMA
}

func NewSMACrossover(fastPeriod, slowPeriod int) *SMACrossover {
    return &SMACrossover{
        fast: indicators.NewSMA(fastPeriod),
        slow: indicators.NewSMA(slowPeriod),
    }
}

func (s *SMACrossover) OnBar(bar data.OHLCV) strategy.Signal {
    s.fast.Update(bar.Close)
    s.slow.Update(bar.Close)

    if !s.fast.Ready() || !s.slow.Ready() {
        return strategy.SignalNone
    }

    fastVal := s.fast.Value()
    slowVal := s.slow.Value()

    // Golden cross: fast MA crosses above slow MA
    if fastVal > slowVal && s.fast.Prev() <= s.slow.Prev() {
        return strategy.SignalBuy
    }

    // Death cross: fast MA crosses below slow MA
    if fastVal < slowVal && s.fast.Prev() >= s.slow.Prev() {
        return strategy.SignalSell
    }

    return strategy.SignalNone
}

func (s *SMACrossover) Name() string {
    return "SMA Crossover"
}
```

---

## Configuration Example

```yaml
# config.yaml
backtest:
  symbol: BTCUSDT
  interval: 1h
  start_date: 2023-01-01
  end_date: 2024-01-01
  initial_capital: 10000.0
  commission: 0.001  # 0.1%
  slippage: 0.0005   # 0.05%

strategy:
  name: sma_crossover
  params:
    fast_period: 10
    slow_period: 30

risk:
  position_size: 0.95  # % of capital per trade
  stop_loss: 0.02      # 2%
  take_profit: 0.05    # 5%

data:
  source: binance
  cache_dir: ./data

output:
  results_dir: ./results
  format: html  # html, json, csv
```

---

## Key Decisions Log

### 2026-08-29: Language Choice — Golang
**Rationale:**
- Native concurrency for real-time data streams
- Static typing reduces runtime errors
- Fast compile times for rapid iteration
- Single binary deployment (no runtime dependencies)
- Strong standard library

### 2026-08-29: Storage Format — Parquet
**Rationale:**
- Columnar format = efficient time-series queries
- Compression (5-10x smaller than CSV)
- Schema enforcement
- Wide ecosystem support (Python, R for analysis)

---

## Performance Targets

- **Backtest 1 year of 1h bars**: < 1 second
- **Backtest 1 year of 1m bars**: < 10 seconds
- **Parameter optimization (100 combinations)**: < 2 minutes
- **Memory footprint**: < 500 MB for typical backtest

---

## Testing Strategy

1. **Unit Tests**: Every indicator, strategy component
2. **Integration Tests**: Full backtest runs with known outcomes
3. **Property Tests**: Invariants (e.g., equity never negative without margin)
4. **Regression Tests**: Lock in metrics for canonical strategies
5. **Benchmarks**: Track performance regressions

```bash
make test          # Run all tests
make test-unit     # Unit tests only
make test-integration
make bench         # Benchmarks
make coverage      # Generate coverage report
```

---

## Contributing Guidelines

1. **Code Style**: Follow `gofmt`, use `golangci-lint`
2. **Commits**: Conventional commits (feat, fix, docs, test, refactor)
3. **PRs**: Include tests, update docs if API changes
4. **Benchmarks**: Required for performance-critical code

---

## Resources

### Books
- *Advances in Financial Machine Learning* — Marcos López de Prado
- *Algorithmic Trading* — Ernie Chan
- *Quantitative Trading* — Ernie Chan

### Papers
- "The Deflated Sharpe Ratio" (Bailey & López de Prado)
- "Backtesting" (Campbell Harvey)

### Communities
- QuantConnect Forum
- /r/algotrading
- Golang #trading Slack

---

## License

MIT (to be decided)

---

**Last Updated:** 2026-08-29  
**Maintainer:** zeroxx 🦀
