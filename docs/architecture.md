# Architecture Design — backtest-go

## 1. High-Level Architecture

Sistem terdiri dari 4 layer modular:

```
+-------------------------------------------------------------+
|                        AI LAYER                             |
|          (Hypothesis, Optimization, Safety)                 |
+------------------------------+------------------------------+
                               | Schema I/O (JSON)
                               v
+-------------------------------------------------------------+
|                     STRATEGY FRAMEWORK                      |
|           (Strategy Interface, Parameter Schema)            |
+------------------------------+------------------------------+
                               | Signal / Event
                               v
+-------------------------------------------------------------+
|                      BACKTEST ENGINE                        |
|       (State Machine, Execution Sim, Position Tracker)      |
+------------------------------+------------------------------+
                               | Request / Data
                               v
+-------------------------------------------------------------+
|                        DATA PIPELINE                        |
|         (Exchange Client, Fetcher, Normalizer, Cache)       |
+-------------------------------------------------------------+
```

---

## 2. Component Interaction

### Backtest Loop (Event-Driven)

```
Data Pipeline          Backtest Engine         Strategy Framework        Execution Simulator
     |                        |                        |                          |
     |-- GetBar(time) ------->|                        |                          |
     |   (OHLCV)              |-- OnBar(OHLCV) ------->|                          |
     |                        |   (State, Pos)         |                          |
     |                        |<-- Return Signal ------|                          |
     |                        |    (BUY/SELL/HOLD)     |                          |
     |                        |-- ProcessSignal() ------------------------------->|
     |                        |                                                   |-- Execute() --+
     |                        |                                                   |   (Fee/Slip)  |
     |                        |<-- OrderFill Event -------------------------------+           |
     |                        |-- UpdatePosition() --->|                                      |
```

---

## 3. Data Flow

### Historical Data Ingestion

```
[Exchange REST API] ---> [Fetcher] ---> [Validator] ---> [Normalizer] ---> [Local Cache (Parquet)]
                                             |
                                             +---> Fail ---> [Discard/Log Error]
```

### Backtest Exec & Reporting

```
[Local Cache] ---> [Engine Event Loop] ---> [Metrics Calculator] ---> [HTML/JSON Report Generator]
```

---

## 4. Package Structure & Interface Design

### 4.1. Core Interfaces

#### Data Source
```go
package data

type OHLCV struct {
	Timestamp int64   `parquet:"name=timestamp, type=INT64"`
	Open      float64 `parquet:"name=open, type=DOUBLE"`
	High      float64 `parquet:"name=high, type=DOUBLE"`
	Low       float64 `parquet:"name=low, type=DOUBLE"`
	Close     float64 `parquet:"name=close, type=DOUBLE"`
	Volume    float64 `parquet:"name=volume, type=DOUBLE"`
}

type DataReader interface {
	Read(symbol string, timeframe string, start, end int64) ([]OHLCV, error)
}
```

#### Strategy
```go
package strategy

import "backtest-go/pkg/data"

type SignalType int
const (
	SignalHold SignalType = iota
	SignalBuy
	SignalSell
)

type Signal struct {
	Type           SignalType
	Price          float64
	Quantity       float64
	OrderType      string // "market", "limit", "stop"
}

type Strategy interface {
	Initialize(params map[string]interface{}) error
	OnBar(bar data.OHLCV, position Position) (Signal, error)
	Teardown() error
}

type Position interface {
	Size() float64
	EntryPrice() float64
	UnrealizedPnL(currentPrice float64) float64
}
```

#### Execution Simulator
```go
package execution

import (
	"backtest-go/pkg/data"
	"backtest-go/pkg/strategy"
)

type Order struct {
	ID        string
	Symbol    string
	Type      string // "market", "limit", "stop"
	Side      string // "buy", "sell"
	Price     float64
	Quantity  float64
	Timestamp int64
}

type FillResult struct {
	OrderID       string
	Price         float64
	Quantity      float64
	Commission    float64
	Slippage      float64
	ExecutionTime int64
}

type Simulator interface {
	SubmitOrder(order Order) error
	OnBar(bar data.OHLCV) ([]FillResult, error)
	GetPendingOrders() []Order
}
```

---

## 5. AI Integration Architecture

AI berkomunikasi secara asinkron dengan runner via JSON Schema.

### Workflow Optimisasi
1. **AI Generator** menulis Parameter Range (JSON).
2. **Go Engine** parsing, menjalankan parallel Grid/Genetic search.
3. **Go Engine** serialize metrics hasil eksekusi ke JSON.
4. **AI Analyzer** membaca JSON hasil untuk mendeteksi overfitting.

```
[AI Researcher Agent]
       |
       +---> Writes: config.json (ranges, constraints)
       |        |
       |        v
       |   [Go CLI Run (Grid/Walk-Forward)]
       |        |
       |        v
       |<-- Writes: results.json (metrics, out-of-sample)
       |
[Evaluation & Decision]
```

ponytail: target integrasi gRPC ditunda sampai CLI JSON robust.

---