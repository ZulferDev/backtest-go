# Technical Indicators Guide

## Overview

Package `internal/indicators` menyediakan math primitives untuk strategi trading. Semua fungsi dirancang untuk:
- **Zero allocation** (minimize GC pressure)
- **Type-safe** (compile-time error detection)
- **Reproducible** (deterministic output)

## Available Indicators

### 1. SMA (Simple Moving Average)

```go
import "github.com/ZulferDev/backtest-go/internal/indicators"

// Calculate SMA for entire series
prices := []float64{10, 11, 12, 13, 14, 15}
sma, err := indicators.SMA(prices, 3)
// Result: [0, 0, 11, 12, 13, 14]

// Calculate only latest SMA (more efficient)
latest, err := indicators.SMALast(prices, 3)
// Result: 14.0 (average of 13, 14, 15)
```

**Use cases:**
- Trend identification
- Support/resistance levels
- SMA crossover strategies

---

### 2. EMA (Exponential Moving Average)

```go
// Calculate EMA for entire series
prices := []float64{10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
ema, err := indicators.EMA(prices, 5)

// Incremental EMA update (real-time)
newPrice := 21.0
prevEMA := ema[len(ema)-1]
newEMA := indicators.EMALast(newPrice, prevEMA, 5)
```

**Formula:**
```
k = 2 / (period + 1)
EMA = Price * k + EMA(prev) * (1 - k)
```

**Use cases:**
- More responsive than SMA
- Short-term trend following
- MACD calculation

---

### 3. RSI (Relative Strength Index)

```go
prices := []float64{44, 44.34, 44.09, 43.61, 44.33, 44.83, 45.10, 45.42, 45.84, 46.08}
rsi, err := indicators.RSI(prices, 14)

// Interpretation:
// RSI > 70: Overbought (potential sell signal)
// RSI < 30: Oversold (potential buy signal)
```

**Formula:**
```
RS = Average Gain / Average Loss
RSI = 100 - (100 / (1 + RS))
```

**Use cases:**
- Overbought/oversold detection
- Divergence analysis
- Mean reversion strategies

---

### 4. MACD (Moving Average Convergence Divergence)

```go
prices := make([]float64, 100)
// ... populate prices ...

macd, err := indicators.MACD(prices, 12, 26, 9)
// macd.MACD: MACD line (fast EMA - slow EMA)
// macd.Signal: Signal line (EMA of MACD)
// macd.Histogram: MACD - Signal

// Trading signals:
// Bullish: MACD crosses above Signal (histogram > 0)
// Bearish: MACD crosses below Signal (histogram < 0)
```

**Standard periods:**
- Fast: 12
- Slow: 26
- Signal: 9

**Use cases:**
- Trend direction identification
- Momentum confirmation
- Crossover strategies

---

### 5. ATR (Average True Range)

```go
high := []float64{52, 53, 54, 55, 56}
low := []float64{50, 51, 52, 53, 54}
close := []float64{51, 52, 53, 54, 55}

atr, err := indicators.ATR(high, low, close, 14)

// ATR measures volatility (not direction)
// Higher ATR = Higher volatility
// Use for stop-loss placement: StopLoss = EntryPrice - (ATR * multiplier)
```

**Formula:**
```
TR = max(High - Low, |High - Close(prev)|, |Low - Close(prev)|)
ATR = EMA(TR, period)
```

**Use cases:**
- Volatility measurement
- Stop-loss calculation
- Position sizing (risk-adjusted)

---

### 6. Bollinger Bands

```go
prices := []float64{20, 21, 22, 23, 24, 25, 24, 23, 22, 21, 20}
bb, err := indicators.Bollinger(prices, 20, 2.0)

// bb.Upper: Middle + (2 * StdDev)
// bb.Middle: SMA(20)
// bb.Lower: Middle - (2 * StdDev)

// Trading signals:
// Price touching upper band: Overbought
// Price touching lower band: Oversold
// Band squeeze: Low volatility (breakout imminent)
```

**Standard settings:**
- Period: 20
- Multiplier: 2.0

**Use cases:**
- Volatility-based mean reversion
- Breakout detection
- Price range boundaries

---

## Example Strategy: SMA Crossover

```go
package strategies

import (
    "github.com/ZulferDev/backtest-go/internal/indicators"
    "github.com/ZulferDev/backtest-go/pkg/sdk"
)

type SMACrossover struct {
    fastPeriod int
    slowPeriod int
}

func (s *SMACrossover) Init(ctx sdk.InitContext) error {
    s.fastPeriod = 10
    s.slowPeriod = 30
    return nil
}

func (s *SMACrossover) OnBar(ctx sdk.BarContext, bar sdk.OHLCV) error {
    // Get historical close prices
    history := ctx.History(s.slowPeriod)
    if len(history) < s.slowPeriod {
        return nil // Not enough data
    }

    closes := make([]float64, len(history))
    for i, h := range history {
        closes[i] = h.Close
    }

    // Calculate SMAs
    fastSMA, _ := indicators.SMALast(closes, s.fastPeriod)
    slowSMA, _ := indicators.SMALast(closes, s.slowPeriod)

    // Trading logic
    if !ctx.HasOpenPosition() {
        if fastSMA > slowSMA {
            // Golden cross: Buy signal
            ctx.MarketBuy(1.0)
        }
    } else {
        if fastSMA < slowSMA {
            // Death cross: Sell signal
            ctx.CloseAll()
        }
    }

    return nil
}
```

---

## Performance Considerations

### 1. Use `Last()` variants when possible
```go
// ❌ Slow: Calculate entire series every bar
for i := range bars {
    sma, _ := indicators.SMA(prices[:i+1], 20)
}

// ✅ Fast: Calculate only latest value
sma, _ := indicators.SMALast(prices, 20)
```

### 2. Cache indicator state (Phase 2.1 TODO)
```go
// Future: Indicators with internal state for incremental updates
type CachedSMA struct {
    window []float64
    sum    float64
}
```

### 3. Avoid repeated allocations
```go
// ❌ Allocates new slice every call
closes := make([]float64, len(history))

// ✅ Reuse preallocated buffer (strategy struct field)
type MyStrategy struct {
    closeBuffer []float64
}
```

---

## Validation Rules

1. **Period > 0**: Semua period parameter harus positif
2. **Sufficient data**: Minimal `period` data points untuk kalkulasi valid
3. **Array length matching**: ATR memerlukan `high`, `low`, `close` dengan panjang sama
4. **No NaN/Inf**: Input data tidak boleh NaN atau Infinity

---

## Testing

Run indicator tests:
```bash
go test ./internal/indicators -v
```

Benchmark performance:
```bash
go test ./internal/indicators -bench=. -benchmem
```

---

## Next Steps (Phase 2.2)

- [ ] Position sizing helpers (Kelly, Fixed Fractional)
- [ ] Stop-loss primitives (Fixed, ATR-based, Trailing)
- [ ] Multi-timeframe data access
- [ ] Custom window functions (Rolling min/max, Percentile)
