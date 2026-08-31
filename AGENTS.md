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
- [x] Advanced Fees & Slippage models

#### Sub-phase 1.3: Metrics & Reporting
- [x] PnL, Sharpe, Max Drawdown, Sortino
- [x] JSON trade log & metrics report
- [x] HTML Report builder

---

### Phase 2: Rich Strategy Framework (AI Building Blocks)
**Objective: Sediakan 'Lego Blocks' agar AI bisa merakit strategi kompleks**

#### Sub-phase 2.1: Technical Indicators Library
- [x] Math primitives (SMA, EMA, RSI, MACD, ATR, Bollinger)
- [x] Custom window functions (Rolling min/max, standard deviation)
- [x] Caching mekanisme indikator agar kalkulasi efisien

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
- [x] **Orchestrator CLI** - Automated end-to-end execution
- [x] **Context Window Management** - Focused prompts to prevent hallucination

#### Sub-phase 3.2: Analytical Feedback Loop
- [x] Parser `results.json` ke format ringkas untuk prompt AI
- [x] Framework evaluasi hipotesa (bandingkan ekspektasi vs realita backtest)
- [x] Memory state untuk menyimpan "insight riset"
- [x] **Research Memory Database** - SQLite persistent storage for learning
- [x] **Structured Feedback Format** - JSON schema for AI consumption

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
- [x] Live execution bridge
- [x] Alerting & Kill switches

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

---

## AI Workflow Rules (STRICT - Anti-Hallucination)

### Context Window Management
AI agent memiliki konteks window terbatas. Untuk menghindari halusinasi dan mempertahankan fokus:

#### Rule 1: One Task, One Context
- **SETIAP FASE** lifecycle harus dijalankan dalam konteks terpisah
- Jangan mencampur CONCEIVE + WRITE dalam satu prompt
- Jangan mencampur BACKTEST + ANALYZE dalam satu execution
- Gunakan file intermediary untuk pass data antar fase

#### Rule 2: Explicit State Management
- AI WAJIB membaca state dari file sebelum melanjutkan:
  - `research_logs/<strategy_id>/hypothesis.md` - Apa yang ingin dicapai
  - `research_logs/<strategy_id>/strategy.go` - Kode strategi saat ini
  - `research_logs/<strategy_id>/results.json` - Hasil backtest terakhir
  - `research_logs/<strategy_id>/memory.json` - Pembelajaran masa lalu
- Jangan assume/remember dari percakapan sebelumnya
- Selalu verify dengan read file actual

#### Rule 3: Incremental Refinement
- Setiap iterasi hanya boleh mengubah **SATU ASPEK** dari strategi:
  - Iterasi 1: Tambah filter trend (ADX)
  - Iterasi 2: Adjust position sizing
  - Iterasi 3: Tambah stop loss
- DILARANG rewrite seluruh strategi sekaligus
- Setiap perubahan harus traceable dan reversible

#### Rule 4: Validation Gates
Setiap output AI harus melewati gate sebelum lanjut:
- **WRITE → LINT**: AST validation wajib pass
- **LINT → TEST**: Unit test wajib pass (no panic)
- **TEST → BACKTEST**: Compile wajib sukses
- **BACKTEST → ANALYZE**: Results.json wajib exist dan valid JSON
- Jika gate gagal, AI harus fix di konteks yang sama sebelum lanjut

#### Rule 5: Structured Output Format
AI output harus mengikuti schema ketat:

**CONCEIVE Phase Output:**
```markdown
# Hypothesis: [Nama Strategi]

## Market Observation
- [Observasi 1]
- [Observasi 2]

## Core Thesis
[Penjelasan hipotesa dalam 2-3 kalimat]

## Expected Edge
- Entry: [Kapan buy/sell]
- Exit: [Kapan close]
- Risk: [Bagaimana manage risk]

## Success Criteria
- Sharpe Ratio > 1.5
- Max Drawdown < 15%
- Win Rate > 45%
```

**ANALYZE Phase Output:**
```json
{
  "strategy_id": "rsi_mean_reversion_v1",
  "backtest_results": {
    "total_return": 0.234,
    "sharpe_ratio": 1.82,
    "max_drawdown": 0.12,
    "win_rate": 0.52
  },
  "hypothesis_validation": {
    "thesis_confirmed": true,
    "evidence": ["Sharpe > 1.5 achieved", "Drawdown within limit"]
  },
  "identified_weaknesses": [
    "Consecutive losses during strong trends",
    "Entry timing suboptimal in ranging market"
  ],
  "next_iteration_plan": {
    "focus": "Add trend filter (ADX > 25)",
    "rationale": "Avoid mean reversion in trending markets",
    "expected_improvement": "Reduce consecutive losses by 30%"
  }
}
```

#### Rule 6: Memory Persistence
Setiap iterasi AI WAJIB update memory file:

```json
{
  "strategy_lineage": [
    {
      "version": "v1",
      "date": "2026-08-30",
      "changes": "Initial RSI mean reversion",
      "result": {"sharpe": 1.2, "drawdown": 0.18},
      "lesson": "RSI alone insufficient, trending markets cause losses"
    },
    {
      "version": "v2",
      "date": "2026-08-30",
      "changes": "Added ADX trend filter",
      "result": {"sharpe": 1.82, "drawdown": 0.12},
      "lesson": "Trend filter significantly improved performance"
    }
  ],
  "learned_patterns": [
    "RSI mean reversion works best in sideways markets (ADX < 25)",
    "Position sizing should scale with volatility (ATR)"
  ],
  "failed_approaches": [
    "Fixed stop loss (5%) - too tight, stopped out prematurely",
    "RSI period 14 - too slow, late entries"
  ]
}
```

#### Rule 7: Error Recovery Protocol
Jika AI encounter error:
1. **Compile Error**: Read compiler output, fix syntax only, re-validate
2. **AST Validation Error**: Read validator output, remove unsafe imports/code
3. **Test Failure**: Read test output, fix logic bug, re-run test
4. **Backtest Error**: Read error log, check data availability or logic error
5. **Max 3 attempts** per gate - jika masih gagal, flag for human review

---

## Orchestrator CLI Specification

Orchestrator adalah automation layer yang menjalankan full lifecycle:

### Command Structure
```bash
# Initialize new research session
go run cmd/orchestrator/main.go init \
  --strategy-id "rsi_mean_reversion" \
  --hypothesis-file "hypothesis.md"

# Run full cycle (WRITE → LINT → TEST → BACKTEST → ANALYZE)
go run cmd/orchestrator/main.go run \
  --strategy-id "rsi_mean_reversion" \
  --data-file "data/BTCUSDT_1h.json" \
  --ai-model "gpt-4"

# Run single phase (for debugging)
go run cmd/orchestrator/main.go phase \
  --strategy-id "rsi_mean_reversion" \
  --phase "BACKTEST"

# View research history
go run cmd/orchestrator/main.go history \
  --strategy-id "rsi_mean_reversion"
```

### Orchestrator Responsibilities
1. **State Management**: Create/read/update research_logs directory structure
2. **Phase Execution**: Call appropriate Go packages (validator, backtest, analyzer)
3. **Gate Enforcement**: Block progression if validation fails
4. **Context Isolation**: Each phase runs in clean environment
5. **Error Logging**: Structured error output for AI consumption
6. **Memory Updates**: Auto-update memory.json after each cycle

### Directory Structure
```
research_logs/
├── rsi_mean_reversion/
│   ├── hypothesis.md          # Initial thesis
│   ├── strategy_v1.go         # Version 1 code
│   ├── strategy_v2.go         # Version 2 code (after refinement)
│   ├── strategy_current.go    # Symlink to latest version
│   ├── results_v1.json        # Backtest results v1
│   ├── results_v2.json        # Backtest results v2
│   ├── memory.json            # Accumulated learning
│   ├── validation_errors.log  # AST/Test errors (if any)
│   └── analysis_v2.json       # Structured analysis output
└── sma_crossover/
    └── ...
```

---

## AI Prompt Templates

### CONCEIVE Phase Prompt
```
You are a quantitative researcher. Based on market observation, formulate a trading hypothesis.

Context:
- Market: Cryptocurrency (BTCUSDT)
- Timeframe: 1 hour
- Available indicators: SMA, EMA, RSI, MACD, ATR, Bollinger Bands
- Risk management: Position sizing, stop loss, trailing stop

Task:
Write a hypothesis.md file following the structured format.
Focus on ONE clear edge. Do not write code yet.

Output: hypothesis.md content only.
```

### WRITE Phase Prompt
```
You are a Go developer implementing a trading strategy.

Context:
- Read hypothesis from: research_logs/{strategy_id}/hypothesis.md
- Implement interface: pkg/sdk.Strategy (Init, OnBar methods)
- Use SDK methods: ctx.Buy(), ctx.Sell(), ctx.CloseAll()
- Use indicators: ctx.Indicator.SMA(), ctx.Indicator.RSI(), etc.

Constraints:
- No imports beyond SDK and indicators packages
- No goroutines, no channels
- Deterministic logic only
- Add // RATIONALE: comments explaining logic

Task:
Write complete strategy.go file implementing the hypothesis.

Output: Go code only, no markdown fences.
```

### ANALYZE Phase Prompt
```
You are a quantitative analyst evaluating backtest results.

Context:
- Strategy hypothesis: [read from hypothesis.md]
- Backtest results: [read from results.json]
- Previous iterations: [read from memory.json]

Task:
1. Validate if hypothesis was confirmed or rejected
2. Identify specific weaknesses (e.g., "losses during trend", "late entries")
3. Propose ONE focused improvement for next iteration
4. Update memory.json with learnings

Output: analysis.json in structured format.
```

---

## Success Metrics for AI Pipeline

### Phase 3 Complete When:
- [ ] Orchestrator CLI dapat run full cycle tanpa manual intervention
- [ ] AI dapat generate valid strategy code (pass AST validation) in 1-2 attempts
- [ ] AI dapat analyze results dan propose refinement dengan reasoning clear
- [ ] Memory system dapat track 5+ iterations dengan lineage jelas
- [ ] Walk-forward test menunjukkan AI tidak overfit (IS/OOS gap < 10%)
- [ ] Documentation lengkap untuk setup AI agent (LLM API keys, prompts)

### Anti-Hallucination Validation:
- [ ] AI tidak assume data yang tidak ada di file
- [ ] AI tidak generate random parameters tanpa reasoning
- [ ] AI tidak claim performance improvement tanpa backtest evidence
- [ ] AI dapat recover dari errors dalam max 3 attempts
- [ ] AI learning terdokumentasi dan traceable di memory.json