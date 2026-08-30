# AGENTS.md — backtest-go Project Guide

## Vision

Backtest framework ini adalah **research infrastructure** interaktif. AI Agent bukan sekadar optimizer parameter (grid searcher), melainkan **Quantitative Researcher & Code Creator**. AI bertugas meriset pasar, merumuskan hipotesa, dan **menulis kode strategi (Go)** yang kompleks dari nol menggunakan SDK framework, memvalidasinya, dan belajar dari hasil backtest untuk iterasi riset selanjutnya.

## Core Principles

### 1. Limitless Strategy Creation (AI-Driven)
- AI bebas berkreasi menulis logika strategi, indikator kustom, dan rule trading dalam bentuk kode Go murni.
- Kebebasan logika dibatasi oleh **Strategy SDK Context** agar aman dan tidak merusak sistem eksekusi core.

### 2. Strict Safety Boundaries
- Kode strategi buatan AI divalidasi secara ketat via AST (Abstract Syntax Tree) sebelum di-compile.
- Larangan keras pada import `os`, `net`, `go routines` di layer strategi.
- Automated unit testing wajib lulus sebelum kode diizinkan masuk pipeline backtest.

### 3. Continuous Learning Loop
- AI membaca log kegagalan dan kesuksesan dari metrik.
- AI menyimpan memori strategis: apa yang berhasil di market tertentu, apa yang gagal dan alasannya (overfitting, logic error).

### 4. Accuracy & Reproducibility
- Core backtest engine murni berjalan di Go tanpa intervensi AI selama kalkulasi.
- Eksekusi order, slippage, dan PnL adalah kebenaran mutlak (Source of Truth).

### 5. CI/CD with CircleCI (STRICT RULE)
- **STRICT:** DILARANG RUNNING TEST LOKAL. Setiap perubahan WAJIB di-commit & push ke GitHub.
- Biarkan CircleCI yang melakukan test, linting, dan benchmarking.
- AI mengecek status CircleCI via CLI (`gh run list`, `gh run view`).
- Every PR/Commit must pass all CircleCI checks.
- **WAJIB:** Setiap selesai sub-phase, COMMIT & PUSH ke GitHub untuk trigger CircleCI.
- **LAPORAN:** Completion report dibuat per PHASE (bukan per sub-phase) setelah semua sub-phase selesai.

---

## Development Phases

### Phase 0: Foundation & Research
**Objective: Establish solid ground before coding**

#### Sub-phase 0.1: Documentation & Methodology
- [x] Document research methodology (AI Code Gen paradigm)
- [x] Design system architecture (Strategy SDK + AI Code Gen)
- [x] Establish coding standards (AI restricted code rules)
- [x] CircleCI automated testing pipeline

#### Sub-phase 0.2: Exchange API Research
- [x] API endpoint & limitation documentation
- [x] Test Binance vs Bybit
- [x] Define normalization format

#### Sub-phase 0.3: Data Quality Framework
- [x] OHLCV Validator
- [x] Completeness & Gaps detection
- [x] Outlier consistency checks

---

### Phase 1: Core Backtest Engine
**Objective: Accurate, reliable backtest foundation**

#### Sub-phase 1.1: Data Pipeline
- [x] Binance/Bybit historical fetcher
- [x] Data normalizer & local JSON storage
- [x] Ingestion validator integration

#### Sub-phase 1.2: Backtest Core & SDK Context
- [x] Event-driven Engine (`OnBar`, `OnTick`)
- [x] **Strategy Context SDK** (Sandboxed API: `ctx.Buy()`, `ctx.CloseAll()`)
- [x] Execution Simulator (basic implementation)
- [ ] Advanced Fees & Slippage models (deferred to optimization phase)

#### Sub-phase 1.3: Metrics & Reporting
- [x] PnL, Sharpe, Max Drawdown, Sortino
- [x] JSON trade log & metrics report
- [ ] HTML Report builder (deferred to Phase 4)

---

### Phase 2: Rich Strategy Framework (AI Building Blocks)
**Objective: Sediakan 'Lego Blocks' agar AI bisa merakit strategi kompleks**

#### Sub-phase 2.1: Technical Indicators Library
- [x] Math primitives (SMA, EMA, RSI, MACD, ATR, Bollinger)
- [ ] Custom window functions (Rolling min/max, standard deviation) — deferred to Phase 2.2
- [ ] Caching mekanisme indikator agar kalkulasi efisien — deferred to Phase 4 (optimization)

#### Sub-phase 2.2: Signal & Risk Management Primitives
- [x] Position Sizing primitives (Fixed fractional, Kelly criterion)
- [x] Stop Loss & Trailing Stop helpers (Fixed, Percent, ATR-based, Trailing)
- [x] Multi-timeframe helper (timeframe conversion, aggregation, alignment)

#### Sub-phase 2.3: Safe Code Validation System
- [x] AST-based Go code linter (detects unsafe imports, goroutines, syscalls)
- [x] Auto-generate strategy unit test template

---

### Phase 3: AI Researcher Integration Layer
**Objective: Hubungkan AI sebagai programmer & data scientist otonom**

#### Sub-phase 3.1: Code Generation Pipeline
- [x] System prompt untuk AI code generator
- [x] Pipeline: AI tulis `.go` -> CLI baca -> Lint AST -> Test -> Backtest
- [x] Error feedback loop (compile errors, validation errors)

#### Sub-phase 3.2: Analytical Feedback Loop
- [x] Parser `results.json` ke format ringkas untuk prompt AI
- [x] Framework evaluasi hipotesa (bandingkan ekspektasi vs realita backtest)
- [x] Memory state untuk menyimpan "insight riset"

#### Sub-phase 3.3: Overfitting Prevention
- [x] Walk-forward test orchestrator
- [x] In-sample vs Out-of-sample gap analyzer
- [x] AI prompt khusus untuk mendeteksi curve-fitting

---

### Phase 4: Scaling & Live Trading
**Objective: Optimisasi massal dan deploy**

#### Sub-phase 4.1: Mass Optimization
- [x] Parallel strategy execution (test ratusan kode AI serentak)
- [x] Parameter space search assistance

#### Sub-phase 4.2: Real-time Simulation (Paper Trading)
- [x] WebSocket market data listener
- [x] Paper trading execution state

#### Sub-phase 4.3: Deployment Automation
- [ ] Live execution bridge
- [ ] Alerting & Kill switches

---

## AI Agent Guidelines (As Code Creator)

### DO:
- ✅ Tulis strategi menggunakan SDK murni (`ctx.Indicator.SMA()`, `ctx.MarketOrder()`).
- ✅ Bebas berkreasi dengan state lokal (struct fields) untuk membuat custom kalkulasi (misal: hitung korelasi harga internal).
- ✅ Analisa hasil secara objektif, jika jelek, hapus logikanya dan tulis kode pendekatan baru (pivot).
- ✅ Tambahkan komentar `// RATIONALE:` di kode strategi agar manusia paham niat dari kode tersebut.
- ✅ Ingat pembelajaran masa lalu (misal: "RSI default sering false breakout saat trending, saya akan buat RSI dengan filter ADX").

### DON'T:
- ❌ DILARANG import package standar seperti `os`, `net/http`, `io/ioutil`, `syscall`.
- ❌ DILARANG menggunakan goroutines (`go func()`) atau channel internal di dalam logika strategi. Algoritma harus sinkron dan deterministik.
- ❌ Jangan asumsikan harga akan tereksekusi pas di harga `Close` (Engine akan memasukkan slippage).
- ❌ Jangan terpaku pada satu indikator jika terbukti gagal di out-of-sample.

---

## AI Strategy Code Lifecycle:

```
1. CONCEIVE : AI riset teori, tulis definisi hipotesa (Markdown).
2. WRITE    : AI tulis file `strategy_name.go` mengimplementasikan interface `Strategy`.
3. LINT     : Backtest-go melakukan AST check (Tolak jika ada unsafe import/logic).
4. TEST     : Backtest-go auto-generate/run unit test. Jika panic/fail -> AI perbaiki.
5. BACKTEST : Engine eksekusi historis.
6. ANALYZE  : AI baca JSON metric & equity curve. Cari kelemahan (misal: "Drawdown besar saat sideways").
7. REFINE   : AI memodifikasi kode `.go` untuk menambahkan filter/logic baru.
```