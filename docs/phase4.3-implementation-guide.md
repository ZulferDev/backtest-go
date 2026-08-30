# Sub-phase 4.3: Deployment Automation - Implementation Guide

## Overview
Sub-phase 4.3 implements the final components for live trading: Live Execution Bridge, Alerting System, and Kill Switch mechanisms.

## Components Implemented

### 1. Live Execution Bridge (`internal/execution/bridge.go`)

**Purpose**: Connect paper trading system to live exchange execution

**Key Features**:
- `ExchangeAdapter` interface for exchange-agnostic implementation
- Real-time order placement and tracking (Market & Limit orders)
- Position synchronization with exchange
- Health monitoring and automatic reconnection
- Integration with Kill Switch for safety

**Core Types**:
```go
type Bridge struct {
    adapter       ExchangeAdapter
    strategy      sdk.Strategy
    currentPosition *Position
    orders        map[string]*Order
    killSwitch    *KillSwitch
}
```

**Usage**:
```go
bridge, err := NewBridge(ctx, BridgeConfig{
    Adapter:     binanceAdapter,
    Strategy:    myStrategy,
    Symbol:      "BTCUSDT",
    InitialCash: 10000,
    MaxDrawdown: 0.10,
})

bridge.Start()
defer bridge.Stop()
```

### 2. Kill Switch (`internal/execution/killswitch.go`)

**Purpose**: Automatic emergency stop on dangerous conditions

**Safety Limits**:
- Maximum drawdown from peak equity
- Maximum daily loss (absolute value)
- Maximum position size
- Maximum daily trades
- Maximum consecutive losses

**Key Methods**:
```go
func (ks *KillSwitch) Evaluate(currentEquity, initialEquity float64, position *Position) (bool, string)
func (ks *KillSwitch) RecordTrade(pnl float64)
func (ks *KillSwitch) Reset(currentEquity float64)
```

**Configuration Example**:
```go
config := KillSwitchConfig{
    MaxDrawdown:          0.10,  // 10% max drawdown
    MaxDailyLoss:         1000,  // $1000 max daily loss
    MaxPositionSize:      1.0,   // 1 BTC max position
    MaxDailyTrades:       20,    // 20 trades per day max
    MaxConsecutiveLosses: 5,     // 5 consecutive losses max
}
```

### 3. Alerting System (`internal/execution/alert.go`)

**Purpose**: Real-time monitoring and notification system

**Alert Levels**:
- `INFO`: Informational events
- `WARNING`: Potential issues
- `CRITICAL`: Immediate attention required

**Alert Types**:
- Order events (filled, failed)
- Position events (opened, closed)
- Risk events (drawdown, kill switch)
- Connection issues
- Errors

**Handlers**:
- Console output (built-in)
- File logging (JSON format)
- Webhook integration (extensible)

**Usage**:
```go
alertMgr, _ := NewAlertManager("alerts.log", AlertLevelInfo)
alertMgr.AddHandler(NewFileHandler("critical.log"))
alertMgr.AddHandler(NewWebhookHandler("https://hooks.slack.com/..."))

// Send alerts
alertMgr.Send(NewOrderAlert(AlertLevelInfo, AlertTypeOrderFilled, "Order filled", order))
alertMgr.Send(NewKillSwitchAlert("Max drawdown exceeded", equity))
```

### 4. Context Implementation (`internal/execution/context.go`)

**Purpose**: SDK context for live trading (implements `sdk.BarContext`)

**Features**:
- Forwards strategy calls to live exchange
- Position and equity queries from real exchange state
- Synchronous execution model (waits for order fills)

### 5. Tests (`internal/execution/execution_test.go`)

**Coverage**:
- Bridge creation and configuration
- Kill switch trigger conditions
- Alert manager filtering and storage
- Kill switch state management
- Trade recording and consecutive loss tracking

**Mock Exchange Adapter** provided for testing without live connections.

---

## Integration Architecture

```
Strategy (sdk.Strategy)
    ↓
liveBarContext (sdk.BarContext)
    ↓
Bridge
    ↓
ExchangeAdapter (Binance/Bybit)
    ↓
Live Exchange API
```

**Safety Layer**:
```
Bridge → KillSwitch → Evaluate → Emergency Stop
       ↓
   AlertManager → Handlers (Console/File/Webhook)
```

---

## Next Steps for Live Deployment

### 1. Implement Exchange Adapters
Create concrete implementations of `ExchangeAdapter` for:
- Binance Futures
- Bybit Perpetuals

**Required Methods**:
```go
func (a *BinanceAdapter) PlaceOrder(ctx, symbol, side, type, qty, price) (*Order, error)
func (a *BinanceAdapter) GetBalance(ctx) (float64, error)
func (a *BinanceAdapter) GetPosition(ctx, symbol) (*Position, error)
```

### 2. Create Live Trading CLI
Add command to `cmd/`:
```bash
go run cmd/live-trading/main.go \
  --strategy strategies/my_strategy.go \
  --symbol BTCUSDT \
  --exchange binance \
  --initial-capital 10000 \
  --max-drawdown 0.10
```

### 3. Add Strategy Hot-Reload
Allow strategy updates without stopping execution:
```go
bridge.UpdateStrategy(newStrategy)
```

### 4. Implement Trade Journal
Persistent storage for all trades:
```go
type TradeJournal struct {
    trades []Trade
    db     *sql.DB
}
```

### 5. Add Performance Dashboard
Real-time monitoring UI:
- Current position & PnL
- Equity curve
- Recent alerts
- Kill switch status

---

## Risk Management Checklist

Before going live:

- [ ] Test with paper trading minimum 1 week
- [ ] Verify kill switch triggers correctly
- [ ] Confirm alert delivery (webhook/email)
- [ ] Set conservative position sizing
- [ ] Enable API IP whitelist on exchange
- [ ] Use API keys with withdrawal disabled
- [ ] Test emergency stop manually
- [ ] Document rollback procedure
- [ ] Set up monitoring dashboard
- [ ] Prepare incident response plan

---

## Configuration Example

```go
// Production-safe configuration
config := BridgeConfig{
    Adapter:     binanceAdapter,
    Strategy:    validatedStrategy,
    Symbol:      "BTCUSDT",
    InitialCash: 10000,
    
    // Conservative risk limits
    MaxPositionSize: 0.1,    // 0.1 BTC max
    MaxDrawdown:     0.05,   // 5% max drawdown
    MaxDailyLoss:    200,    // $200 daily loss limit
}

killConfig := KillSwitchConfig{
    MaxDrawdown:          0.05,
    MaxDailyLoss:         200,
    MaxPositionSize:      0.1,
    MaxDailyTrades:       10,
    MaxConsecutiveLosses: 3,
}
```

---

## API Safety Guidelines

1. **Never expose private keys** in code or logs
2. **Use environment variables** for API credentials
3. **Enable IP whitelisting** on exchange
4. **Disable withdrawal** permissions on API keys
5. **Test on testnet first** before mainnet
6. **Monitor rate limits** to avoid bans
7. **Implement exponential backoff** on errors
8. **Log all orders** for audit trail

---

## Files Created

- `internal/execution/bridge.go` - Live execution bridge (482 lines)
- `internal/execution/context.go` - Live SDK context (141 lines)
- `internal/execution/killswitch.go` - Kill switch mechanism (201 lines)
- `internal/execution/alert.go` - Alerting system (307 lines)
- `internal/execution/execution_test.go` - Unit tests (301 lines)

**Total**: 1,432 lines of production code

---

## Status

✅ Live Execution Bridge implemented
✅ Kill Switch with multiple safety conditions
✅ Alerting system with multiple handlers
✅ SDK context for live trading
✅ Unit tests with mock exchange
✅ Documentation complete

**Ready for**:
- Exchange adapter implementation
- CLI integration
- Paper trading validation before live deployment
