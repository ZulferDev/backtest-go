# Research Methodology — backtest-go

**Version:** 1.0  
**Last Updated:** 2026-08-29  
**Status:** Active

---

## Overview

backtest-go bukan sekadar backtesting framework — ini adalah **research infrastructure** untuk systematic strategy discovery. Dokumen ini mendefinisikan metodologi riset yang akan digunakan untuk menemukan, meneliti, dan memvalidasi trading strategies.

---

## Core Philosophy: Hypothesis-Driven Testing

### Traditional Approach (❌ AVOID)
```
1. Pilih indikator random ("SMA kayaknya bagus")
2. Tweak parameter sampai backtest "looks good"
3. Deploy ke live trading
4. Rugi karena overfitting
```

### Our Approach (✅ FOLLOW)
```
1. OBSERVE: Identifikasi market behavior / pattern
2. HYPOTHESIZE: Form testable hypothesis ("If X, then Y")
3. DESIGN: Design strategy yang test hypothesis
4. BACKTEST: Run pada historical data
5. ANALYZE: Objective metrics, bukan "looks good"
6. VALIDATE: Out-of-sample testing (walk-forward)
7. ITERATE: Refine atau pivot based on evidence
8. DEPLOY: Only after passing all validation gates
```

---

## Research Cycle

### Phase 1: Observation & Hypothesis Formation

**Objective:** Identify potential edge/alpha source

**Activities:**
- Market analysis (price action, volume patterns)
- Literature review (papers, existing strategies)
- Statistical analysis (correlation, seasonality)

**Output:** Written hypothesis document

**Example Hypothesis:**
```
Hypothesis: Bitcoin exhibits mean-reversion behavior on 1-hour timeframe
when RSI < 30 (oversold condition).

Rationale: Crypto markets overreact to short-term news, creating
temporary price dislocations that correct within 4-12 hours.

Testable Prediction: 
- Entry when RSI < 30
- Exit when RSI > 50 or after 12 bars
- Expected win rate > 55%
- Expected profit factor > 1.5
```

**Validation Criteria:**
- ✅ Hypothesis is specific and testable
- ✅ Clear entry/exit rules defined
- ✅ Expected metrics stated upfront
- ✅ Rationale grounded in market behavior (not random)

---

### Phase 2: Strategy Design

**Objective:** Translate hypothesis into executable strategy code

**Guidelines:**

1. **Clarity Over Cleverness**
   - Code harus readable, bukan "clever"
   - Prefer explicit logic over implicit
   - Comments explain "why", bukan "what"

2. **No Look-Ahead Bias**
   - Strategy hanya pakai data yang tersedia at bar close
   - No peeking ke future bars
   - Indikator tidak boleh use future data

3. **Parameter Constraints**
   - Define min/max bounds untuk setiap parameter
   - Document parameter rationale
   - Avoid too many parameters (overfitting risk)

4. **Risk Management**
   - Position sizing rules
   - Stop loss / take profit logic
   - Max drawdown limits

**Output:** Strategy code + parameter schema

**Validation Criteria:**
- ✅ Code compiles dan runs
- ✅ No look-ahead bias detected
- ✅ Parameters within reasonable bounds
- ✅ Risk management implemented
- ✅ Unit tests pass (>80% coverage)

---

### Phase 3: Backtesting

**Objective:** Test strategy on historical data

**Data Requirements:**
- Minimum 1 year historical data
- Include different market regimes:
  - Bull market (uptrend)
  - Bear market (downtrend)
  - Sideways / ranging market
  - High volatility periods
  - Low volatility periods

**Execution:**
```bash
# Run backtest dengan config
go run cmd/backtest/main.go \
  --config configs/btc_1h_rsi_mean_reversion.yaml \
  --output results/rsi_mr_v1/
```

**Output:**
- Trade log (JSON/CSV)
- Equity curve
- Performance metrics
- HTML report

**Initial Validation Criteria:**
- ✅ Backtest completes without errors
- ✅ Results reproducible (same input → same output)
- ✅ No obvious bugs (e.g., equity going negative)
- ✅ Trade count > 30 (statistical significance)

---

### Phase 4: Objective Analysis

**Objective:** Evaluate strategy using objective metrics

**Primary Metrics:**

1. **Total Return**
   - Formula: `(Final Equity - Initial Capital) / Initial Capital`
   - Target: > 20% annually
   - Benchmark: Compare vs buy-and-hold

2. **Sharpe Ratio** (risk-adjusted return)
   - Formula: `(Mean Return - Risk-Free Rate) / Std Dev Returns`
   - Target: > 1.0 (good), > 2.0 (excellent)
   - Annualized: multiply by √252

3. **Sortino Ratio** (downside risk)
   - Formula: `(Mean Return - Risk-Free Rate) / Downside Deviation`
   - Target: > 1.5
   - Better than Sharpe (only penalize downside volatility)

4. **Maximum Drawdown** (worst loss)
   - Formula: `(Trough - Peak) / Peak`
   - Target: < 20%
   - Critical: measure pain tolerance

5. **Win Rate**
   - Formula: `Winning Trades / Total Trades`
   - Target: > 50% (for mean-reversion strategies)
   - Note: Not sole indicator (40% win rate can still be profitable)

6. **Profit Factor**
   - Formula: `Gross Profit / Gross Loss`
   - Target: > 1.5
   - Critical: measures robustness

7. **Average Trade Duration**
   - Measure: median holding period
   - Check: aligns with hypothesis timeframe

**Secondary Metrics:**
- Calmar Ratio (Return / Max Drawdown)
- Recovery Time (time to recover from drawdown)
- Consecutive Losses (longest losing streak)
- Trade Distribution (histogram of P&L)

**Red Flags:**
- 🚩 Single huge winner (luck, not edge)
- 🚩 Equity curve flat for long periods
- 🚩 Win rate < 30% (for mean-reversion)
- 🚩 Max drawdown > 30%
- 🚩 Profit factor < 1.2
- 🚩 Fewer than 30 trades (insufficient data)

**Decision Matrix:**

| Sharpe | Max DD | Profit Factor | Decision |
|--------|--------|---------------|----------|
| > 2.0  | < 15%  | > 2.0         | ✅ Excellent — proceed to validation |
| > 1.5  | < 20%  | > 1.5         | ✅ Good — proceed to validation |
| > 1.0  | < 25%  | > 1.3         | ⚠️ Marginal — refine or reconsider |
| < 1.0  | > 25%  | < 1.3         | ❌ Reject — pivot hypothesis |

---

### Phase 5: Out-of-Sample Validation

**Objective:** Test pada data yang tidak pernah dilihat strategy

**Why This Matters:**
- In-sample: data yang digunakan untuk develop/optimize strategy
- Out-of-sample: "future" data yang strategy belum pernah "lihat"
- Overfitting terdeteksi jika:
  - In-sample Sharpe = 2.5
  - Out-of-sample Sharpe = 0.8
  - Gap terlalu besar = strategy memorize noise

**Method 1: Simple Train/Test Split**
```
Data: 2022-01-01 to 2024-12-31 (3 years)

Train (in-sample): 2022-01-01 to 2023-12-31 (2 years)
Test (out-of-sample): 2024-01-01 to 2024-12-31 (1 year)
```

**Method 2: Walk-Forward Analysis** (preferred)
```
Window 1:
  Train: 2022-01 to 2022-06 (6 months)
  Test:  2022-07 to 2022-07 (1 month)

Window 2:
  Train: 2022-02 to 2022-07 (6 months)
  Test:  2022-08 to 2022-08 (1 month)

... roll forward monthly
```

**Validation Criteria:**

✅ **PASS** if:
- Out-of-sample Sharpe > 1.0
- Performance degradation < 30% vs in-sample
- Still profitable (positive returns)
- Max drawdown stays within acceptable range

❌ **FAIL** if:
- Out-of-sample Sharpe < 0.5
- Performance drops > 50%
- Strategy becomes unprofitable
- Max drawdown exceeds 40%

⚠️ **REVIEW** if:
- Marginal performance (Sharpe 0.5-1.0)
- Inconsistent across walk-forward windows
- High variance in returns

---

### Phase 6: Robustness Testing

**Objective:** Ensure strategy tidak fragile

**Test 1: Parameter Sensitivity**
- Tweak parameters ±10-20%
- Strategy should still work (not cliff edge)
- Example:
  - RSI threshold: 25, 30, 35
  - Stop loss: 1.5%, 2%, 2.5%
- Fragile strategy = small parameter change → drastically different result

**Test 2: Market Regime Analysis**
- Bull market performance
- Bear market performance
- Sideways market performance
- Strategy should work across regimes (or explicitly state regime dependency)

**Test 3: Different Symbols**
- Test on BTC, ETH, other major coins
- If strategy is "universal", should work across assets
- If symbol-specific, document why

**Test 4: Transaction Cost Sensitivity**
- Increase commission 2x
- Increase slippage 2x
- Strategy should still be profitable (just reduced)

**Validation Criteria:**
- ✅ Performance degrades gracefully (not cliff)
- ✅ Works across multiple market regimes
- ✅ Robust to transaction cost variations
- ✅ Not hyper-sensitive to single parameter

---

### Phase 7: Iteration & Refinement

**If Strategy Failed Validation:**

1. **Analyze Why:**
   - Look at losing trades
   - Identify common patterns
   - Check if hypothesis assumptions hold

2. **Decide:**
   - **Refine:** Minor adjustments (risk management, filters)
   - **Pivot:** Hypothesis fundamentally wrong, try new approach
   - **Abandon:** No edge exists, move to different strategy

3. **Document:**
   - What didn't work
   - Why it didn't work
   - What was learned
   - Store in research log (avoid repeating mistakes)

**If Strategy Passed Validation:**

1. **Optimize (Carefully):**
   - Fine-tune parameters using walk-forward
   - Avoid overfitting (use regularization)
   - Goal: improve robustness, not just metrics

2. **Document:**
   - Final parameters
   - Expected performance (conservative estimate)
   - Known limitations
   - Market regimes where it works best

3. **Prepare for Paper Trading:**
   - Create deployment config
   - Set up monitoring
   - Define failure criteria (when to stop)

---

### Phase 8: Paper Trading (Pre-Production)

**Objective:** Validate strategy in real-time (simulated)

**Duration:** Minimum 30 days (1 month)

**Activities:**
- Run strategy on live data stream
- Execute trades in simulation (no real money)
- Monitor performance daily
- Compare actual vs backtest expectation

**Metrics to Track:**
- Real-time P&L
- Execution slippage (actual vs expected)
- Latency (signal generation to order placement)
- Strategy errors / exceptions
- Downtime / connection issues

**Success Criteria:**
- ✅ No critical errors (crashes, infinite loops)
- ✅ Performance within ±30% of backtest expectation
- ✅ Execution slippage < 0.1%
- ✅ Strategy behaves as expected

**Failure Criteria:**
- ❌ Consistent losses (diverges from backtest)
- ❌ Excessive slippage (> 0.3%)
- ❌ Critical bugs / crashes
- ❌ Execution delays > 5 seconds

**Action on Failure:**
- Stop paper trading immediately
- Debug root cause
- Fix issues
- Re-run backtest to verify fix
- Restart paper trading (reset 30-day clock)

---

## Success Metrics Definition

### Phase-Level Success Criteria

**Phase 0: Foundation**
- ✅ All documentation complete and reviewed
- ✅ Exchange API researched and validated
- ✅ Data quality framework implemented
- ✅ CircleCI pipeline green

**Phase 1: Core Backtest Engine**
- ✅ Can backtest 1 year of 1h data in < 5 seconds
- ✅ Results reproducible (100% same output)
- ✅ All metrics validated vs hand calculation
- ✅ No data corruption or race conditions

**Phase 2: Strategy Framework**
- ✅ Strategy interface unambiguous
- ✅ Example strategies pass all tests
- ✅ Validation catches look-ahead bias
- ✅ Test coverage > 80%

**Phase 3: AI Integration**
- ✅ AI can generate valid hypotheses
- ✅ No hallucinations in 100 test runs
- ✅ Safety checks prevent dangerous strategies
- ✅ Schema validation catches errors

**Phase 4: Production Readiness**
- ✅ Walk-forward optimization working
- ✅ Overfitting detection reliable
- ✅ Paper trading runs 7 days error-free
- ✅ Risk management layer solid

### Strategy-Level Success Criteria

**Minimum Viable Strategy:**
- Total Return > 20% annually
- Sharpe Ratio > 1.0
- Max Drawdown < 25%
- Profit Factor > 1.3
- Win Rate > 40% (or 30% if profit factor > 2.0)
- Trade count > 30 (statistical significance)

**Production-Ready Strategy:**
- Total Return > 30% annually
- Sharpe Ratio > 1.5
- Max Drawdown < 20%
- Profit Factor > 1.5
- Win Rate > 50%
- Passes walk-forward validation
- Robust to parameter variations
- Works across market regimes
- Paper trading successful (30 days)

---

## Validation Criteria Per Phase

### Data Validation

**Completeness:**
- No missing bars (gaps in timestamp)
- All required fields present (OHLCV)
- Timeframe consistent (no irregular intervals)

**Consistency:**
- Open ≤ High, Open ≥ Low
- Close ≤ High, Close ≥ Low
- Volume ≥ 0
- No outliers (price spike > 50% in single bar)

**Accuracy:**
- Cross-validate with multiple exchanges
- Compare key levels (e.g., daily high/low)
- Verify against public charts (TradingView)

**Actions on Failure:**
- Flag invalid bars
- Re-fetch from exchange
- If persistent, exclude problematic data range
- Document data quality issues

### Code Validation

**Compilation:**
- Code compiles without errors
- No unused imports/variables
- Linter passes (golangci-lint)

**Correctness:**
- Unit tests pass (100%)
- Integration tests pass
- Manual verification of sample results

**Performance:**
- Benchmarks within acceptable range
- No memory leaks (profiling)
- CPU usage reasonable

**Actions on Failure:**
- Fix bugs immediately
- Add regression test to prevent recurrence
- Update documentation if behavior changes

### Backtest Validation

**Reproducibility:**
- Run 3 times → same results (deterministic)
- Same config → same output (no randomness)

**Correctness:**
- Manually verify sample trades
- Check edge cases (first/last bar)
- Validate metrics vs hand calculation

**Realism:**
- Commission applied correctly
- Slippage modeled
- No look-ahead bias

**Actions on Failure:**
- Debug root cause
- Add test case to catch similar issues
- Re-run full test suite

---

## Failure Recovery Protocol

### Type 1: Data Quality Issues

**Symptoms:**
- Missing bars
- Suspicious price spikes
- Volume anomalies

**Recovery Steps:**
1. Identify affected date range
2. Re-fetch data from exchange
3. If issue persists, try alternative data source
4. Document issue in data quality log
5. Exclude bad data if unfixable
6. Re-run affected backtests

**Prevention:**
- Automated data validation on ingestion
- Multiple data sources (redundancy)
- Regular data quality audits

---

### Type 2: Code Bugs

**Symptoms:**
- Unexpected results
- Crashes / panics
- Memory leaks
- Wrong calculations

**Recovery Steps:**
1. Isolate bug (minimal reproduction case)
2. Write failing test
3. Fix bug
4. Verify fix (test passes)
5. Add regression test
6. Re-run affected backtests
7. Review similar code for same bug pattern

**Prevention:**
- High test coverage (>80%)
- Code review
- Property-based testing
- Continuous integration (CircleCI)

---

### Type 3: Strategy Overfitting

**Symptoms:**
- Great in-sample, terrible out-of-sample
- Parameter cliff edges
- Works on 1 symbol, fails on others

**Recovery Steps:**
1. Acknowledge overfitting (don't rationalize)
2. Analyze why (too many parameters? too specific?)
3. Simplify strategy (reduce parameters)
4. Re-validate with walk-forward
5. If still fails, pivot to new hypothesis

**Prevention:**
- Walk-forward validation (mandatory)
- Parameter count < 5
- Test on multiple symbols
- Avoid optimization loops

---

### Type 4: Infrastructure Failures

**Symptoms:**
- Exchange API down
- CircleCI failures
- Disk full
- Network issues

**Recovery Steps:**
1. Identify root cause
2. Implement workaround if possible
3. Wait for service recovery (if external)
4. Retry failed operations
5. Document incident

**Prevention:**
- Retry logic with exponential backoff
- Fallback data sources
- Monitoring & alerts
- Regular backups

---

## AI Agent Integration

### AI Role in Research Cycle

AI **DOES:**
- Generate hypothesis based on data patterns
- Suggest parameter ranges
- Analyze backtest results
- Identify potential improvements
- Flag anomalies / red flags

**AI DOES NOT:**
- Make final decisions (human reviews)
- Bypass validation gates
- Generate code without validation
- Optimize parameters without constraints

### AI Input Format (Structured)

```json
{
  "backtest_result": {
    "strategy": "RSI Mean Reversion",
    "parameters": {
      "rsi_period": 14,
      "oversold_threshold": 30,
      "overbought_threshold": 70
    },
    "metrics": {
      "total_return": 0.45,
      "sharpe_ratio": 1.8,
      "max_drawdown": 0.15,
      "profit_factor": 2.1,
      "win_rate": 0.58,
      "trade_count": 87
    },
    "equity_curve": [...],
    "trade_log": [...]
  },
  "market_context": {
    "symbol": "BTCUSDT",
    "timeframe": "1h",
    "date_range": "2023-01-01 to 2024-01-01",
    "market_regime": "mixed"
  }
}
```

### AI Output Format (Structured)

```json
{
  "analysis": {
    "strengths": [
      "High Sharpe ratio indicates good risk-adjusted returns",
      "Win rate above 55% suggests consistent edge",
      "Max drawdown within acceptable range"
    ],
    "weaknesses": [
      "Limited sample size (87 trades) - need more data",
      "Performance not tested out-of-sample yet"
    ],
    "red_flags": []
  },
  "recommendations": [
    {
      "action": "validate",
      "description": "Run walk-forward analysis (6-month train, 1-month test)",
      "priority": "high"
    },
    {
      "action": "test_robustness",
      "description": "Test with RSI oversold thresholds: 25, 30, 35",
      "priority": "medium"
    }
  ],
  "next_hypothesis": null
}
```

### Validation Protocol for AI Outputs

1. **Schema Validation**
   - Check JSON structure
   - Verify required fields present
   - Type checking

2. **Bounds Checking**
   - Parameters within acceptable ranges
   - Metrics within realistic values
   - No NaN or Inf values

3. **Logic Validation**
   - Recommendations make sense
   - No contradictory statements
   - Red flags correctly identified

4. **Human Review**
   - Final approval by human
   - Sanity check recommendations
   - Override if necessary

---

## Documentation Standards

Every research iteration must produce:

1. **Hypothesis Document**
   - What we're testing
   - Why we think it will work
   - Expected metrics

2. **Strategy Code**
   - Well-commented
   - Unit tested
   - Parameter schema

3. **Backtest Report**
   - Input config
   - Output metrics
   - Trade log
   - Equity curve

4. **Analysis Document**
   - What worked
   - What didn't work
   - Why (root cause analysis)
   - Next steps

5. **Research Log**
   - Chronological record
   - All iterations (success + failure)
   - Lessons learned

**Location:** `research/` directory

**Format:** Markdown (for readability) + JSON (for structured data)

---

## Continuous Improvement

This methodology document is living — akan diupdate based on:
- Lessons learned dari research cycles
- New best practices discovered
- Feedback dari paper trading / live trading
- Industry developments

**Update Protocol:**
1. Propose change (document rationale)
2. Review by team (or self-review)
3. Update document
4. Bump version number
5. Communicate changes

---

## Summary

**Key Principles:**
1. Hypothesis-driven (bukan random trial)
2. Objective metrics (bukan "looks good")
3. Out-of-sample validation (cegah overfitting)
4. Robustness testing (not fragile)
5. Paper trading (bridge to live)
6. Document everything (reproducibility)

**Success = Systematic + Disciplined + Evidence-Based**

---

**End of Document**