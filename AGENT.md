# AGENT.md — Framework Backtesting Crypto

## Identitas Proyek

**Nama:** backtest-go  
**Bahasa:** Golang  
**Tujuan:** Framework backtesting strategi trading cryptocurrency dengan performa tinggi  
**Pemilik:** Fajar Hadi Tama  
**Dibuat:** 2026-08-29  

---

## Visi

Membangun mesin backtesting yang production-ready untuk strategi trading crypto dengan fokus:
- **Kecepatan**: Operasi tervektorisasi, pemrosesan data concurrent
- **Akurasi**: Eksekusi order realistis (slippage, fee, simulasi latensi)
- **Ekstensibilitas**: Sistem strategi berbasis plugin, dukungan indikator custom
- **Observabilitas**: Metrik lengkap, equity curve, analisis per-trade

---

## Prinsip Arsitektur

1. **Pemisahan Concern**
   - Layer data: fetch, simpan, validasi
   - Layer strategi: generasi sinyal, komposisi indikator
   - Layer eksekusi: simulasi order, manajemen posisi
   - Layer analitik: metrik performa, visualisasi

2. **Performance First**
   - Minimalisir alokasi di hot path
   - Gunakan sync.Pool untuk objek yang sering dialokasikan
   - Optimisasi berbasis benchmark

3. **Testability**
   - Setiap komponen punya unit test
   - Integration test untuk flow end-to-end
   - Property-based testing untuk edge case

4. **Production-Ready**
   - Konfigurasi via YAML/env
   - Structured logging (zerolog)
   - Error handling yang graceful

---

## Tech Stack

| Komponen | Pilihan | Alasan |
|----------|---------|--------|
| CLI | cobra | Standar, feature-rich |
| Config | viper | Dukungan format fleksibel |
| HTTP | resty | Ergonomis, retry logic |
| Storage | parquet-go | Columnar, efisien |
| Math | gonum | Mature, scientific computing |
| Charts | go-echarts | Output HTML, portable |
| Logging | zerolog | Cepat, terstruktur |
| Testing | testify | Assertions, mocking |

---

## Struktur Proyek

```
backtest-go/
├── cmd/
│   ├── backtest/         # CLI entrypoint utama
│   └── fetch/            # Utility downloader data
├── pkg/
│   ├── data/            # Fetcher OHLCV, storage, validasi
│   ├── strategy/        # Interface strategi & implementasi base
│   ├── indicators/      # Library indikator teknikal
│   ├── backtest/        # Engine, portfolio, eksekusi order
│   └── metrics/         # Analitik performa
├── internal/
│   └── util/            # Helper internal
├── strategies/          # Plugin strategi user-defined
├── testdata/            # Sample data untuk testing
├── results/             # Output backtest
├── .gitignore
├── go.mod
├── go.sum
├── Makefile
└── AGENT.md
```

---

## Fase Development

### Phase 1: Fondasi
**Deliverable:**
- [ ] Scaffolding proyek (go mod, struktur folder)
- [ ] Data fetcher (Binance REST API dengan rate limiting)
- [ ] Storage OHLCV (CSV dengan validasi schema)
- [ ] Indikator dasar (SMA, EMA, RSI, MACD)
- [ ] Unit test (coverage >80% untuk data pipeline)

**Kriteria Sukses:**
- Bisa fetch 1 tahun data BTCUSDT 1h dalam <30 detik
- Semua indikator menghasilkan nilai yang benar (tervalidasi vs TradingView)

**Penjelasan:**
Fase ini membangun fondasi infrastruktur data dan library indikator. Tanpa data bersih, hasil backtest tidak bisa dipercaya (garbage in, garbage out). Indikator adalah building block untuk semua strategi.

**Komponen Utama:**
1. **Data Fetcher**: Client HTTP untuk Binance public API endpoint `/api/v3/klines`
   - Rate limiting: 1200 req/min (sesuai limit Binance)
   - Retry dengan exponential backoff
   - Pagination untuk data >1000 bar

2. **Storage CSV**: Format awal sebelum migrate ke Parquet
   ```csv
   timestamp,open,high,low,close,volume
   1640995200000,46000.5,47500.0,45800.0,47200.0,1234.56
   ```
   - Validasi: timestamp monoton, harga tidak negatif
   - Buffer untuk performance

3. **Indikator Dasar**:
   - **SMA**: Simple Moving Average (rolling window)
   - **EMA**: Exponential MA (bobot lebih ke harga terbaru)
   - **RSI**: Relative Strength Index (momentum oscillator 0-100)
   - **MACD**: Moving Average Convergence Divergence (trend-following)

---

### Phase 2: Core Engine
**Deliverable:**
- [ ] Backtest engine (event-driven loop)
- [ ] Portfolio manager (tracking posisi, PnL)
- [ ] Order execution simulator (market order)
- [ ] Modeling commission & slippage
- [ ] Integration test

**Kriteria Sukses:**
- Backtest buy-and-hold match dengan calculated return (selisih <0.1%)
- Engine bisa proses 10,000 bar dalam <5 detik

**Penjelasan:**
Ini "jantung" framework. Engine harus event-driven (proses bar-by-bar) untuk hindari look-ahead bias. Realisme penting: eksekusi yang terlalu optimis = false confidence.

**Komponen Utama:**
1. **Event-Driven Loop**:
   ```go
   for _, bar := range historicalData {
       signal := strategy.OnBar(bar)  // Strategi generate sinyal
       portfolio.ProcessSignal(signal, bar)  // Eksekusi
       portfolio.UpdateEquity(bar.Close)  // Update equity
   }
   ```
   - Tidak boleh "peek" ke data masa depan
   - Semua indikator dan strategi operasi pada timestamp yang sama

2. **Portfolio Manager**:
   ```go
   type Position struct {
       Symbol     string
       Size       float64    // jumlah unit
       EntryPrice float64
       EntryTime  time.Time
       StopLoss   float64    // opsional
       TakeProfit float64    // opsional
   }
   ```
   - Track posisi open/closed
   - Hitung PnL: unrealized (posisi open) & realized (posisi closed)
   - Track equity: `cash + unrealizedPnL`

3. **Order Execution (Market Order)**:
   ```
   Fill Price = Current Close × (1 + slippage)
   Commission = Fill Price × Size × Commission Rate
   ```
   - Market order: eksekusi langsung di harga saat ini
   - Limit/stop order belum di-support (Phase 4)

4. **Commission & Slippage**:
   - Commission (default Binance Futures):
     - Maker: 0.02% (0.0002)
     - Taker: 0.04% (0.0004)
   - Slippage: simulasi market impact
     - Order kecil: ~0.01-0.05%
     - Order besar: bisa 0.1%+ tergantung likuiditas

---

### Phase 3: Analitik
**Deliverable:**
- [ ] Metrik performa (Sharpe, Sortino, drawdown)
- [ ] Generasi equity curve
- [ ] Export trade log (JSON/CSV)
- [ ] HTML report dengan chart
- [ ] Framework walk-forward validation

**Kriteria Sukses:**
- Bisa generate report lengkap untuk strategi apapun dalam <1 menit
- Semua metrik match dengan perhitungan manual (verified)

**Penjelasan:**
Metrik menentukan viabilitas strategi. Visualisasi reveal pattern yang tidak obvious dari raw number. Walk-forward validation paling mendekati performa real-world.

**Komponen Utama:**
1. **Metrik Performa**:

   **a. Sharpe Ratio**
   ```
   Sharpe = (Mean Return - Risk-Free Rate) / Std Dev Returns
   ```
   - Ukur risk-adjusted return
   - >1 = bagus, >2 = sangat bagus, >3 = excellent
   - Annualized: kalikan √252 (trading days)

   **b. Sortino Ratio**
   ```
   Sortino = (Mean Return - Risk-Free Rate) / Downside Deviation
   ```
   - Seperti Sharpe, tapi hanya penalize downside volatility
   - Lebih realistis (upside volatility itu bagus)

   **c. Maximum Drawdown (MDD)**
   ```
   MDD = (Trough Value - Peak Value) / Peak Value
   ```
   - Penurunan terburuk dari peak ke trough
   - Penting: "berapa banyak saya bisa rugi?"
   - Recovery time juga penting

   **d. Win Rate**
   ```
   Win Rate = Winning Trades / Total Trades
   ```
   - Bukan satu-satunya metric (bisa 40% win rate tapi profitable)

   **e. Profit Factor**
   ```
   Profit Factor = Gross Profit / Gross Loss
   ```
   - >1 = profitable, >2 = kuat

2. **Equity Curve**:
   - Plot equity over time
   - Identifikasi:
     - Smooth uptrend = strategi konsisten
     - Sharp spike = luck atau overfitting
     - Long flat period = strategi tidak work

3. **Trade Log**:
   ```json
   {
     "trades": [
       {
         "entry_time": "2023-01-15T10:00:00Z",
         "exit_time": "2023-01-16T14:30:00Z",
         "symbol": "BTCUSDT",
         "side": "long",
         "entry_price": 21500.0,
         "exit_price": 22100.0,
         "size": 0.5,
         "pnl": 300.0,
         "pnl_pct": 2.79,
         "commission": 10.5
       }
     ]
   }
   ```
   - Enable post-analysis di Python/R
   - Audit trail: reproduce hasil backtest

4. **HTML Report**:
   - Equity curve (line chart)
   - Drawdown chart (underwater plot)
   - Monthly returns heatmap
   - Trade distribution histogram
   - Tabel summary metrik
   - Gunakan `go-echarts` untuk interactive chart

5. **Walk-Forward Validation**:
   ```
   Train Period 1 → Test Period 1 →
   Train Period 2 → Test Period 2 →
   ...
   ```
   - Cegah overfitting
   - Lebih realistis dari single train/test split
   - Contoh: train 6 bulan, test 1 bulan, roll forward

---

### Phase 4: Fitur Advanced
**Deliverable:**
- [ ] Simulasi limit order
- [ ] Strategi multi-timeframe
- [ ] Portfolio backtesting (multiple symbol)
- [ ] Optimisasi parameter (grid search, genetic algo)
- [ ] Live trading adapter (paper trading)

**Kriteria Sukses:**
- Strategi yang dioptimasi beat baseline >20% di walk-forward test
- Paper trading berjalan 7 hari tanpa error

**Penjelasan:**
Fitur production-grade dan tools optimisasi. Limit order lebih realistis, multi-timeframe untuk strategi advanced, optimisasi untuk cari parameter terbaik (tapi hati-hati overfitting!), paper trading sebagai bridge antara backtest dan live trading.

**Komponen Utama:**
1. **Limit Order Simulation**:
   - **Limit buy**: eksekusi jika harga turun ke/di bawah limit
   - **Limit sell**: eksekusi jika harga naik ke/di atas limit
   - **Fill logic**:
     ```go
     if bar.Low <= limitBuyPrice && bar.High >= limitBuyPrice {
         // Assume fill di limit price (optimistic)
         // Atau worst case: gunakan bar.Open (pessimistic)
     }
     ```
   - **Partial fill**: split order jika volume tidak cukup

2. **Multi-Timeframe Strategy**:
   - Contoh: gunakan trend daily, eksekusi di 1h bar
   ```go
   type MultiTFStrategy struct {
       dailyTrend  *indicators.SMA  // 200-day SMA di daily
       hourlyEntry *indicators.RSI  // RSI di 1h
   }
   ```
   - Challenge: sinkronisasi timestamp antar timeframe

3. **Portfolio Backtesting**:
   - Alokasi modal di BTC, ETH, SOL, dll
   - Analisis korelasi: benefit diversifikasi
   - Strategi rebalancing: equal-weight, risk-parity, dll

4. **Optimisasi Parameter**:
   
   **a. Grid Search**
   ```go
   for fastPeriod := 5; fastPeriod <= 20; fastPeriod += 5 {
       for slowPeriod := 30; slowPeriod <= 100; slowPeriod += 10 {
           result := runBacktest(fastPeriod, slowPeriod)
           // Track best Sharpe Ratio
       }
   }
   ```
   - Exhaustive, tapi lambat untuk parameter space besar
   
   **b. Genetic Algorithm**
   - Evolve parameter set over generation
   - Fitness function: Sharpe Ratio, profit factor, dll
   - Lebih cepat dari grid search untuk dimensi tinggi

5. **Live Trading Adapter (Paper Trading)**:
   - Connect ke exchange WebSocket untuk data real-time
   - Eksekusi sinyal strategi secara real-time (no money, mode simulasi)
   - Log paper trade vs backtest actual
   - Validasi final sebelum deploy real money

---

## Workflow Development Strategi

1. **Hipotesis**: Define edge/alpha source
2. **Indikator**: Implement indikator teknikal yang diperlukan
3. **Strategi**: Code logika entry/exit
4. **Backtest**: Run di data historical
5. **Analisis**: Review metrik, equity curve, distribusi trade
6. **Iterasi**: Refine parameter, risk management
7. **Validasi**: Walk-forward test di unseen data
8. **Paper Trade**: Test real-time (simulated)
9. **Deploy**: Live trading dengan alokasi modal

---

## Contoh Strategi

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

    // Golden cross: fast MA cross di atas slow MA
    if fastVal > slowVal && s.fast.Prev() <= s.slow.Prev() {
        return strategy.SignalBuy
    }

    // Death cross: fast MA cross di bawah slow MA
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

## Contoh Konfigurasi

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
  position_size: 0.95  # % modal per trade
  stop_loss: 0.02      # 2%
  take_profit: 0.05    # 5%

data:
  source: binance
  cache_dir: ./data

output:
  results_dir: ./results
  format: html  # html, json, csv
```

**Note:** Config structure ini akan diimplementasikan menggunakan `viper` di fase awal development.

---

## Log Keputusan Kunci

### 2026-08-29: Pilihan Bahasa — Golang
**Rationale:**
- Concurrency native untuk real-time data stream
- Static typing kurangi runtime error
- Compile time cepat untuk iterasi rapid
- Single binary deployment (no runtime dependencies)
- Standard library yang kuat

### 2026-08-29: Format Storage — Parquet
**Rationale:**
- Format columnar = query time-series efisien
- Kompresi (5-10x lebih kecil dari CSV)
- Schema enforcement
- Ekosistem luas (Python, R untuk analisis)

### 2026-08-29: Documentation-First Approach
**Keputusan:** Tulis AGENT.md komprehensif sebelum coding
**Rationale:**
- Arsitektur jelas cegah scope creep
- Iterasi di design lebih mudah dari di code
- Jadi spec untuk implementasi
- Future maintainer (termasuk future-you) paham "why"

---

## Target Performa (Aspirational)

*Target awal ini akan disesuaikan berdasarkan benchmark real-world.*

- **Backtest 1 tahun 1h bar**: < 1-5 detik
- **Backtest 1 tahun 1m bar**: < 10-30 detik
- **Optimisasi parameter (100 kombinasi)**: < 2-10 menit
- **Memory footprint**: < 500 MB untuk backtest typical

**Environment benchmark:** Akan di-establish di Phase 1

---

## Strategi Testing

1. **Unit Test**: Setiap indikator, komponen strategi
2. **Integration Test**: Full backtest run dengan known outcome
3. **Property Test**: Invariant (e.g., equity never negative tanpa margin)
4. **Regression Test**: Lock in metrik untuk canonical strategy
5. **Benchmark**: Track performance regression

```bash
go test ./...              # Run semua test
go test -v ./pkg/...       # Unit test saja
go test -bench=.           # Benchmark
go test -cover ./...       # Coverage report
```

---

## Risiko & Mitigasi

### 1. Look-Ahead Bias
**Risiko:** Gunakan data masa depan di sinyal  
**Mitigasi:** Arsitektur event-driven strict, tidak ada index peeking

### 2. Survivorship Bias
**Risiko:** Hanya test di coin yang masih listed  
**Mitigasi:** Include delisted coin di dataset (jika tersedia)

### 3. Overfitting
**Risiko:** Strategi work di historical, fail di live  
**Mitigasi:** Walk-forward validation, out-of-sample testing

### 4. Kualitas Data
**Risiko:** Missing bar, OHLCV incorrect  
**Mitigasi:** Validation pipeline, multiple data source

### 5. Slippage Underestimation
**Risiko:** Eksekusi terlalu optimis  
**Mitigasi:** Model slippage konservatif, test di market volatil

---

## Guideline Contributing

1. **Code Style**: Follow `gofmt`, gunakan `golangci-lint`
2. **Commit**: Conventional commit (feat, fix, docs, test, refactor)
3. **PR**: Include test, update doc jika API berubah
4. **Benchmark**: Required untuk performance-critical code

---

## Resource

### Buku
- *Advances in Financial Machine Learning* — Marcos López de Prado
- *Algorithmic Trading* — Ernie Chan
- *Quantitative Trading* — Ernie Chan

### Paper
- "The Deflated Sharpe Ratio" (Bailey & López de Prado)
- "Backtesting" (Campbell Harvey)

### Komunitas
- QuantConnect Forum
- /r/algotrading
- Golang #trading Slack

### Similar Project (Referensi)
- **backtrader** (Python): mature, ecosystem luas
- **VectorBT** (Python): cepat, vectorized
- **Jesse** (Python): khusus crypto
- **Comparison**: Kita build dari scratch untuk full control, zero bloat, dan Go performance

---

## Status Saat Ini

**Phase:** 0 (Pre-development)  
**Progress:** Struktur proyek established, arsitektur designed  
**Next Milestone:** Phase 1 — Implementasi data pipeline  
**Blocker:** Tidak ada  
**Last Activity:** 2026-08-29

---

## License

Proprietrary (akan jadi MIT setelah stabilisasi)

---

**Last Updated:** 2026-08-29  
**Maintainer:** zeroxx 🦀
