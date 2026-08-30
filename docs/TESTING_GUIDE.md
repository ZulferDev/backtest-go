# Testing Guide - backtest-go

## Overview

This guide documents the testing strategy and how to run tests for the backtest-go framework.

## Test Structure

```
backtest-go/
├── internal/
│   ├── analyzer/          # Walk-forward analysis tests
│   ├── backtest/          # Engine & fees integration tests
│   ├── codegen/           # Code generation pipeline tests
│   ├── execution/         # Fee & slippage model tests
│   ├── indicators/        # Technical indicator tests (including cache)
│   ├── paper/             # Paper trading executor tests
│   ├── report/            # HTML report generator tests
│   ├── validator/         # AST code validation tests
│   └── ...
└── test/
    └── integration/       # End-to-end integration tests
```

## Running Tests

### Run All Tests
```bash
go test ./...
```

### Run Without Cache (for CI/reproducibility)
```bash
go test ./... -count=1
```

### Run Specific Package
```bash
go test ./internal/execution -v
go test ./test/integration -v
```

### Run Specific Test
```bash
go test ./internal/execution -run TestFixedPercentageFee -v
go test ./test/integration -run TestEndToEndPipeline -v
```

### Run with Coverage
```bash
go test ./... -cover
go test ./internal/execution -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Test Categories

### Unit Tests

**Location:** `internal/*/`

**Purpose:** Test individual components in isolation

**Examples:**
- `internal/execution/fees_slippage_test.go` - Fee/slippage calculations
- `internal/indicators/rolling_test.go` - Rolling window functions
- `internal/indicators/cache_test.go` - Indicator caching with concurrency
- `internal/validator/validator_test.go` - AST-based code validation

**Key Principles:**
- Fast execution (< 100ms per test)
- No external dependencies
- Deterministic results
- Use epsilon comparison for floating point (1e-6 tolerance)

### Integration Tests

**Location:** `test/integration/`

**Purpose:** Verify end-to-end workflows

**Test Files:**

1. **pipeline_test.go**
   - TestEndToEndPipeline: Code gen → Validate → Backtest → Analyze
   - TestCodeValidation: AST validation for safe/unsafe code
   - TestCompilePipeline: Go code compilation

2. **walkforward_test.go**
   - TestWalkForwardOrchestrator: Multi-window backtesting
   - Overfitting detection
   - In-sample vs out-of-sample analysis

3. **optimizer_test.go**
   - TestParallelOptimizer: Concurrent parameter optimization
   - 9 parameter combinations
   - Task completion verification

4. **paper_test.go**
   - TestPaperTradingExecution: Position management
   - TestPaperTradingContext: Strategy execution with bars
   - Fee calculation verification

**Characteristics:**
- Longer execution (200-500ms)
- Test multiple components together
- Verify real workflows
- Use temporary directories for file I/O

## Common Testing Patterns

### Floating Point Comparison

**Problem:** `50000 * 0.0006 = 29.999999...` not `30.0`

**Solution:**
```go
const epsilon = 1e-6

func almostEqual(a, b float64) bool {
    return math.Abs(a-b) < epsilon
}

// Usage
if !almostEqual(result, expected) {
    t.Errorf("Expected %.2f, got %.2f", expected, result)
}
```

### Testing Concurrent Code

```go
func TestCacheConcurrency(t *testing.T) {
    cache := NewCache()
    var wg sync.WaitGroup
    
    // Launch 10 goroutines
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            // Perform 100 operations
            for j := 0; j < 100; j++ {
                cache.Set(key, value)
                cache.Get(key)
            }
        }(i)
    }
    
    wg.Wait()
    // Verify no race conditions
}
```

### Integration Test Template

```go
func TestFeatureIntegration(t *testing.T) {
    // Setup
    tmpDir := t.TempDir()
    component := NewComponent(tmpDir)
    
    // Execute
    result, err := component.Process(input)
    
    // Verify
    if err != nil {
        t.Fatalf("Unexpected error: %v", err)
    }
    
    if !almostEqual(result.Metric, expectedValue) {
        t.Errorf("Expected %.2f, got %.2f", expectedValue, result.Metric)
    }
    
    // Cleanup automatic with t.TempDir()
}
```

## Test Data

### Mock Data Generation

```go
func generateTestData(bars int) []data.OHLCV {
    result := make([]data.OHLCV, bars)
    for i := 0; i < bars; i++ {
        result[i] = data.OHLCV{
            Timestamp: int64(1000 * (i + 1)),
            Open:      50000.0,
            High:      50100.0,
            Low:       49900.0,
            Close:     50000.0,
            Volume:    100.0,
        }
    }
    return result
}
```

### Realistic Test Scenarios

```go
// Fees & Slippage Test
func TestRealisticTrading(t *testing.T) {
    feeModel := BinanceSpotFeeModel()        // 0.1% taker
    slippageModel := DefaultSlippageModel()   // 5 bps
    
    simulator := NewExecutionSimulator(feeModel, slippageModel)
    
    fillPrice, totalCost, fee, err := simulator.SimulateExecution(
        50000.0, // price
        0.5,     // quantity
        "buy",   // side
    )
    
    // Verify realistic values
    if fillPrice < 50000 {
        t.Error("Buy fill price should be >= market price")
    }
    if fee <= 0 {
        t.Error("Fee should be positive")
    }
}
```

## Known Issues & Solutions

### Issue 1: Floating Point Precision
**Symptom:** Tests fail with "Expected 30.00, got 30.00"

**Fix:** Use `almostEqual()` with epsilon tolerance instead of `==`

### Issue 2: Test Cache Pollution
**Symptom:** Tests pass individually but fail when run together

**Fix:** Use `go test ./... -count=1` to disable caching

### Issue 3: Race Conditions
**Symptom:** Tests fail randomly with data races

**Fix:** Run with race detector `go test ./... -race`

## Continuous Integration

### CircleCI Configuration

Tests run automatically on every push to GitHub:
```yaml
- run: go test ./... -v -race -count=1
- run: go test ./... -coverprofile=coverage.out
- run: go tool cover -func=coverage.out
```

### Pre-Commit Checks

Recommended local checks before committing:
```bash
# Run all tests
go test ./...

# Check for race conditions
go test ./... -race

# Verify formatting
go fmt ./...

# Run linter
golangci-lint run
```

## Test Metrics

### Current Coverage (2026-08-30)
- **Total Tests:** 96
- **Pass Rate:** 100%
- **Packages:** 14
- **Total Lines:** ~3,700 (test code only)

### Package Details
```
✓ internal/analyzer      - 2+ tests
✓ internal/backtest      - 4 integration tests
✓ internal/cache         - 2+ tests
✓ internal/codegen       - 3+ tests
✓ internal/execution     - 18 tests (fees/slippage)
✓ internal/indicators    - 23 tests (rolling + cache)
✓ internal/metrics       - 2+ tests
✓ internal/optimizer     - 2+ tests
✓ internal/paper         - 3+ tests
✓ internal/report        - 3 tests
✓ internal/risk          - 4+ tests
✓ internal/signal        - 2+ tests
✓ internal/validator     - 5 tests
✓ test/integration       - 4 tests
```

## Best Practices

1. **Write Tests First** - TDD approach for new features
2. **Test Edge Cases** - Zero values, negative numbers, empty data
3. **Test Concurrency** - Use goroutines and sync primitives
4. **Use Epsilon Comparison** - For all floating point assertions
5. **Clean Test Data** - Use `t.TempDir()` for file operations
6. **Verify Real Results** - Don't just check for no errors
7. **Document Expectations** - Add comments explaining test logic
8. **Run Without Cache** - Use `-count=1` for reproducibility

## Debugging Failed Tests

### Step 1: Run with Verbose Output
```bash
go test ./internal/execution -v -run TestFailingTest
```

### Step 2: Add Debug Logging
```go
t.Logf("Debug: value = %.20f", value)
t.Logf("Expected: %.20f", expected)
```

### Step 3: Isolate the Problem
```bash
# Run single test
go test -run TestSpecific

# Run with race detector
go test -race -run TestSpecific

# Check for floating point issues
# Print values with high precision: %.20f
```

### Step 4: Check Test Dependencies
- Verify test data generation
- Check mock setup
- Ensure proper cleanup
- Look for shared state

## Contributing

When adding new features:
1. Write tests alongside implementation
2. Ensure all tests pass locally
3. Run `go test ./... -count=1 -race`
4. Add integration tests for workflows
5. Update this guide if adding new patterns

## Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Table-Driven Tests](https://go.dev/wiki/TableDrivenTests)
- [Go Test Race Detector](https://go.dev/doc/articles/race_detector)
- [COMPLETION_REPORT.md](../COMPLETION_REPORT.md) - Latest test results
