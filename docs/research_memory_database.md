# Research Memory Database - SQLite Implementation

## Overview

Persistent SQLite database for storing AI research learning across iterations. Replaces ephemeral JSON memory with queryable persistent storage.

## Schema

### Tables

#### strategies
- `id`: Auto-increment primary key
- `strategy_id`: Unique strategy identifier
- `name`: Strategy name
- `created_at`, `updated_at`: Timestamps
- `market_condition`: Market type (ranging, trending, etc.)
- `status`: Strategy status (active, archived, etc.)

#### hypotheses
- `id`: Auto-increment primary key
- `strategy_id`: Foreign key to strategies
- `hypothesis_id`: Unique hypothesis identifier
- `description`: Hypothesis text
- `created_at`: Timestamp
- `status`: active, confirmed, rejected, modified
- `related_code`: Code reference
- `evaluation_json`: Serialized evaluation results
- `lessons_json`: Serialized lessons learned

#### iterations
- `id`: Auto-increment primary key
- `strategy_id`: Foreign key to strategies
- `iteration_id`: Unique iteration identifier
- `code_version`: Version identifier (v1, v2, etc.)
- `created_at`: Timestamp
- `metrics_json`: Serialized backtest metrics
- `changes_json`: List of changes made
- `rationale`: Explanation of changes
- `improvement_pct`: Performance improvement percentage

#### insights
- `id`: Auto-increment primary key
- `strategy_id`: Foreign key to strategies
- `insight`: Insight text
- `created_at`: Timestamp
- `category`: Insight category
- `confidence`: 0.0-1.0 confidence score

#### patterns
- `id`: Auto-increment primary key
- `strategy_id`: Foreign key to strategies
- `pattern`: Pattern description
- `context`: Context where observed
- `frequency`: Observation count
- `first_observed`, `last_observed`: Timestamps
- `confidence`: 0.0-1.0 confidence score
- `actionable`: Boolean flag

#### feedback_logs
- `id`: Auto-increment primary key
- `strategy_id`: Foreign key to strategies
- `iteration_id`: Optional iteration reference
- `phase`: Lifecycle phase (CONCEIVE, WRITE, LINT, TEST, BACKTEST, ANALYZE)
- `feedback_json`: Serialized structured feedback
- `created_at`: Timestamp

## Usage

```go
// Initialize database
db, err := db.NewResearchDB("research_logs/research.db")
if err != nil {
    log.Fatal(err)
}
defer db.Close()

// Create strategy
db.CreateStrategy("rsi_mean_reversion", "RSI Mean Reversion", "ranging")

// Add hypothesis
db.AddHypothesis("rsi_mean_reversion", "hyp_001", "RSI oversold signals work in ranging markets", "strategy_v1.go")

// Record iteration
metrics := map[string]interface{}{
    "sharpe_ratio": 1.8,
    "total_return": 32.5,
    "max_drawdown": 14.2,
}
changes := []string{"Added ADX filter", "Adjusted position sizing"}
db.AddIteration("rsi_mean_reversion", "iter_002", "v2", metrics, changes, "Risk optimization", 15.5)

// Observe pattern
db.ObservePattern("rsi_mean_reversion", "Consecutive losses during strong trends", "Uptrend context", 0.8, true)

// Get summary
summary, _ := db.GetResearchSummary("rsi_mean_reversion")
fmt.Printf("Total hypotheses: %d, Confirmed: %d\n", summary.TotalHypotheses, summary.Confirmed)
```

## Benefits

1. **Persistence**: Data survives process restarts
2. **Queryability**: SQL queries for complex analysis
3. **Scalability**: Handles thousands of iterations efficiently
4. **Traceability**: Complete audit trail of research evolution
5. **Multi-strategy**: Single database tracks multiple strategies

## Integration Points

- Orchestrator CLI: Saves feedback after each phase
- Analyzer: Queries historical data for learning context
- AI Prompts: Retrieves relevant patterns and lessons
