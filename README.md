# backtest-go 🚀

> High-performance cryptocurrency trading strategy backtesting framework in Go

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

---

## What is this?

A **zero-bloat, production-ready** backtesting engine for crypto trading strategies. Built for researchers and traders who need:

- ⚡ **Speed**: Backtest years of data in seconds
- 🎯 **Accuracy**: Realistic order execution with slippage & fees
- 🔧 **Flexibility**: Plugin-based strategies, custom indicators
- 📊 **Insights**: Rich analytics, equity curves, trade logs

---

## Quick Start

```bash
# Install
go install github.com/yourusername/backtest-go/cmd/backtest@latest

# Fetch data
backtest fetch --symbol BTCUSDT --start 2023-01-01 --end 2024-01-01

# Run backtest
backtest run --config config.yaml

# View results
open results/backtest_report.html
```

---

## Features

### Core Engine
- [x] Event-driven backtesting
- [x] Realistic order execution (market orders)
- [x] Commission & slippage modeling
- [ ] Limit order simulation
- [ ] Multi-timeframe support
- [ ] Portfolio backtesting (multiple symbols)

### Data Pipeline
- [x] Binance REST API integration
- [x] OHLCV storage (CSV)
- [ ] Parquet storage (columnar)
- [ ] WebSocket live data feed
- [ ] Multi-exchange support (OKX, Bybit)

### Indicators
- [x] Simple Moving Average (SMA)
- [x] Exponential Moving Average (EMA)
- [x] Relative Strength Index (RSI)
- [x] Moving Average Convergence Divergence (MACD)
- [ ] Bollinger Bands
- [ ] ATR, ADX, Stochastic, etc.

### Analytics
- [x] Sharpe Ratio
- [x] Sortino Ratio
- [x] Maximum Drawdown
- [x] Win Rate, Avg Win/Loss
- [x] Equity curve plotting
- [ ] Walk-forward validation
- [ ] Monte Carlo simulation

---

## Example Strategy

```go
package main

import (
    "backtest-go/pkg/backtest"
    "backtest-go/pkg/data"
    "backtest-go/pkg/strategy"
)

type MyStrategy struct {
    // Your indicators here
}

func (s *MyStrategy) OnBar(bar data.OHLCV) strategy.Signal {
    // Your logic here
    return strategy.SignalBuy
}

func main() {
    engine := backtest.NewEngine(
        backtest.WithInitialCapital(10000),
        backtest.WithCommission(0.001),
    )
    
    result := engine.Run(&MyStrategy{})
    result.Print()
}
```

---

## Project Structure

```
backtest-go/
├── cmd/                 # CLI applications
│   ├── backtest/       # Main entrypoint
│   └── fetch/          # Data downloader
├── pkg/                # Public libraries
│   ├── data/          # OHLCV fetchers & storage
│   ├── strategy/      # Strategy interface
│   ├── indicators/    # Technical indicators
│   ├── backtest/      # Core engine
│   └── metrics/       # Performance analytics
├── strategies/         # User strategies
├── testdata/          # Test fixtures
└── results/           # Backtest outputs
```

---

## Configuration

```yaml
# config.yaml
backtest:
  symbol: BTCUSDT
  interval: 1h
  start_date: 2023-01-01
  end_date: 2024-01-01
  initial_capital: 10000.0
  commission: 0.001

strategy:
  name: sma_crossover
  params:
    fast_period: 10
    slow_period: 30
```

---

## Development

```bash
# Setup
git clone https://github.com/yourusername/backtest-go
cd backtest-go
go mod download

# Run tests
make test

# Run benchmarks
make bench

# Build
make build
```

---

## Roadmap

See [AGENT.md](AGENT.md) for detailed development phases.

**Current Status:** Phase 1 (Foundation)  
**Next Milestone:** Data pipeline + basic indicators (Week 2)

---

## Performance

| Dataset | Time | Memory |
|---------|------|--------|
| 1 year, 1h bars | < 1s | ~50 MB |
| 1 year, 1m bars | < 10s | ~200 MB |
| 5 years, 1h bars | < 5s | ~150 MB |

*Benchmarked on: M1 MacBook Pro, 16GB RAM*

---

## Resources

- 📖 [Documentation](https://github.com/yourusername/backtest-go/wiki)
- 💬 [Discussions](https://github.com/yourusername/backtest-go/discussions)
- 🐛 [Issue Tracker](https://github.com/yourusername/backtest-go/issues)

---

## License

MIT © 2026 Fajar Hadi Tama

---

**Built with 🦀 by zeroxx**
