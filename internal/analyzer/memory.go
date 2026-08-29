package analyzer

import (
	"encoding/json"
	"fmt"
	"time"
)

// ResearchMemory stores AI research insights across iterations
type ResearchMemory struct {
	Hypotheses      []HypothesisRecord   `json:"hypotheses"`
	Iterations      []IterationRecord    `json:"iterations"`
	Insights        []string             `json:"insights"`
	Patterns        []PatternObservation `json:"patterns"`
	CreatedAt       time.Time            `json:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at"`
	StrategyName    string               `json:"strategy_name"`
	MarketCondition string               `json:"market_condition"`
}

// HypothesisRecord tracks a single hypothesis and its evaluation
type HypothesisRecord struct {
	ID              string           `json:"id"`
	Description     string           `json:"description"`
	Timestamp       time.Time        `json:"timestamp"`
	Evaluation      EvaluationResult `json:"evaluation"`
	Status          string           `json:"status"` // "active", "rejected", "confirmed", "modified"
	RelatedCode     string           `json:"related_code"`
	LessonsLearned  []string         `json:"lessons_learned"`
}

// IterationRecord tracks each backtest iteration
type IterationRecord struct {
	IterationID   string                 `json:"iteration_id"`
	Timestamp     time.Time              `json:"timestamp"`
	CodeVersion   string                 `json:"code_version"`
	Metrics       SummaryMetrics         `json:"metrics"`
	ChangesMade   []string               `json:"changes_made"`
	Rationale     string                 `json:"rationale"`
	Improvement   float64                `json:"improvement"` // % change from previous
}

// PatternObservation captures recurring patterns discovered during research
type PatternObservation struct {
	Pattern       string    `json:"pattern"`
	Context       string    `json:"context"`
	Frequency     int       `json:"frequency"`
	FirstObserved time.Time `json:"first_observed"`
	LastObserved  time.Time `json:"last_observed"`
	Confidence    float64   `json:"confidence"` // 0-1 scale
	Actionable    bool      `json:"actionable"`
}

// NewResearchMemory creates a new research memory store
func NewResearchMemory(strategyName string) *ResearchMemory {
	now := time.Now()
	return &ResearchMemory{
		Hypotheses:   []HypothesisRecord{},
		Iterations:   []IterationRecord{},
		Insights:     []string{},
		Patterns:     []PatternObservation{},
		CreatedAt:    now,
		UpdatedAt:    now,
		StrategyName: strategyName,
	}
}

// AddHypothesis adds a new hypothesis to track
func (m *ResearchMemory) AddHypothesis(description string, codeRef string) string {
	id := generateID()
	record := HypothesisRecord{
		ID:          id,
		Description: description,
		Timestamp:   time.Now(),
		Status:      "active",
		RelatedCode: codeRef,
	}
	m.Hypotheses = append(m.Hypotheses, record)
	m.UpdatedAt = time.Now()
	return id
}

// UpdateHypothesisEvaluation updates a hypothesis with evaluation results
func (m *ResearchMemory) UpdateHypothesisEvaluation(hypothesisID string, eval EvaluationResult, lessons []string) error {
	for i, h := range m.Hypotheses {
		if h.ID == hypothesisID {
			m.Hypotheses[i].Evaluation = eval
			m.Hypotheses[i].LessonsLearned = lessons
			
			// Update status based on evaluation
			if eval.Supported && eval.ConfidenceLevel == "high" {
				m.Hypotheses[i].Status = "confirmed"
			} else if !eval.Supported && len(eval.Contradictions) > 2 {
				m.Hypotheses[i].Status = "rejected"
			}
			
			m.UpdatedAt = time.Now()
			return nil
		}
	}
	return nil
}

// AddIteration records a new backtest iteration
func (m *ResearchMemory) AddIteration(codeVer string, metrics SummaryMetrics, changes []string, rationale string) {
	iterationID := generateID()
	
	var improvement float64
	if len(m.Iterations) > 0 {
		prevMetrics := m.Iterations[len(m.Iterations)-1].Metrics
		if prevMetrics.TotalReturn != 0 {
			improvement = (metrics.TotalReturn - prevMetrics.TotalReturn) / prevMetrics.TotalReturn * 100
		}
	}
	
	record := IterationRecord{
		IterationID: iterationID,
		Timestamp:   time.Now(),
		CodeVersion: codeVer,
		Metrics:     metrics,
		ChangesMade: changes,
		Rationale:   rationale,
		Improvement: improvement,
	}
	
	m.Iterations = append(m.Iterations, record)
	m.UpdatedAt = time.Now()
}

// AddInsight adds a research insight
func (m *ResearchMemory) AddInsight(insight string) {
	m.Insights = append(m.Insights, insight)
	m.UpdatedAt = time.Now()
}

// ObservePattern records a pattern observation
func (m *ResearchMemory) ObservePattern(pattern, context string, confidence float64, actionable bool) {
	now := time.Now()
	
	// Check if pattern already exists
	for i, p := range m.Patterns {
		if p.Pattern == pattern {
			m.Patterns[i].Frequency++
			m.Patterns[i].LastObserved = now
			m.Patterns[i].Confidence = (p.Confidence + confidence) / 2
			m.UpdatedAt = now
			return
		}
	}
	
	// Add new pattern
	m.Patterns = append(m.Patterns, PatternObservation{
		Pattern:       pattern,
		Context:       context,
		Frequency:     1,
		FirstObserved: now,
		LastObserved:  now,
		Confidence:    confidence,
		Actionable:    actionable,
	})
	m.UpdatedAt = now
}

// GetSummary returns a summary of research progress
func (m *ResearchMemory) GetSummary() ResearchSummary {
	confirmed := 0
	rejected := 0
	active := 0
	
	for _, h := range m.Hypotheses {
		switch h.Status {
		case "confirmed":
			confirmed++
		case "rejected":
			rejected++
		default:
			active++
		}
	}
	
	var avgImprovement float64
	if len(m.Iterations) > 1 {
		totalImpr := 0.0
		for _, it := range m.Iterations[1:] {
			totalImpr += it.Improvement
		}
		avgImprovement = totalImpr / float64(len(m.Iterations)-1)
	}
	
	return ResearchSummary{
		TotalHypotheses:    len(m.Hypotheses),
		Confirmed:          confirmed,
		Rejected:           rejected,
		Active:             active,
		TotalIterations:    len(m.Iterations),
		AverageImprovement: avgImprovement,
		TotalInsights:      len(m.Insights),
		PatternsFound:      len(m.Patterns),
	}
}

// ResearchSummary provides overview of research state
type ResearchSummary struct {
	TotalHypotheses    int     `json:"total_hypotheses"`
	Confirmed          int     `json:"confirmed"`
	Rejected           int     `json:"rejected"`
	Active             int     `json:"active"`
	TotalIterations    int     `json:"total_iterations"`
	AverageImprovement float64 `json:"average_improvement"`
	TotalInsights      int     `json:"total_insights"`
	PatternsFound      int     `json:"patterns_found"`
}

// ToJSON serializes memory to JSON
func (m *ResearchMemory) ToJSON() ([]byte, error) {
	return json.MarshalIndent(m, "", "  ")
}

// FromJSON deserializes memory from JSON
func FromJSON(data []byte) (*ResearchMemory, error) {
	var mem ResearchMemory
	err := json.Unmarshal(data, &mem)
	return &mem, err
}

func generateID() string {
	return time.Now().Format("20060102150405")
}

// BuildAIFeedback constructs feedback message for AI researcher
func (m *ResearchMemory) BuildAIFeedback() string {
	summary := m.GetSummary()
	
	feedback := "## Research Progress Summary\n\n"
	feedback += fmt.Sprintf("Hypotheses Tested: %d (Confirmed: %d, Rejected: %d, Active: %d)\n", 
		summary.TotalHypotheses, summary.Confirmed, summary.Rejected, summary.Active)
	feedback += fmt.Sprintf("Iterations Completed: %d\n", summary.TotalIterations)
	feedback += fmt.Sprintf("Average Improvement: %.2f%%\n", summary.AverageImprovement)
	feedback += fmt.Sprintf("Key Insights: %d\n\n", summary.TotalInsights)
	
	if len(m.Insights) > 0 {
		feedback += "### Recent Insights:\n"
		start := len(m.Insights) - 3
		if start < 0 {
			start = 0
		}
		for _, insight := range m.Insights[start:] {
			feedback += fmt.Sprintf("- %s\n", insight)
		}
		feedback += "\n"
	}
	
	if len(m.Patterns) > 0 {
		feedback += "### Observed Patterns:\n"
		for _, p := range m.Patterns {
			if p.Frequency >= 2 && p.Confidence > 0.6 {
				feedback += fmt.Sprintf("- **%s** (confidence: %.0f%%, observed %d times)\n", 
					p.Pattern, p.Confidence*100, p.Frequency)
			}
		}
		feedback += "\n"
	}
	
	return feedback
}
