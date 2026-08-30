package main

// Memory represents the persistent learning state
type Memory struct {
	StrategyID       string            `json:"strategy_id"`
	StrategyLineage  []StrategyVersion `json:"strategy_lineage"`
	LearnedPatterns  []string          `json:"learned_patterns"`
	FailedApproaches []string          `json:"failed_approaches"`
	MarketInsights   []string          `json:"market_insights"`
}

// StrategyVersion represents one iteration
type StrategyVersion struct {
	Version         string           `json:"version"`
	Date            string           `json:"date"`
	Changes         string           `json:"changes"`
	BacktestResults BacktestSummary  `json:"backtest_results"`
	Lesson          string           `json:"lesson"`
}

// BacktestSummary contains key metrics
type BacktestSummary struct {
	SharpeRatio float64 `json:"sharpe_ratio"`
	MaxDrawdown float64 `json:"max_drawdown"`
	WinRate     float64 `json:"win_rate"`
	TotalReturn float64 `json:"total_return"`
}

// Analysis represents structured analysis output
type Analysis struct {
	StrategyID     string           `json:"strategy_id"`
	Version        string           `json:"version"`
	BacktestResults AnalysisResults `json:"backtest_results"`
	HypothesisValidation HypothesisValidation `json:"hypothesis_validation"`
	IdentifiedWeaknesses []string    `json:"identified_weaknesses"`
	NextIterationPlan    IterationPlan `json:"next_iteration_plan"`
}

// AnalysisResults contains detailed backtest metrics
type AnalysisResults struct {
	TotalReturn   float64 `json:"total_return"`
	SharpeRatio   float64 `json:"sharpe_ratio"`
	SortinoRatio  float64 `json:"sortino_ratio"`
	MaxDrawdown   float64 `json:"max_drawdown"`
	WinRate       float64 `json:"win_rate"`
	TotalTrades   int     `json:"total_trades"`
	WinningTrades int     `json:"winning_trades"`
	LosingTrades  int     `json:"losing_trades"`
	AvgWin        float64 `json:"avg_win"`
	AvgLoss       float64 `json:"avg_loss"`
	ProfitFactor  float64 `json:"profit_factor"`
}

// HypothesisValidation checks if hypothesis was confirmed
type HypothesisValidation struct {
	ThesisConfirmed     bool                       `json:"thesis_confirmed"`
	SuccessCriteriaMet  map[string]CriteriaCheck   `json:"success_criteria_met"`
	Evidence            []string                   `json:"evidence"`
}

// CriteriaCheck represents a single success criterion
type CriteriaCheck struct {
	Target float64 `json:"target"`
	Actual float64 `json:"actual"`
	Met    bool    `json:"met"`
}

// IterationPlan defines next refinement
type IterationPlan struct {
	Focus               string   `json:"focus"`
	Rationale           string   `json:"rationale"`
	ExpectedImprovement string   `json:"expected_improvement"`
	ImplementationSteps []string `json:"implementation_steps"`
}
