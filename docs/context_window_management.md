# Context Window Management

## Overview

Context Window Management adalah sistem yang memastikan AI agent bekerja dengan konteks terfokus pada setiap fase lifecycle, mencegah halusinasi dan mempertahankan konsistensi.

## Problem Statement

AI dengan context window terbatas cenderung:
- Mencampur informasi dari fase berbeda
- Membuat asumsi tentang data yang tidak ada
- Kehilangan fokus saat konteks terlalu besar
- Menghasilkan output yang tidak konsisten

## Solution: Phase Isolation

Setiap fase lifecycle (CONCEIVE, WRITE, LINT, TEST, BACKTEST, ANALYZE) dijalankan dalam **isolated context** dengan:
- Input files yang eksplisit
- Output files yang terdefinisi
- Focused prompt yang spesifik untuk fase

## Architecture

### PhaseContext Structure

```go
type PhaseContext struct {
    StrategyID  string                 // Unique strategy identifier
    Phase       string                 // Current lifecycle phase
    Version     int                    // Strategy version number
    InputFiles  []string               // Files to read for this phase
    OutputFiles []string               // Files to write in this phase
    Metadata    map[string]interface{} // Phase-specific metadata
}
```

### ContextManager

```go
type ContextManager struct {
    baseDir string // Base directory for research_logs
}
```

## Phase-Specific Contexts

### CONCEIVE Phase
**Input:** None (fresh start)  
**Output:** `hypothesis.md`  
**Focus:** Market observation → Core thesis → Success criteria

### WRITE Phase
**Input:** `hypothesis.md`  
**Output:** `strategy_v{N}.go`  
**Focus:** Translate hypothesis to Go code using SDK

### LINT Phase
**Input:** `strategy_v{N}.go`  
**Output:** `validation_errors.log`  
**Focus:** AST validation, unsafe code detection

### TEST Phase
**Input:** `strategy_v{N}.go`  
**Output:** `validation_errors.log`  
**Focus:** Compilation check, syntax validation

### BACKTEST Phase
**Input:** `strategy_v{N}.go`  
**Output:** `results_v{N}.json`  
**Focus:** Execute historical simulation

### ANALYZE Phase
**Input:** `hypothesis.md`, `results_v{N}.json`, `memory.json`  
**Output:** `analysis_v{N}.json`, `memory.json` (updated)  
**Focus:** Hypothesis validation, weakness identification, learning update

## Usage Example

```go
import "backtest-go/internal/context"

// Initialize context manager
cm := context.NewContextManager("research_logs")

// Get phase context
ctx, err := cm.GetPhaseContext("rsi_mean_reversion", "WRITE", 1)
if err != nil {
    log.Fatal(err)
}

// Read phase-specific inputs
inputs, err := cm.ReadPhaseInputs(ctx)
if err != nil {
    log.Fatal(err)
}

// Generate focused prompt for AI
prompt, err := cm.GenerateFocusedPrompt(ctx, inputs)
if err != nil {
    log.Fatal(err)
}

// Send prompt to AI...
// AI generates output...

// Validate outputs were created
if err := cm.ValidatePhaseOutputs(ctx); err != nil {
    log.Fatal("AI failed to create required outputs:", err)
}

// Save context for debugging
cm.SavePhaseContext(ctx)
```

## Focused Prompt Generation

Each phase has a tailored prompt template:

### CONCEIVE Prompt
```
You are a quantitative researcher. Formulate a trading hypothesis.

Context:
- Market: Cryptocurrency (BTCUSDT)
- Timeframe: 1 hour
- Available indicators: SMA, EMA, RSI, MACD, ATR, Bollinger Bands

Task: Write hypothesis.md following structured format.
```

### WRITE Prompt
```
You are a Go developer implementing a trading strategy.

Hypothesis to implement:
---
{hypothesis content}
---

Constraints:
- No imports beyond SDK
- No goroutines
- Deterministic logic only

Task: Write strategy.go implementing the hypothesis.
```

### ANALYZE Prompt
```
You are a quantitative analyst evaluating backtest results.

Hypothesis:
---
{hypothesis}
---

Backtest Results (v{N}):
---
{results JSON}
---

Previous Learning:
---
{memory JSON}
---

Task:
1. Validate if hypothesis confirmed/rejected
2. Identify specific weaknesses
3. Propose ONE focused improvement
4. Update memory.json
```

## Anti-Hallucination Features

### 1. Explicit File Dependencies
AI cannot "remember" data from previous phases - it MUST read from files.

### 2. Output Validation
Every phase validates that required outputs were created before proceeding.

### 3. Single Responsibility
Each phase focuses on ONE task only, reducing cognitive load.

### 4. Version Tracking
Every iteration has explicit version numbers, preventing confusion.

### 5. Context Persistence
Phase contexts are saved to disk for debugging and audit trails.

## Integration with Orchestrator

```go
// Orchestrator uses ContextManager for each phase
cm := context.NewContextManager("research_logs")

for _, phase := range []string{"CONCEIVE", "WRITE", "LINT", "TEST", "BACKTEST", "ANALYZE"} {
    ctx, _ := cm.GetPhaseContext(strategyID, phase, version)
    inputs, _ := cm.ReadPhaseInputs(ctx)
    prompt, _ := cm.GenerateFocusedPrompt(ctx, inputs)
    
    // Execute AI with focused prompt
    aiOutput := executeAI(prompt)
    
    // Validate outputs created
    if err := cm.ValidatePhaseOutputs(ctx); err != nil {
        log.Fatal("Phase failed:", err)
    }
}
```

## Benefits

1. **Reduced Hallucination**: AI cannot invent data - must read from files
2. **Clear Boundaries**: Each phase has explicit inputs/outputs
3. **Focused Prompts**: Smaller, targeted prompts improve AI accuracy
4. **Reproducibility**: Context saved to disk for debugging
5. **Audit Trail**: Complete history of what AI read/wrote in each phase
6. **Error Isolation**: Failures localized to specific phase

## Testing

Comprehensive test coverage includes:
- Phase context creation for all phases
- Input file reading (with optional files)
- Output validation
- Focused prompt generation
- Context persistence (save/load)
- Cleanup functionality

Run tests:
```bash
go test ./internal/context/... -v
```

## Directory Structure

```
research_logs/
├── rsi_mean_reversion/
│   ├── hypothesis.md              # CONCEIVE output
│   ├── strategy_v1.go             # WRITE output
│   ├── validation_errors.log      # LINT/TEST output
│   ├── results_v1.json            # BACKTEST output
│   ├── analysis_v1.json           # ANALYZE output
│   ├── memory.json                # ANALYZE output (updated)
│   ├── context_CONCEIVE_v1.json   # Phase context (debug)
│   ├── context_WRITE_v1.json
│   └── ...
```

## Best Practices

1. **Always use ContextManager** when orchestrating AI workflows
2. **Never pass data between phases** except through files
3. **Validate outputs** before proceeding to next phase
4. **Save contexts** for production debugging
5. **Clean up contexts** after successful completion (optional)
6. **Version everything** - never overwrite files

## Future Enhancements

- Token counting for context size limits
- Automatic context compression when approaching limits
- Multi-model support (different models per phase)
- Parallel phase execution where independent
