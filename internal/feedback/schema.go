package feedback

import (
	"encoding/json"
	"time"
)

// StructuredFeedback is the main feedback schema for AI consumption
type StructuredFeedback struct {
	StrategyID       string                  `json:"strategy_id"`
	IterationID      string                  `json:"iteration_id"`
	Phase            string                  `json:"phase"` // CONCEIVE, WRITE, LINT, TEST, BACKTEST, ANALYZE
	Timestamp        time.Time               `json:"timestamp"`
	Status           string                  `json:"status"` // success, failure, warning
	BacktestResults  *BacktestFeedback       `json:"backtest_results,omitempty"`
	HypothesisCheck  *HypothesisValidation   `json:"hypothesis_validation,omitempty"`
	IdentifiedIssues []Issue                 `json:"identified_issues"`
	NextAction       *ActionRecommendation   `json:"next_action"`
	LearningContext  *LearningContext        `json:"learning_context,omitempty"`
	Metadata         map[string]interface{}  `json:"metadata,omitempty"`
}

// BacktestFeedback contains parsed backtest results
type BacktestFeedback struct {
	TotalReturn    float64            `json:"total_return"`
	SharpeRatio    float64            `json:"sharpe_ratio"`
	MaxDrawdown    float64            `json:"max_drawdown"`
	WinRate        float64            `json:"win_rate"`
	TotalTrades    int                `json:"total_trades"`
	ProfitFactor   float64            `json:"profit_factor"`
	SortinoRatio   float64            `json:"sortino_ratio"`
	AverageWin     float64            `json:"average_win"`
	AverageLoss    float64            `json:"average_loss"`
	RawMetrics     map[string]float64 `json:"raw_metrics"`
	PerformanceTag string             `json:"performance_tag"` // excellent, good, acceptable, poor, failing
}

// HypothesisValidation checks if hypothesis was confirmed
type HypothesisValidation struct {
	ThesisConfirmed bool     `json:"thesis_confirmed"`
	Evidence        []string `json:"evidence"`
	Contradictions  []string `json:"contradictions"`
	ConfidenceLevel string   `json:"confidence_level"` // high, medium, low
	SupportScore    int      `json:"support_score"`    // Evidence count - Contradiction count
}

// Issue represents a problem or concern identified
type Issue struct {
	Severity    string `json:"severity"`    // critical, high, medium, low
	Category    string `json:"category"`    // performance, risk, overfitting, logic_error, code_quality
	Description string `json:"description"`
	Evidence    string `json:"evidence"`
	Suggestion  string `json:"suggestion"`
}

// ActionRecommendation suggests next steps for AI
type ActionRecommendation struct {
	Action         string   `json:"action"`          // continue, refine, pivot, abort
	Focus          string   `json:"focus"`           // What to improve next
	Rationale      string   `json:"rationale"`       // Why this action
	SpecificTasks  []string `json:"specific_tasks"`  // Concrete tasks to execute
	ExpectedImpact string   `json:"expected_impact"` // What improvement to expect
}

// LearningContext provides historical context
type LearningContext struct {
	PreviousIterations   int                    `json:"previous_iterations"`
	BestPerformance      map[string]float64     `json:"best_performance"`
	WorstPerformance     map[string]float64     `json:"worst_performance"`
	ImprovementTrend     string                 `json:"improvement_trend"` // improving, plateauing, degrading
	LearnedPatterns      []string               `json:"learned_patterns"`
	FailedApproaches     []string               `json:"failed_approaches"`
	SuccessfulTechniques []string               `json:"successful_techniques"`
	CurrentVersion       string                 `json:"current_version"`
}

// NewStructuredFeedback creates a new feedback instance
func NewStructuredFeedback(strategyID, iterationID, phase string) *StructuredFeedback {
	return &StructuredFeedback{
		StrategyID:       strategyID,
		IterationID:      iterationID,
		Phase:            phase,
		Timestamp:        time.Now(),
		IdentifiedIssues: []Issue{},
		Metadata:         make(map[string]interface{}),
	}
}

// AddIssue adds an issue to the feedback
func (f *StructuredFeedback) AddIssue(severity, category, description, evidence, suggestion string) {
	f.IdentifiedIssues = append(f.IdentifiedIssues, Issue{
		Severity:    severity,
		Category:    category,
		Description: description,
		Evidence:    evidence,
		Suggestion:  suggestion,
	})
}

// SetBacktestResults sets backtest performance data
func (f *StructuredFeedback) SetBacktestResults(metrics map[string]float64) {
	f.BacktestResults = &BacktestFeedback{
		TotalReturn:  metrics["total_return"],
		SharpeRatio:  metrics["sharpe_ratio"],
		MaxDrawdown:  metrics["max_drawdown"],
		WinRate:      metrics["win_rate"],
		TotalTrades:  int(metrics["total_trades"]),
		ProfitFactor: metrics["profit_factor"],
		SortinoRatio: metrics["sortino_ratio"],
		AverageWin:   metrics["average_win"],
		AverageLoss:  metrics["average_loss"],
		RawMetrics:   metrics,
	}

	// Tag performance level
	f.BacktestResults.PerformanceTag = f.classifyPerformance()
}

// classifyPerformance tags performance level
func (f *StructuredFeedback) classifyPerformance() string {
	if f.BacktestResults == nil {
		return "unknown"
	}

	sharpe := f.BacktestResults.SharpeRatio
	returns := f.BacktestResults.TotalReturn
	drawdown := f.BacktestResults.MaxDrawdown

	// Excellent: Sharpe > 2, Returns > 50%, Drawdown < 15%
	if sharpe > 2.0 && returns > 50 && drawdown < 15 {
		return "excellent"
	}

	// Good: Sharpe > 1.5, Returns > 20%, Drawdown < 20%
	if sharpe > 1.5 && returns > 20 && drawdown < 20 {
		return "good"
	}

	// Acceptable: Sharpe > 1.0, Returns > 10%, Drawdown < 25%
	if sharpe > 1.0 && returns > 10 && drawdown < 25 {
		return "acceptable"
	}

	// Poor: Positive returns but weak metrics
	if returns > 0 {
		return "poor"
	}

	// Failing: Negative returns
	return "failing"
}

// SetHypothesisValidation sets hypothesis check results
func (f *StructuredFeedback) SetHypothesisValidation(confirmed bool, evidence, contradictions []string, confidence string) {
	f.HypothesisCheck = &HypothesisValidation{
		ThesisConfirmed: confirmed,
		Evidence:        evidence,
		Contradictions:  contradictions,
		ConfidenceLevel: confidence,
		SupportScore:    len(evidence) - len(contradictions),
	}
}

// SetNextAction sets recommended next action
func (f *StructuredFeedback) SetNextAction(action, focus, rationale string, tasks []string, expectedImpact string) {
	f.NextAction = &ActionRecommendation{
		Action:         action,
		Focus:          focus,
		Rationale:      rationale,
		SpecificTasks:  tasks,
		ExpectedImpact: expectedImpact,
	}
}

// SetLearningContext adds historical learning context
func (f *StructuredFeedback) SetLearningContext(ctx *LearningContext) {
	f.LearningContext = ctx
}

// ToJSON serializes feedback to JSON
func (f *StructuredFeedback) ToJSON() ([]byte, error) {
	return json.MarshalIndent(f, "", "  ")
}

// ToCompactJSON serializes feedback to compact JSON
func (f *StructuredFeedback) ToCompactJSON() ([]byte, error) {
	return json.Marshal(f)
}

// FromJSON deserializes feedback from JSON
func FromJSON(data []byte) (*StructuredFeedback, error) {
	var feedback StructuredFeedback
	err := json.Unmarshal(data, &feedback)
	return &feedback, err
}

// ToAIPromptFormat converts feedback to human-readable format for AI prompts
func (f *StructuredFeedback) ToAIPromptFormat() string {
	prompt := "# Strategy Analysis Feedback\n\n"
	prompt += "## Context\n"
	prompt += "- Strategy: " + f.StrategyID + "\n"
	prompt += "- Iteration: " + f.IterationID + "\n"
	prompt += "- Phase: " + f.Phase + "\n"
	prompt += "- Status: " + f.Status + "\n\n"

	if f.BacktestResults != nil {
		prompt += "## Backtest Performance\n"
		prompt += "- Overall: " + f.BacktestResults.PerformanceTag + "\n"
		prompt += "- Total Return: " + formatFloat(f.BacktestResults.TotalReturn) + "%\n"
		prompt += "- Sharpe Ratio: " + formatFloat(f.BacktestResults.SharpeRatio) + "\n"
		prompt += "- Max Drawdown: " + formatFloat(f.BacktestResults.MaxDrawdown) + "%\n"
		prompt += "- Win Rate: " + formatFloat(f.BacktestResults.WinRate) + "%\n"
		prompt += "- Total Trades: " + formatInt(f.BacktestResults.TotalTrades) + "\n"
		prompt += "- Profit Factor: " + formatFloat(f.BacktestResults.ProfitFactor) + "\n\n"
	}

	if f.HypothesisCheck != nil {
		prompt += "## Hypothesis Validation\n"
		if f.HypothesisCheck.ThesisConfirmed {
			prompt += "✓ Hypothesis CONFIRMED (confidence: " + f.HypothesisCheck.ConfidenceLevel + ")\n\n"
		} else {
			prompt += "✗ Hypothesis REJECTED (confidence: " + f.HypothesisCheck.ConfidenceLevel + ")\n\n"
		}

		if len(f.HypothesisCheck.Evidence) > 0 {
			prompt += "Evidence:\n"
			for _, e := range f.HypothesisCheck.Evidence {
				prompt += "- " + e + "\n"
			}
			prompt += "\n"
		}

		if len(f.HypothesisCheck.Contradictions) > 0 {
			prompt += "Contradictions:\n"
			for _, c := range f.HypothesisCheck.Contradictions {
				prompt += "- " + c + "\n"
			}
			prompt += "\n"
		}
	}

	if len(f.IdentifiedIssues) > 0 {
		prompt += "## Identified Issues\n"
		for _, issue := range f.IdentifiedIssues {
			prompt += "### " + issue.Severity + " - " + issue.Category + "\n"
			prompt += "Problem: " + issue.Description + "\n"
			if issue.Evidence != "" {
				prompt += "Evidence: " + issue.Evidence + "\n"
			}
			prompt += "Suggestion: " + issue.Suggestion + "\n\n"
		}
	}

	if f.NextAction != nil {
		prompt += "## Next Action\n"
		prompt += "- Action: " + f.NextAction.Action + "\n"
		prompt += "- Focus: " + f.NextAction.Focus + "\n"
		prompt += "- Rationale: " + f.NextAction.Rationale + "\n"
		prompt += "- Expected Impact: " + f.NextAction.ExpectedImpact + "\n\n"

		if len(f.NextAction.SpecificTasks) > 0 {
			prompt += "Tasks:\n"
			for i, task := range f.NextAction.SpecificTasks {
				prompt += formatInt(i+1) + ". " + task + "\n"
			}
			prompt += "\n"
		}
	}

	if f.LearningContext != nil {
		prompt += "## Learning Context\n"
		prompt += "- Previous Iterations: " + formatInt(f.LearningContext.PreviousIterations) + "\n"
		prompt += "- Improvement Trend: " + f.LearningContext.ImprovementTrend + "\n"
		prompt += "- Current Version: " + f.LearningContext.CurrentVersion + "\n\n"

		if len(f.LearningContext.LearnedPatterns) > 0 {
			prompt += "Learned Patterns:\n"
			for _, p := range f.LearningContext.LearnedPatterns {
				prompt += "- " + p + "\n"
			}
			prompt += "\n"
		}

		if len(f.LearningContext.FailedApproaches) > 0 {
			prompt += "Failed Approaches (avoid):\n"
			for _, fa := range f.LearningContext.FailedApproaches {
				prompt += "- " + fa + "\n"
			}
			prompt += "\n"
		}

		if len(f.LearningContext.SuccessfulTechniques) > 0 {
			prompt += "Successful Techniques (leverage):\n"
			for _, st := range f.LearningContext.SuccessfulTechniques {
				prompt += "- " + st + "\n"
			}
			prompt += "\n"
		}
	}

	return prompt
}


