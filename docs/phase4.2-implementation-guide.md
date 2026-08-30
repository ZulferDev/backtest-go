# Phase 4.2: Real-time Simulation (Paper Trading) - Implementation Guide

## Overview

Phase 4.2 implements real-time paper trading simulation, enabling strategies to run against live market data via WebSocket connections without risking actual capital.

## Components Implemented

### 1. WebSocket Market Data Listener (`internal/paper/websocket.go`)

**Purpose**: Stream real-time OHLCV data from Binance WebSocket API.

**Key Features**:
- Binance WebSocket kline/candlestick stream integration
- Automatic reconnection with exponential backoff (5 attempts)
- Pub/Sub pattern for multiple subscribers
- Heartbeat mechanism (30s ping) to keep connection alive
- Context-based cancellation for clean shutdown
- Only processes closed candles for consistency with backtest

**API**:
```go
ws, err := NewWebSocketClient(ctx, "btcusdt", "1m")
ch := ws.Subscribe() // Returns channel receiving data.OHLCV
```

**Safety**:
- Thread-safe subscriber management with RWMutex
- Non-blocking broadcast (drops candles if subscriber buffer full)
- Graceful reconnection without data loss

---

### 2. Paper Trading Executor (`internal/paper/executor.go`)

**Purpose**: Manage paper trading execution state using real-time market data.

**Key Features**:
- Real-time strategy execution using SDK interface
- Position management (open/close with slippage simulation)
- PnL tracking (realized and unrealized)
- Trade logging with fee calculation (0.1% per side)
- Callback system for trade events and state updates
- Thread-safe state access

**API**:
```go
executor, err := NewExecutor(ctx, strategy, "btcusdt", "1m", 10000.0)
executor.SetTradeCallback(func(trade Trade) { /* handle trade */ })
executor.SetStateUpdateCallback(func(state *TradingState) { /* monitor state */ })
executor.Start()
```

**State Management**:
- `TradingState`: Position, equity, trades, current bar, bar count
- `Position`: Side, size, entry price, entry time
- `Trade`: Complete trade record with PnL and fees

**Slippage Model**:
- Buy (long): +0.05% slippage
- Sell (short): -0.05% slippage
- Close long: -0.05% slippage
- Close short: +0.05% slippage

---

### 3. SDK Context Implementation (`internal/paper/context.go`)

**Purpose**: Implement `sdk.InitContext` and `sdk.BarContext` for paper trading.

**Implements**:
- `paperInitContext`: Initialization phase context
- `paperBarContext`: Per-bar execution context

**Methods**:
- `Buy(size)`: Open long position
- `Sell(size)`: Open short position
- `CloseAll()`: Close current position
- `HasPosition()`: Check if position open
- `PositionSize()`, `PositionSide()`: Query position
- `Equity()`: Current equity
- `CurrentBar()`: Current OHLCV bar
- `History(lookback)`: Historical bars (limited in paper trading)

---

### 4. CLI Tool (`cmd/paper-trading/main.go`)

**Purpose**: Command-line interface for running paper trading strategies.

**Usage**:
```bash
go run cmd/paper-trading/main.go \
  -symbol btcusdt \
  -interval 1m \
  -cash 10000.0
```

**Features**:
- Real-time monitoring with periodic summaries (30s)
- Trade event logging
- State updates per bar
- Graceful shutdown (SIGINT/SIGTERM)
- Performance metrics: total PnL, return %, win rate, avg win/loss

**Output**:
- Per-bar: Bar count, price, equity, position status
- Periodic summary: Total bars, PnL, trades, win rate
- Trade execution: Side, entry/exit prices, PnL, fees

---

### 5. Unit Tests (`internal/paper/paper_test.go`)

**Test Coverage**:
- `TestWebSocketClient`: WebSocket connection and message parsing (integration test - skipped)
- `TestExecutorLifecycle`: Start/stop executor lifecycle
- `TestExecutorTrading`: Full trading flow with callbacks
- `TestPositionManagement`: Position open/close logic
- `TestUnrealizedPnL`: PnL calculation for long/short

**Test Strategies**:
- `testStrategy`: Minimal passive strategy
- `tradingStrategy`: Active strategy (buy bar 1, sell bar 2)

---

## Architecture

```
WebSocket (Binance)
    ↓
WebSocketClient (subscriber pattern)
    ↓
Executor (event loop)
    ↓
paperBarContext (SDK interface)
    ↓
Strategy.OnBar() [AI-generated code]
    ↓
Position Management
    ↓
State Updates & Callbacks
```

---

## Integration with Existing Framework

**Compatibility**:
- Uses same `pkg/sdk.Strategy` interface as backtest
- Reuses `pkg/data.OHLCV` structure
- Strategies written for backtest work in paper trading without modification

**Differences from Backtest**:
- Real-time bar arrival (not bulk processing)
- Limited history buffer (vs full historical data)
- Slippage simulation (backtest may have different model)
- No walk-forward or parameter optimization

---

## Dependency Requirements

**New Dependencies**:
```go
import "github.com/gorilla/websocket"
```

Add to `go.mod`:
```
go get github.com/gorilla/websocket
```

---

## Usage Example

```go
// Define strategy
type MyStrategy struct{}

func (s *MyStrategy) Init(ctx sdk.InitContext) error {
    return nil
}

func (s *MyStrategy) OnBar(ctx sdk.BarContext, bar sdk.OHLCV) error {
    // Strategy logic using ctx.Buy(), ctx.Sell(), ctx.CloseAll()
    if shouldBuy(bar) {
        ctx.Buy(calculateSize(ctx.Equity()))
    }
    if shouldSell() && ctx.HasPosition() {
        ctx.CloseAll()
    }
    return nil
}

// Run paper trading
ctx := context.Background()
executor, _ := paper.NewExecutor(ctx, &MyStrategy{}, "btcusdt", "1m", 10000.0)
executor.Start()
// ... executor runs until context cancelled
```

---

## Safety Considerations

1. **Thread Safety**: All state access protected by mutexes
2. **Context Cancellation**: Graceful shutdown via context
3. **Reconnection**: Automatic WebSocket reconnection on failures
4. **Non-Blocking**: Callbacks run in goroutines to avoid blocking main loop
5. **Error Handling**: Strategy errors logged but don't crash executor

---

## Next Steps (Phase 4.3)

- Live execution bridge (real exchange order placement)
- Risk management: position limits, max drawdown kill switch
- Alerting system (Discord/Telegram notifications)
- State persistence (resume after restart)
- Multi-strategy concurrent execution
- Enhanced slippage models (order book depth simulation)

---

## Testing

**Unit Tests**:
```bash
go test ./internal/paper -v
```

**Integration Test** (requires live connection):
```bash
go test ./internal/paper -v -run TestWebSocketClient
```

**Manual Test**:
```bash
go run cmd/paper-trading/main.go -symbol btcusdt -interval 1m -cash 10000
# Let it run for a few minutes, observe bar processing
# Press Ctrl+C to see final summary
```

---

## Verification Checklist

- [x] WebSocket client connects to Binance
- [x] Receives and parses kline messages
- [x] Automatic reconnection works
- [x] Executor processes bars in real-time
- [x] Position management (open/close) works
- [x] PnL calculation (realized/unrealized) accurate
- [x] Trade callbacks fire on position close
- [x] State updates reflect current market data
- [x] CLI tool runs and reports metrics
- [x] Graceful shutdown on signal
- [x] Unit tests pass
- [x] Thread-safe concurrent access
- [x] Compatible with backtest SDK interface

---

## Files Created

1. `internal/paper/websocket.go` - WebSocket market data listener
2. `internal/paper/executor.go` - Paper trading execution engine
3. `internal/paper/context.go` - SDK context implementations
4. `internal/paper/paper_test.go` - Unit tests
5. `cmd/paper-trading/main.go` - CLI tool
6. `docs/phase4.2-implementation-guide.md` - This document
