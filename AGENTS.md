# AGENTS.md — backtest-go Project Guide

## Vision

Backtest framework bukan hanya alat testing biasa — ini adalah **research infrastructure** untuk menemukan, meneliti, dan mengoptimasi strategy trading yang dapat diotomatisasi. AI Agent berperan sebagai researcher yang sistematis, bukan oracle yang menebak.

## Core Principles

### 1. Accuracy First
**Backtest = Source of Truth**
- Backtest accuracy adalah fondasi — garbage in, garbage out
- Setiap metric harus verifiable dan reproducible
- Data quality checks BEFORE backtest execution
- Objective metrics > subjective assessment

### 2. Structured AI Integration
**AI sebagai Researcher, bukan Oracle**
- AI menerima **structured input** (backtest results, metrics, constraints)
- AI menghasilkan **structured output** (hypothesis, strategy parameters)
- Panduan baku untuk implementasi strategy (no hallucination)
- Clear protocol untuk research iteration

### 3. Phased Development
**Divide & Conquer**
- Complex system → multiple phases & sub-phases
- Focus on one feature per sub-phase
- Integration testing at end of each phase
- Objective validation tools (not just AI judgment)

### 4. CI/CD with CircleCI
**Automated Testing Pipeline**
- CircleCI untuk automated testing (hardware resource constraints)
- Every PR must pass all tests before merge
- Automated data quality checks
- Performance benchmarking per commit

---

## Development Phases

### Phase 0: Foundation & Research
**Objective: Establish solid ground before coding**

Phase ini adalah fase persiapan yang KRITIS. Tidak boleh ada coding untuk backtest engine sampai fase ini selesai. Fokus pada research, dokumentasi, dan tool preparation.

#### Sub-phase 0.1: Documentation & Methodology
**Duration: 2-3 days**

**Tasks:**
1. **Create `docs/methodology.md`**
   - Research methodology (hypothesis-driven testing)
   - Success metrics definition
   - Validation criteria per phase
   - Failure recovery protocol

2. **Create `docs/architecture.md`**
   - High-level system design
   - Component interaction diagram
   - Data flow architecture
   - AI integration points

3. **Create `docs/coding-standards.md`**
   - Go coding conventions
   - Error handling patterns
   - Testing requirements (min 80% coverage)
   - Documentation standards

4. **Create `.circleci/config.yml`**
   - Basic CircleCI workflow
   - Go test runner setup
   - Linting (golangci-lint)
   - Coverage reporting

**Deliverables:**
- `docs/methodology.md` (min 1000 words)
- `docs/architecture.md` with diagrams
- `docs/coding-standards.md`
- `.circleci/config.yml` (working pipeline)

**Exit Criteria:**
- ✅ All documentation files exist and reviewed
- ✅ CircleCI pipeline runs successfully (even with empty tests)
- ✅ Architecture approved by user
- ✅ Clear definition of "done" for each phase

---

#### Sub-phase 0.2: Exchange API Research
**Duration: 3-4 days**

**Objective:** Thoroughly research and document exchange APIs to ensure data quality.

**Tasks:**
1. **Binance API Research**
   - Latest REST API endpoints for historical data
   - WebSocket API for real-time data
   - Rate limits and restrictions
   - Data granularity options (1m, 5m, 15m, 1h, 4h, 1d)
   - Test API reliability (uptime, response time)
   - Document breaking changes from older versions

2. **Bybit API Research**
   - Same as Binance above
   - Compare data format differences
   - Identify normalization requirements

3. **Data Quality Validation**
   - Fetch sample data from both exchanges
   - Check for gaps/missing candles
   - Verify OHLCV accuracy
   - Compare cross-exchange data consistency
   - Document known data issues

4. **Create `docs/exchange-api.md`**
   - Complete API reference
   - Rate limit strategies
   - Error handling patterns
   - Data format specifications
   - Sample request/response

5. **Create Test Scripts**
   - `scripts/test-binance-api.go` — verify Binance API
   - `scripts/test-bybit-api.go` — verify Bybit API
   - `scripts/compare-exchange-data.go` — cross-validate data

**Deliverables:**
- `docs/exchange-api.md` (comprehensive reference)
- Working test scripts in `scripts/`
- Data quality report (markdown document)
- Recommendation for primary exchange (based on data quality)

**Exit Criteria:**
- ✅ Exchange APIs thoroughly documented
- ✅ Test scripts successfully fetch and validate data
- ✅ Data quality issues identified and documented
- ✅ Clear understanding of rate limits and constraints
- ✅ Normalization requirements defined

---

#### Sub-phase 0.3: Data Quality Framework
**Duration: 2-3 days**

**Objective:** Build validation tools to ensure data quality before backtest execution.

**Tasks:**
1. **Design Data Validation Schema**
   - Define required fields for OHLCV data
   - Specify data type constraints
   - Define acceptable ranges (e.g., volume > 0)
   - Create validation rules

2. **Implement Validation Tools**
   - `internal/validator/ohlcv.go` — OHLCV data validator
   - `internal/validator/completeness.go` — check for gaps
   - `internal/validator/consistency.go` — outlier detection
   - `internal/validator/accuracy.go` — sanity checks

3. **Create Data Quality Metrics**
   - Completeness score (% of expected data points)
   - Consistency score (outlier rate)
   - Freshness score (data staleness)
   - Overall quality score (weighted average)

4. **Build Data Storage Format**
   - Choose storage format (CSV, Parquet, or SQLite)
   - Design schema for local caching
   - Implement compression strategy
   - Define retention policy

5. **Create `docs/data-quality.md`**
   - Validation methodology
   - Quality metrics explanation
   - Threshold definitions
   - Troubleshooting guide

**Deliverables:**
- Data validation package in `internal/validator/`
- Quality metrics implementation
- Storage format specification
- `docs/data-quality.md`
- Unit tests for validators (100% coverage)

**Exit Criteria:**
- ✅ Validators can detect common data issues
- ✅ Quality metrics produce meaningful scores
- ✅ Storage format chosen and justified
- ✅ All validator tests pass
- ✅ CircleCI runs validators automatically

---

**Phase 0 Complete Exit Criteria:**
- ✅ All sub-phase deliverables completed
- ✅ Documentation reviewed and approved
- ✅ Exchange API research validated with real data
- ✅ Data quality framework working
- ✅ CircleCI pipeline green
- ✅ Ready to start Phase 1 implementation

**DO NOT proceed to Phase 1 until ALL Phase 0 criteria met.**

---

### Phase 1: Core Backtest Engine
**Objective: Accurate, reliable backtest foundation**

Phase ini membangun backtest engine yang akurat. Fokus pada correctness > performance. Optimisasi dilakukan di fase selanjutnya.

#### Sub-phase 1.1: Data Pipeline
**Duration: 4-5 days**

**Objective:** Build reliable data fetching and normalization layer.

**Tasks:**
1. **Exchange Client Implementation**
   - `internal/exchange/binance/client.go` — Binance HTTP client
   - `internal/exchange/bybit/client.go` — Bybit HTTP client
   - Implement retry logic with exponential backoff
   - Handle rate limiting automatically
   - Implement connection pooling

2. **Data Fetcher**
   - `internal/datafetcher/fetcher.go` — main fetcher interface
   - `internal/datafetcher/historical.go` — fetch historical OHLCV
   - Implement parallel fetching with concurrency limit
   - Add progress tracking
   - Handle pagination for large date ranges

3. **Data Normalizer**
   - `internal/normalizer/ohlcv.go` — normalize to standard format
   - Convert timestamps to consistent timezone (UTC)
   - Standardize field names
   - Handle exchange-specific quirks

4. **Local Cache Implementation**
   - `internal/cache/storage.go` — local data storage
   - `internal/cache/reader.go` — read cached data
   - `internal/cache/writer.go` — write data to cache
   - Implement cache invalidation strategy
   - Add compression (if using files)

5. **Integration with Validators**
   - Run validators on fetched data before caching
   - Reject low-quality data automatically
   - Generate data quality reports

**Deliverables:**
- Complete data pipeline in `internal/`
- CLI tool: `cmd/fetch-data/main.go` for manual data fetching
- Integration tests for each component
- CircleCI tests for data pipeline
- Performance benchmarks (throughput, memory usage)

**Exit Criteria:**
- ✅ Can fetch historical data from exchanges
- ✅ Data normalized to standard format
- ✅ Data cached locally with good performance
- ✅ Data quality validated automatically
- ✅ All tests pass in CircleCI
- ✅ No data corruption detected in tests

---

#### Sub-phase 1.2: Backtest Core
**Duration: 5-7 days**

**Objective:** Implement accurate order execution simulation and position tracking.

**Tasks:**
1. **Backtest State Machine**
   - `internal/backtest/engine.go` — main backtest engine
   - `internal/backtest/state.go` — maintain backtest state
   - Implement time progression (bar-by-bar)
   - Handle strategy signals
   - Manage order queue

2. **Order Execution Simulator**
   - `internal/execution/simulator.go` — simulate order fills
   - Implement realistic fill logic:
     - Market orders: fill at next bar open
     - Limit orders: fill when price touches limit
     - Stop orders: trigger and fill logic
   - Handle partial fills (for illiquid markets)
   - Simulate slippage based on volume

3. **Position Tracking**
   - `internal/position/tracker.go` — track open positions
   - Calculate unrealized P&L
   - Update position on fills
   - Handle multiple positions (if supported)
   - Track average entry price

4. **Fee & Slippage Calculation**
   - `internal/execution/fees.go` — calculate trading fees
   - Support maker/taker fee tiers
   - Implement configurable fee rates
   - `internal/execution/slippage.go` — slippage model
   - Volume-based slippage estimation

5. **Trade Logger**
   - `internal/backtest/logger.go` — log all trades
   - Record: timestamp, side, price, quantity, fees, P&L
   - Support JSON/CSV output formats

**Deliverables:**
- Complete backtest engine in `internal/backtest/`
- Order execution simulator with realistic fill logic
- Position tracker with accurate P&L calculation
- Fee and slippage models
- Trade logger
- Unit tests for each component (80%+ coverage)
- Integration tests for full backtest flow

**Exit Criteria:**
- ✅ Backtest engine executes strategies correctly
- ✅ Order fills are realistic and reproducible
- ✅ Position tracking accurate (verified manually)
- ✅ Fees and slippage calculated correctly
- ✅ All tests pass (unit + integration)
- ✅ No race conditions (run with -race flag)

---

#### Sub-phase 1.3: Metrics & Reporting
**Duration: 3-4 days**

**Objective:** Implement comprehensive performance metrics and reporting.

**Tasks:**
1. **Core Metrics Implementation**
   - `internal/metrics/pnl.go` — P&L calculation
   - `internal/metrics/returns.go` — returns (absolute, percentage)
   - `internal/metrics/sharpe.go` — Sharpe ratio
   - `internal/metrics/sortino.go` — Sortino ratio
   - `internal/metrics/drawdown.go` — max drawdown
   - `internal/metrics/win_rate.go` — win rate, profit factor

2. **Metric Validators**
   - `internal/metrics/validator.go` — validate metric calculations
   - Compare against known benchmarks
   - Implement sanity checks (e.g., Sharpe in reasonable range)

3. **Report Generator**
   - `internal/report/generator.go` — generate reports
   - Support multiple formats: JSON, CSV, HTML
   - Include equity curve data
   - Trade-by-trade breakdown
   - Summary statistics

4. **Visualization Tools** (optional, nice-to-have)
   - Generate equity curve charts
   - Drawdown visualization
   - Monthly returns heatmap

5. **Benchmark Comparisons**
   - `internal/benchmark/buyhold.go` — buy-and-hold benchmark
   - Compare strategy vs benchmark
   - Calculate alpha (excess returns)

**Deliverables:**
- Complete metrics package in `internal/metrics/`
- Report generator with multiple output formats
- Benchmark comparison tools
- CLI tool: `cmd/backtest/main.go` for running backtests
- Metric validation tests (with known results)
- CircleCI integration for metric tests

**Exit Criteria:**
- ✅ All core metrics implemented correctly
- ✅ Metrics validated against known benchmarks
- ✅ Reports generated successfully
- ✅ Buy-and-hold comparison works
- ✅ All tests pass
- ✅ Can run full backtest end-to-end

---

**Phase 1 Complete Exit Criteria:**
- ✅ Data pipeline works reliably
- ✅ Backtest engine accurate and reproducible
- ✅ All metrics validated
- ✅ Can run backtests and generate reports
- ✅ All CircleCI tests green
- ✅ Performance acceptable (not optimized yet, but usable)
- ✅ Ready for Phase 2 (strategy framework)

---

### Phase 2: Strategy Framework
**Objective: Standardized, non-ambiguous strategy implementation**

Phase ini membangun framework untuk strategy implementation yang baku dan tidak ambigu. AI akan menggunakan framework ini untuk generate strategies.

#### Sub-phase 2.1: Strategy Interface
**Duration: 3-4 days**

**Objective:** Define clear, unambiguous interface for strategies.

**Tasks:**
1. **Strategy Interface Design**
   - `internal/strategy/interface.go` — strategy interface definition
   ```go
   type Strategy interface {
       Initialize(config Config) error
       OnBar(bar OHLCV, position Position) Signal
       OnTrade(trade Trade)
       Teardown()
   }
   ```
   - Clear lifecycle hooks
   - No ambiguity in method signatures

2. **Parameter Schema**
   - `internal/strategy/params.go` — parameter definition
   - Type-safe parameters (no interface{})
   - Parameter validation
   - Default values support
   - Bounds checking (min/max ranges)

3. **Signal Types**
   - `internal/strategy/signal.go` — signal definitions
   - BUY, SELL, HOLD signals
   - Optional: signal strength (0.0-1.0)
   - Order type specification (market, limit, stop)

4. **Strategy Validator**
   - `internal/strategy/validator.go` — validate strategy implementations
   - Check interface compliance
   - Verify parameter bounds
   - Test strategy initialization
   - Detect common errors (e.g., divide by zero)

5. **Example Strategies**
   - `strategies/sma_crossover.go` — simple moving average crossover
   - `strategies/rsi_mean_reversion.go` — RSI mean reversion
   - `strategies/bollinger_bands.go` — Bollinger Bands breakout
   - Each with full documentation and tests

**Deliverables:**
- Strategy interface in `internal/strategy/`
- Parameter schema and validation
- Signal type definitions
- Strategy validator
- 3 example strategies in `strategies/`
- `docs/strategy-guide.md` — how to write strategies
- Unit tests for all components

**Exit Criteria:**
- ✅ Strategy interface clear and well-documented
- ✅ No ambiguity in interface methods
- ✅ Parameter validation working
- ✅ Example strategies run successfully
- ✅ Validator catches common errors
- ✅ All tests pass

---

#### Sub-phase 2.2: Strategy Execution
**Duration: 4-5 days**

**Objective:** Integrate strategies with backtest engine.

**Tasks:**
1. **Strategy Runner**
   - `internal/backtest/runner.go` — execute strategy in backtest
   - Call strategy.OnBar() for each bar
   - Handle strategy signals
   - Pass trades back to strategy via OnTrade()

2. **State Management**
   - Strategy can maintain internal state
   - State persisted across bars
   - Optional: snapshot/restore for debugging

3. **Error Handling**
   - Graceful handling of strategy errors
   - Continue backtest on recoverable errors
   - Log errors for debugging
   - Fail-fast on critical errors

4. **Performance Optimization**
   - Minimize allocations in hot path
   - Benchmark strategy execution overhead
   - Profile memory usage

5. **Strategy Testing Framework**
   - `internal/strategy/testing/` — helpers for strategy tests
   - Mock data generators
   - Assertion helpers
   - Reproducible test scenarios

**Deliverables:**
- Strategy runner integrated with backtest engine
- Error handling implementation
- Strategy testing framework
- Performance benchmarks
- Updated example strategies using runner
- Integration tests

**Exit Criteria:**
- ✅ Strategies execute correctly in backtest
- ✅ Error handling robust
- ✅ Performance overhead acceptable (<5% vs raw backtest)
- ✅ Testing framework usable
- ✅ All tests pass

---

#### Sub-phase 2.3: Strategy Testing & Validation
**Duration: 3-4 days**

**Objective:** Build comprehensive testing and validation tools.

**Tasks:**
1. **Strategy Test Suite**
   - `test/strategies/` — test suite for strategies
   - Test each example strategy thoroughly
   - Edge case testing (empty data, single bar, etc.)
   - Regression tests (lock in known results)

2. **Correctness Validators**
   - `internal/validation/correctness.go` — check strategy logic
   - Verify signal generation is deterministic
   - Check for look-ahead bias
   - Detect data leakage

3. **Performance Benchmarks**
   - `internal/benchmark/strategy.go` — benchmark strategies
   - Measure execution time per bar
   - Memory allocation profiling
   - Compare strategies

4. **Strategy Linter**
   - `cmd/lint-strategy/main.go` — static analysis tool
   - Detect common mistakes
   - Check for best practices
   - Warn about potential issues

5. **CircleCI Integration**
   - Run all strategy tests automatically
   - Fail PR if strategy tests fail
   - Run benchmarks (for tracking performance)

**Deliverables:**
- Comprehensive test suite
- Correctness validators
- Performance benchmarks
- Strategy linter tool
- CircleCI workflow for strategy testing
- `docs/strategy-testing.md` — testing guide

**Exit Criteria:**
- ✅ All example strategies pass tests
- ✅ No look-ahead bias detected
- ✅ Performance benchmarks established
- ✅ Linter catches common errors
- ✅ CircleCI runs all tests successfully
- ✅ Test coverage >80%

---

**Phase 2 Complete Exit Criteria:**
- ✅ Strategy interface well-defined and documented
- ✅ Example strategies work correctly
- ✅ Strategy testing framework complete
- ✅ All validation tools working
- ✅ CircleCI green
- ✅ Ready for AI integration (Phase 3)

---

### Phase 3: AI Integration Layer
**Objective: Structured AI research capabilities**

Phase ini membangun layer untuk AI dapat melakukan research secara sistematis dengan input/output yang terstruktur dan tervalidasi.

#### Sub-phase 3.1: AI Input/Output Protocol
**Duration: 3-4 days**

**Objective:** Define structured schemas for AI interaction.

**Tasks:**
1. **Backtest Result Schema**
   - `internal/ai/schema/backtest_result.go` — AI input schema
   ```go
   type BacktestResult struct {
       Strategy      string
       Parameters    map[string]interface{}
       Metrics       MetricsSummary
       TradeLog      []Trade
       EquityCurve   []float64
       Drawdowns     []Drawdown
       Benchmark     BenchmarkComparison
   }
   ```
   - JSON serialization
   - Schema validation

2. **Strategy Hypothesis Schema**
   - `internal/ai/schema/hypothesis.go` — AI output schema
   ```go
   type StrategyHypothesis struct {
       Name          string
       Description   string
       Logic         string              // Human-readable logic
       Parameters    map[string]Parameter
       Constraints   []Constraint
       ExpectedEdge  string              // Why this might work
       RiskFactors   []string            // Potential issues
   }
   ```
   - Schema validation
   - Parameter bounds checking

3. **Strategy Code Schema**
   - `internal/ai/schema/strategy_code.go` — generated code schema
   - Template-based code generation
   - Validation rules for generated code
   - Safety checks (no arbitrary code execution)

4. **Schema Validators**
   - `internal/ai/validator/` — validate all AI schemas
   - Type checking
   - Bounds checking
   - Constraint validation

5. **Documentation**
   - `docs/ai-protocol.md` — AI integration protocol
   - Input schema specification
   - Output schema specification
   - Examples and best practices

**Deliverables:**
- Complete schema definitions in `internal/ai/schema/`
- Schema validators
- JSON serialization/deserialization
- `docs/ai-protocol.md`
- Unit tests for schemas

**Exit Criteria:**
- ✅ Schemas well-defined and documented
- ✅ Validation catches invalid inputs/outputs
- ✅ JSON serialization works correctly
- ✅ No ambiguity in schema definitions
- ✅ All tests pass

---

#### Sub-phase 3.2: AI Research Tools
**Duration: 5-6 days**

**Objective:** Build tools for AI to perform systematic research.

**Tasks:**
1. **Hypothesis Generator**
   - `internal/ai/hypothesis.go` — generate strategy hypotheses
   - Analyze backtest results
   - Identify patterns
   - Generate testable hypotheses
   - Output as StrategyHypothesis schema

2. **Parameter Optimizer**
   - `internal/optimizer/grid_search.go` — grid search optimizer
   - `internal/optimizer/random_search.go` — random search
   - Support parameter ranges
   - Parallel optimization
   - Early stopping (if results clearly bad)

3. **Result Analyzer**
   - `internal/analyzer/compare.go` — compare multiple results
   - Statistical significance testing
   - Identify best performers
   - Detect overfitting (compare in-sample vs out-of-sample)

4. **Iteration Tracker**
   - `internal/research/tracker.go` — track research iterations
   - Store all hypotheses tested
   - Store all backtest results
   - Track what worked and what didn't
   - Generate research reports

5. **AI Research CLI**
   - `cmd/ai-research/main.go` — CLI for AI research
   - Commands: generate-hypothesis, optimize, compare, report
   - Interactive mode for research sessions

**Deliverables:**
- AI research tools in `internal/ai/` and `internal/optimizer/`
- Result analyzer
- Iteration tracker
- AI research CLI tool
- Integration tests
- `docs/ai-research-guide.md`

**Exit Criteria:**
- ✅ Can generate hypotheses from backtest results
- ✅ Parameter optimization works
- ✅ Result analysis provides meaningful insights
- ✅ Iteration tracking works
- ✅ CLI tool usable
- ✅ All tests pass

---

#### Sub-phase 3.3: Validation & Safety
**Duration: 3-4 days**

**Objective:** Ensure AI-generated strategies are safe and valid.

**Tasks:**
1. **Hypothesis Validator**
   - `internal/ai/validator/hypothesis.go` — validate hypotheses
   - Check parameter bounds
   - Verify constraints are reasonable
   - Detect nonsensical logic
   - Flag potential issues

2. **Parameter Bounds Checker**
   - `internal/ai/validator/bounds.go` — enforce parameter bounds
   - Hard limits (e.g., SMA period > 0)
   - Soft limits (warn if unusual values)
   - Context-aware bounds (e.g., short SMA < long SMA)

3. **Sanity Test Suite**
   - `internal/ai/testing/sanity.go` — sanity tests
   - Test strategy on known scenarios
   - Verify it doesn't do obviously wrong things
   - Check for edge cases

4. **Hallucination Detection**
   - `internal/ai/validator/hallucination.go` — detect AI hallucinations
   - Check if generated code matches hypothesis
   - Verify parameters are within specified ranges
   - Flag inconsistencies

5. **Safety Checks**
   - `internal/ai/safety/` — safety checks
   - Prevent dangerous parameter combinations
   - Detect strategies that could lose all capital quickly
   - Risk limits (max position size, max drawdown, etc.)

**Deliverables:**
- Complete validation package in `internal/ai/validator/`
- Sanity test suite
- Hallucination detector
- Safety checks
- Unit tests for all validators
- `docs/ai-safety.md` — safety guidelines

**Exit Criteria:**
- ✅ Validators catch invalid hypotheses
- ✅ Parameter bounds enforced
- ✅ Sanity tests detect obvious errors
- ✅ Hallucination detection works
- ✅ Safety checks prevent dangerous strategies
- ✅ All tests pass

---

**Phase 3 Complete Exit Criteria:**
- ✅ AI protocol well-defined and documented
- ✅ AI research tools working
- ✅ Validation catches errors and hallucinations
- ✅ Safety checks prevent dangerous strategies
- ✅ CircleCI tests green
- ✅ Ready for optimization (Phase 4)

---

### Phase 4: Optimization & Production Readiness
**Objective: Systematic strategy improvement and production preparation**

Phase terakhir fokus pada optimisasi strategy, research workflow, dan persiapan untuk production (paper trading / live trading).

#### Sub-phase 4.1: Advanced Optimization
**Duration: 4-5 days**

**Objective:** Implement sophisticated optimization techniques.

**Tasks:**
1. **Walk-Forward Testing**
   - `internal/optimizer/walkforward.go` — walk-forward optimization
   - Split data into in-sample and out-of-sample
   - Optimize on in-sample, validate on out-of-sample
   - Roll forward and repeat
   - Detect overfitting automatically

2. **Genetic Algorithm Optimizer** (optional, nice-to-have)
   - `internal/optimizer/genetic.go` — genetic algorithm
   - More efficient than grid search
   - Can explore larger parameter spaces

3. **Overfitting Detection**
   - `internal/validation/overfitting.go` — detect overfitting
   - Compare in-sample vs out-of-sample results
   - Statistical tests (t-test, etc.)
   - Flag suspicious results

4. **Multi-Objective Optimization** (optional)
   - Optimize for multiple metrics (Sharpe + drawdown)
   - Pareto frontier analysis

5. **Optimization Reports**
   - Generate detailed optimization reports
   - Show parameter sensitivity
   - Visualize optimization landscape

**Deliverables:**
- Walk-forward testing implementation
- Overfitting detector
- Optimization report generator
- CLI tool: `cmd/optimize/main.go`
- Unit and integration tests

**Exit Criteria:**
- ✅ Walk-forward testing works correctly
- ✅ Overfitting detected reliably
- ✅ Optimization reports useful
- ✅ All tests pass

---

#### Sub-phase 4.2: Research Workflow
**Duration: 3-4 days**

**Objective:** Build end-to-end research workflow.

**Tasks:**
1. **Research Pipeline**
   - `internal/research/pipeline.go` — full research pipeline
   - Steps: hypothesis → backtest → optimize → validate → report
   - Automated execution
   - Progress tracking

2. **Experiment Tracking**
   - `internal/research/experiments.go` — track experiments
   - Store all experiments (successful and failed)
   - Version control for strategies
   - Reproducibility support

3. **Result Comparison Tools**
   - `internal/research/compare.go` — compare experiments
   - Side-by-side comparison
   - Statistical comparison
   - Generate comparison reports

4. **Research Dashboard** (CLI-based)
   - `cmd/dashboard/main.go` — TUI dashboard
   - View all experiments
   - Compare results
   - Track research progress

5. **Research Database**
   - Use SQLite for storing experiments
   - Schema for experiments, results, strategies
   - Query interface

**Deliverables:**
- Complete research pipeline
- Experiment tracking system
- Result comparison tools
- TUI dashboard (using bubbletea or similar)
- SQLite database integration
- `docs/research-workflow.md`

**Exit Criteria:**
- ✅ Research pipeline runs end-to-end
- ✅ Experiments tracked reliably
- ✅ Comparison tools useful
- ✅ Dashboard functional
- ✅ All tests pass

---

#### Sub-phase 4.3: Production Readiness
**Duration: 4-5 days**

**Objective:** Prepare for paper trading and eventual live trading.

**Tasks:**
1. **Live Trading Simulator**
   - `internal/live/simulator.go` — simulate live trading
   - Real-time bar processing (not historical)
   - Handle strategy state across restarts
   - Order management (place, cancel, modify)

2. **Risk Management**
   - `internal/risk/manager.go` — risk management layer
   - Position size limits
   - Max drawdown stops
   - Exposure limits (per symbol, total)
   - Kill switch (emergency stop all)

3. **Monitoring & Alerts**
   - `internal/monitoring/monitor.go` — monitor live performance
   - Track P&L in real-time
   - Alert on unusual events (large loss, etc.)
   - Health checks

4. **Deployment Checklist**
   - `docs/deployment-checklist.md` — pre-deployment checklist
   - Pre-flight checks (strategy valid, parameters sane, etc.)
   - Connectivity checks (exchange API reachable)
   - Balance checks (sufficient funds)

5. **Paper Trading Integration** (if exchange supports it)
   - Integrate with exchange testnet
   - Or simulate paper trading locally

**Deliverables:**
- Live trading simulator
- Risk management layer
- Monitoring and alerting
- Deployment checklist
- Paper trading integration (if possible)
- `docs/going-live.md` — guide for going live

**Exit Criteria:**
- ✅ Live simulator works like real trading
- ✅ Risk management prevents dangerous actions
- ✅ Monitoring catches issues
- ✅ Deployment checklist comprehensive
- ✅ Paper trading ready
- ✅ All tests pass

---

**Phase 4 Complete Exit Criteria:**
- ✅ Advanced optimization techniques working
- ✅ Research workflow smooth and efficient
- ✅ Production readiness achieved
- ✅ Paper trading ready
- ✅ Risk management solid
- ✅ All CircleCI tests green
- ✅ System ready for real-world use

---

## AI Agent Guidelines

### DO:
- ✅ Always verify backtest results before analysis
- ✅ Use structured schemas for input/output (no free-form)
- ✅ Generate testable hypotheses with clear logic
- ✅ Validate parameters against bounds and constraints
- ✅ Document reasoning and assumptions explicitly
- ✅ Use objective metrics for evaluation (not gut feeling)
- ✅ Admit when data is insufficient or inconclusive
- ✅ Flag potential issues proactively

### DON'T:
- ❌ Hallucinate strategy parameters (always within bounds)
- ❌ Make claims without backtest evidence
- ❌ Ignore validation errors (investigate root cause)
- ❌ Optimize without overfitting checks (use walk-forward)
- ❌ Skip sanity tests (always run them)
- ❌ Trust single backtest results (need out-of-sample validation)
- ❌ Bypass parameter bounds (respect constraints)
- ❌ Generate unsafe strategies (check risk limits)

### Research Protocol:

```
1. ANALYZE: Review backtest results (metrics, trades, equity curve)
2. HYPOTHESIZE: Form testable hypothesis (what, why, how)
3. VALIDATE: Check hypothesis against constraints (bounds, sanity)
4. IMPLEMENT: Generate strategy code (using templates)
5. TEST: Run backtest with validation (in-sample first)
6. VERIFY: Out-of-sample testing (walk-forward)
7. EVALUATE: Compare objective metrics (vs benchmark, vs previous)
8. ITERATE: Refine or pivot based on results
```

**After each iteration:**
- Document what was learned
- Store experiment in research database
- Update research dashboard
- If successful: add to strategy library
- If failed: document why for future reference

---

## Objective Validation Tools

### Data Quality Metrics:
- **Completeness:** % of expected data points present
- **Consistency:** outlier rate, suspicious patterns
- **Timeliness:** data freshness, staleness score
- **Accuracy:** cross-exchange validation, sanity checks

### Backtest Validation:
- **Reproducibility:** same input → same output (deterministic)
- **Benchmark comparison:** vs buy-and-hold (alpha calculation)
- **Statistical significance:** t-test, confidence intervals
- **Walk-forward validation:** in-sample vs out-of-sample

### Strategy Validation:
- **Parameter bounds:** all parameters within specified ranges
- **Logic correctness:** no look-ahead bias, no data leakage
- **Edge case handling:** empty data, single bar, extreme values
- **Overfitting detection:** performance gap analysis

### Integration Testing:
- **End-to-end tests:** full workflow from data fetch to report
- **Performance benchmarks:** execution time, memory usage
- **Error handling:** graceful degradation, recovery
- **Regression tests:** prevent breaking changes

---

## Current Phase: Phase 0 (Foundation)

**Status:** Not started

**Next Steps:**
1. Begin Sub-phase 0.1: Documentation & Methodology
   - Create `docs/methodology.md`
   - Create `docs/architecture.md`
   - Create `docs/coding-standards.md`
   - Setup CircleCI basic pipeline

2. After 0.1 complete → Sub-phase 0.2: Exchange API Research
3. After 0.2 complete → Sub-phase 0.3: Data Quality Framework
4. After all Phase 0 complete → Review and approve before Phase 1

**CRITICAL RULE:**
- **DO NOT proceed to Phase 1 until ALL Phase 0 exit criteria met.**
- **DO NOT skip sub-phases.**
- **DO NOT rush through documentation.**

Foundation is critical. Rushing will cause problems later.

---

## CircleCI Configuration

### Pipeline Structure:

```yaml
# .circleci/config.yml
version: 2.1

jobs:
  test:
    docker:
      - image: cimg/go:1.21
    steps:
      - checkout
      - run: go mod download
      - run: go test -v -race -coverprofile=coverage.out ./...
      - run: go tool cover -html=coverage.out -o coverage.html
      
  lint:
    docker:
      - image: golangci/golangci-lint:latest
    steps:
      - checkout
      - run: golangci-lint run
      
  benchmark:
    docker:
      - image: cimg/go:1.21
    steps:
      - checkout
      - run: go test -bench=. -benchmem ./...

workflows:
  build-test:
    jobs:
      - test
      - lint
      - benchmark
```

### CircleCI Usage:
- Every PR triggers full test suite
- Tests must pass before merge
- Coverage reports generated automatically
- Benchmarks tracked over time
- Linting enforces code quality

---

**This document is the single source of truth for backtest-go development.**