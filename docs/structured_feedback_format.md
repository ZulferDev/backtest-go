# Structured Feedback Format - JSON Schema for AI

## Overview

Standardized JSON schema for AI consumption of backtest results, hypothesis validation, and learning context. Enables consistent, machine-readable feedback across all phases of the strategy lifecycle.

## Schema Structure

### Root: StructuredFeedback

```json
{
  "strategy_id": "string",
  "iteration_id": "string",
  "phase": "CONCEIVE|WRITE|LINT|TEST|BACKTEST|ANALYZE",
  "timestamp": "RFC3339 timestamp",
  "status": "success|failure|warning",
  "backtest_results": { ... },
  "hypothesis_validation": { ... },
  "identified_issues": [ ... ],
  "next_action": { ... },
  "learning_context": { ... },
  "metadata": { ... }
}
```

### BacktestFeedback

```json
{
  "total_return": 32.5,
  "sharpe_ratio": 1.75,
  "max_drawdown": 14.2,
  "win_rate": 62.5,
  "total_trades": 85,
  "profit_factor": 2.1,
  "sortino_ratio": 2.0,
  "average_win": 180.0,
  "average_loss": 95.0,
  "raw_metrics": { ... },
  "performance_tag": "excellent|good|acceptable|poor|failing"
}
```

**Performance Classification:**
- **excellent**: Sharpe > 2.0, Returns > 50%, Drawdown < 15%
- **good**: Sharpe > 1.5, Returns > 20%, Drawdown < 20%
- **acceptable**: Sharpe > 1.0, Returns > 10%, Drawdown < 25%
- **poor**: Positive returns but weak metrics
- **failing**: Negative returns

### HypothesisValidation

```json
{
  "thesis_confirmed": true,
  "evidence": [
    "Strong Sharpe ratio of 1.75",
    "Good win rate of 62.5%"
  ],
  "contradictions": [
    "Drawdown slightly elevated at 14.2%"
  ],
  "confidence_level": "high|medium|low",
  "support_score": 1
}
```

**Support Score**: `len(evidence) - len(contradictions)`

### Issue

```json
{
  "severity": "critical|high|medium|low",
  "category": "performance|risk|overfitting|logic_error|code_quality",
  "description": "Drawdown slightly elevated",
  "evidence": "Max DD = 14.2%",
  "suggestion": "Add trailing stop to protect profits"
}
```

### ActionRecommendation

```json
{
  "action": "continue|refine|pivot|abort",
  "focus": "Risk optimization",
  "rationale": "Further reduce drawdown while maintaining returns",
  "specific_tasks": [
    "Implement ATR-based trailing stop",
    "Test on different timeframes"
  ],
  "expected_impact": "Target max DD < 12%, maintain Sharpe > 1.5"
}
```

**Action Types:**
- **continue**: Strategy performing well, proceed to next phase
- **refine**: Make incremental improvements to existing approach
- **pivot**: Fundamental change needed, revise core logic
- **abort**: Strategy fundamentally flawed, reject hypothesis

### LearningContext

```json
{
  "previous_iterations": 2,
  "best_performance": {
    "sharpe_ratio": 2.1,
    "total_return": 45.0
  },
  "worst_performance": {
    "sharpe_ratio": 0.8,
    "total_return": -5.0
  },
  "improvement_trend": "improving|plateauing|degrading",
  "learned_patterns": [
    "RSI works best in ranging markets (ADX < 25)",
    "Position sizing should scale with volatility (ATR)"
  ],
  "failed_approaches": [
    "Fixed stop loss (5%) - too tight, stopped out prematurely",
    "RSI period 14 - too slow, late entries"
  ],
  "successful_techniques": [
    "ATR-based position sizing",
    "ADX trend filter"
  ],
  "current_version": "v3"
}
```

## Usage Examples

### Example 1: Successful Backtest

```json
{
  "strategy_id": "rsi_mean_reversion",
  "iteration_id": "iter_002",
  "phase": "ANALYZE",
  "timestamp": "2026-08-30T23:45:00Z",
  "status": "success",
  "backtest_results": {
    "total_return": 32.5,
    "sharpe_ratio": 1.75,
    "max_drawdown": 14.2,
    "win_rate": 62.5,
    "total_trades": 85,
    "profit_factor": 2.1,
    "performance_tag": "good"
  },
  "hypothesis_validation": {
    "thesis_confirmed": true,
    "evidence": [
      "Strong Sharpe ratio indicates consistent risk-adjusted returns",
      "Win rate above 60% confirms edge in market"
    ],
    "contradictions": [],
    "confidence_level": "high",
    "support_score": 2
  },
  "identified_issues": [
    {
      "severity": "medium",
      "category": "risk",
      "description": "Drawdown slightly elevated",
      "evidence": "Max DD = 14.2%",
      "suggestion": "Add trailing stop to protect profits"
    }
  ],
  "next_action": {
    "action": "refine",
    "focus": "Risk optimization",
    "rationale": "Performance is good, focus on reducing drawdown",
    "specific_tasks": [
      "Implement ATR-based trailing stop",
      "Test on different timeframes"
    ],
    "expected_impact": "Target max DD < 12%"
  }
}
```

### Example 2: Failed Strategy

```json
{
  "strategy_id": "macd_crossover",
  "iteration_id": "iter_001",
  "phase": "ANALYZE",
  "timestamp": "2026-08-30T23:50:00Z",
  "status": "failure",
  "backtest_results": {
    "total_return": -8.5,
    "sharpe_ratio": -0.3,
    "max_drawdown": 28.0,
    "win_rate": 38.0,
    "total_trades": 120,
    "profit_factor": 0.7,
    "performance_tag": "failing"
  },
  "hypothesis_validation": {
    "thesis_confirmed": false,
    "evidence": [],
    "contradictions": [
      "Negative returns indicate no edge",
      "Win rate below random expectation",
      "Profit factor < 1.0 means losses exceed wins",
      "Excessive drawdown shows poor risk control"
    ],
    "confidence_level": "high",
    "support_score": -4
  },
  "identified_issues": [
    {
      "severity": "critical",
      "category": "performance",
      "description": "Strategy loses money consistently",
      "evidence": "Total return = -8.5%, Profit factor = 0.7",
      "suggestion": "Reject hypothesis and pivot to different approach"
    },
    {
      "severity": "critical",
      "category": "risk",
      "description": "Unacceptable drawdown",
      "evidence": "Max DD = 28.0%",
      "suggestion": "Fundamental risk management issues"
    }
  ],
  "next_action": {
    "action": "pivot",
    "focus": "Fundamental strategy redesign",
    "rationale": "Current approach is fundamentally flawed",
    "specific_tasks": [
      "Research alternative entry signals",
      "Consider mean-reversion instead of trend-following",
      "Add market condition filters"
    ],
    "expected_impact": "Complete strategy overhaul needed"
  },
  "learning_context": {
    "previous_iterations": 0,
    "improvement_trend": "degrading",
    "failed_approaches": [
      "MACD crossover without filters - too many false signals"
    ],
    "current_version": "v1"
  }
}
```

## Integration with Orchestrator

```go
import "backtest-go/internal/feedback"

// Create feedback
fb := feedback.NewStructuredFeedback("rsi_mean_reversion", "iter_002", "ANALYZE")
fb.Status = "success"

// Set backtest results
metrics := map[string]float64{
    "total_return": 32.5,
    "sharpe_ratio": 1.75,
    // ... other metrics
}
fb.SetBacktestResults(metrics)

// Set hypothesis validation
evidence := []string{"Strong Sharpe ratio", "Good win rate"}
contradictions := []string{}
fb.SetHypothesisValidation(true, evidence, contradictions, "high")

// Add issues
fb.AddIssue("medium", "risk", "Drawdown elevated", "Max DD = 14.2%", "Add trailing stop")

// Set next action
tasks := []string{"Implement trailing stop", "Test on different timeframes"}
fb.SetNextAction("refine", "Risk optimization", "Reduce drawdown", tasks, "Target DD < 12%")

// Generate AI prompt
promptText := fb.ToAIPromptFormat()
fmt.Println(promptText)

// Save to database
jsonData, _ := fb.ToJSON()
db.SaveFeedback("rsi_mean_reversion", "iter_002", "ANALYZE", fb)
```

## Benefits

1. **Consistency**: Every phase produces structured output
2. **Machine-readable**: AI can parse and understand without hallucination
3. **Traceable**: Complete history stored in database
4. **Actionable**: Clear next steps guide AI decisions
5. **Learning**: Context accumulates across iterations

## Anti-Hallucination Features

- Explicit performance classification (no subjective interpretation)
- Evidence-based hypothesis validation (support score calculation)
- Structured issue taxonomy (severity + category)
- Clear action vocabulary (continue/refine/pivot/abort)
- Historical context prevents forgetting past lessons
