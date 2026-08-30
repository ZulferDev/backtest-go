package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// ResearchDB handles persistent storage of AI research data
type ResearchDB struct {
	db *sql.DB
}

// NewResearchDB creates or opens a research database
func NewResearchDB(dbPath string) (*ResearchDB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	rdb := &ResearchDB{db: db}
	if err := rdb.initialize(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return rdb, nil
}

// Close closes the database connection
func (r *ResearchDB) Close() error {
	return r.db.Close()
}

// initialize creates database schema
func (r *ResearchDB) initialize() error {
	schema := `
	CREATE TABLE IF NOT EXISTS strategies (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		strategy_id TEXT UNIQUE NOT NULL,
		name TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		market_condition TEXT,
		status TEXT DEFAULT 'active'
	);

	CREATE TABLE IF NOT EXISTS hypotheses (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		strategy_id TEXT NOT NULL,
		hypothesis_id TEXT UNIQUE NOT NULL,
		description TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		status TEXT DEFAULT 'active',
		related_code TEXT,
		evaluation_json TEXT,
		lessons_json TEXT,
		FOREIGN KEY (strategy_id) REFERENCES strategies(strategy_id)
	);

	CREATE TABLE IF NOT EXISTS iterations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		strategy_id TEXT NOT NULL,
		iteration_id TEXT UNIQUE NOT NULL,
		code_version TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		metrics_json TEXT NOT NULL,
		changes_json TEXT,
		rationale TEXT,
		improvement_pct REAL DEFAULT 0.0,
		FOREIGN KEY (strategy_id) REFERENCES strategies(strategy_id)
	);

	CREATE TABLE IF NOT EXISTS insights (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		strategy_id TEXT NOT NULL,
		insight TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		category TEXT,
		confidence REAL DEFAULT 0.5,
		FOREIGN KEY (strategy_id) REFERENCES strategies(strategy_id)
	);

	CREATE TABLE IF NOT EXISTS patterns (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		strategy_id TEXT NOT NULL,
		pattern TEXT NOT NULL,
		context TEXT,
		frequency INTEGER DEFAULT 1,
		first_observed DATETIME DEFAULT CURRENT_TIMESTAMP,
		last_observed DATETIME DEFAULT CURRENT_TIMESTAMP,
		confidence REAL DEFAULT 0.5,
		actionable BOOLEAN DEFAULT 0,
		FOREIGN KEY (strategy_id) REFERENCES strategies(strategy_id),
		UNIQUE(strategy_id, pattern)
	);

	CREATE TABLE IF NOT EXISTS feedback_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		strategy_id TEXT NOT NULL,
		iteration_id TEXT,
		phase TEXT NOT NULL,
		feedback_json TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (strategy_id) REFERENCES strategies(strategy_id)
	);

	CREATE INDEX IF NOT EXISTS idx_strategies_id ON strategies(strategy_id);
	CREATE INDEX IF NOT EXISTS idx_hypotheses_strategy ON hypotheses(strategy_id);
	CREATE INDEX IF NOT EXISTS idx_iterations_strategy ON iterations(strategy_id);
	CREATE INDEX IF NOT EXISTS idx_insights_strategy ON insights(strategy_id);
	CREATE INDEX IF NOT EXISTS idx_patterns_strategy ON patterns(strategy_id);
	CREATE INDEX IF NOT EXISTS idx_feedback_strategy ON feedback_logs(strategy_id);
	`

	_, err := r.db.Exec(schema)
	return err
}

// CreateStrategy creates a new strategy record
func (r *ResearchDB) CreateStrategy(strategyID, name, marketCondition string) error {
	_, err := r.db.Exec(`
		INSERT INTO strategies (strategy_id, name, market_condition)
		VALUES (?, ?, ?)
	`, strategyID, name, marketCondition)
	return err
}

// UpdateStrategyStatus updates strategy status
func (r *ResearchDB) UpdateStrategyStatus(strategyID, status string) error {
	_, err := r.db.Exec(`
		UPDATE strategies 
		SET status = ?, updated_at = CURRENT_TIMESTAMP
		WHERE strategy_id = ?
	`, status, strategyID)
	return err
}

// AddHypothesis adds a new hypothesis
func (r *ResearchDB) AddHypothesis(strategyID, hypothesisID, description, relatedCode string) error {
	_, err := r.db.Exec(`
		INSERT INTO hypotheses (strategy_id, hypothesis_id, description, related_code)
		VALUES (?, ?, ?, ?)
	`, strategyID, hypothesisID, description, relatedCode)
	return err
}

// UpdateHypothesisEvaluation updates hypothesis evaluation
func (r *ResearchDB) UpdateHypothesisEvaluation(hypothesisID string, evaluation interface{}, lessons []string, status string) error {
	evalJSON, err := json.Marshal(evaluation)
	if err != nil {
		return err
	}

	lessonsJSON, err := json.Marshal(lessons)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(`
		UPDATE hypotheses
		SET evaluation_json = ?, lessons_json = ?, status = ?
		WHERE hypothesis_id = ?
	`, string(evalJSON), string(lessonsJSON), status, hypothesisID)
	return err
}

// AddIteration records a new backtest iteration
func (r *ResearchDB) AddIteration(strategyID, iterationID, codeVersion string, metrics interface{}, changes []string, rationale string, improvement float64) error {
	metricsJSON, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	changesJSON, err := json.Marshal(changes)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(`
		INSERT INTO iterations (strategy_id, iteration_id, code_version, metrics_json, changes_json, rationale, improvement_pct)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, strategyID, iterationID, codeVersion, string(metricsJSON), string(changesJSON), rationale, improvement)
	return err
}

// AddInsight adds a research insight
func (r *ResearchDB) AddInsight(strategyID, insight, category string, confidence float64) error {
	_, err := r.db.Exec(`
		INSERT INTO insights (strategy_id, insight, category, confidence)
		VALUES (?, ?, ?, ?)
	`, strategyID, insight, category, confidence)
	return err
}

// ObservePattern records or updates a pattern observation
func (r *ResearchDB) ObservePattern(strategyID, pattern, context string, confidence float64, actionable bool) error {
	// Try to update existing pattern
	result, err := r.db.Exec(`
		UPDATE patterns
		SET frequency = frequency + 1,
		    last_observed = CURRENT_TIMESTAMP,
		    confidence = (confidence + ?) / 2.0,
		    context = ?
		WHERE strategy_id = ? AND pattern = ?
	`, confidence, context, strategyID, pattern)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// If no existing pattern, insert new
	if rowsAffected == 0 {
		_, err = r.db.Exec(`
			INSERT INTO patterns (strategy_id, pattern, context, confidence, actionable)
			VALUES (?, ?, ?, ?, ?)
		`, strategyID, pattern, context, confidence, actionable)
	}

	return err
}

// SaveFeedback saves structured feedback log
func (r *ResearchDB) SaveFeedback(strategyID, iterationID, phase string, feedback interface{}) error {
	feedbackJSON, err := json.Marshal(feedback)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(`
		INSERT INTO feedback_logs (strategy_id, iteration_id, phase, feedback_json)
		VALUES (?, ?, ?, ?)
	`, strategyID, iterationID, phase, string(feedbackJSON))
	return err
}

// GetStrategyIterations retrieves all iterations for a strategy
func (r *ResearchDB) GetStrategyIterations(strategyID string, limit int) ([]IterationRecord, error) {
	query := `
		SELECT iteration_id, code_version, created_at, metrics_json, changes_json, rationale, improvement_pct
		FROM iterations
		WHERE strategy_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := r.db.Query(query, strategyID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var iterations []IterationRecord
	for rows.Next() {
		var iter IterationRecord
		var metricsJSON, changesJSON string
		var createdAt string

		err := rows.Scan(&iter.IterationID, &iter.CodeVersion, &createdAt, &metricsJSON, &changesJSON, &iter.Rationale, &iter.Improvement)
		if err != nil {
			return nil, err
		}

		iter.Timestamp, _ = time.Parse("2006-01-02 15:04:05", createdAt)

		// Parse JSON fields
		json.Unmarshal([]byte(metricsJSON), &iter.Metrics)
		json.Unmarshal([]byte(changesJSON), &iter.ChangesMade)

		iterations = append(iterations, iter)
	}

	return iterations, nil
}

// GetStrategyHypotheses retrieves all hypotheses for a strategy
func (r *ResearchDB) GetStrategyHypotheses(strategyID string) ([]HypothesisRecord, error) {
	query := `
		SELECT hypothesis_id, description, created_at, status, related_code, evaluation_json, lessons_json
		FROM hypotheses
		WHERE strategy_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, strategyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var hypotheses []HypothesisRecord
	for rows.Next() {
		var hyp HypothesisRecord
		var createdAt, evalJSON, lessonsJSON sql.NullString

		err := rows.Scan(&hyp.ID, &hyp.Description, &createdAt, &hyp.Status, &hyp.RelatedCode, &evalJSON, &lessonsJSON)
		if err != nil {
			return nil, err
		}

		if createdAt.Valid {
			hyp.Timestamp, _ = time.Parse("2006-01-02 15:04:05", createdAt.String)
		}

		if evalJSON.Valid {
			json.Unmarshal([]byte(evalJSON.String), &hyp.Evaluation)
		}

		if lessonsJSON.Valid {
			json.Unmarshal([]byte(lessonsJSON.String), &hyp.LessonsLearned)
		}

		hypotheses = append(hypotheses, hyp)
	}

	return hypotheses, nil
}

// GetStrategyInsights retrieves insights for a strategy
func (r *ResearchDB) GetStrategyInsights(strategyID string, limit int) ([]InsightRecord, error) {
	query := `
		SELECT insight, category, confidence, created_at
		FROM insights
		WHERE strategy_id = ?
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := r.db.Query(query, strategyID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var insights []InsightRecord
	for rows.Next() {
		var ins InsightRecord
		var createdAt string
		var category sql.NullString

		err := rows.Scan(&ins.Insight, &category, &ins.Confidence, &createdAt)
		if err != nil {
			return nil, err
		}

		if category.Valid {
			ins.Category = category.String
		}

		ins.Timestamp, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		insights = append(insights, ins)
	}

	return insights, nil
}

// GetStrategyPatterns retrieves observed patterns
func (r *ResearchDB) GetStrategyPatterns(strategyID string) ([]PatternRecord, error) {
	query := `
		SELECT pattern, context, frequency, first_observed, last_observed, confidence, actionable
		FROM patterns
		WHERE strategy_id = ?
		ORDER BY frequency DESC, confidence DESC
	`

	rows, err := r.db.Query(query, strategyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var patterns []PatternRecord
	for rows.Next() {
		var pat PatternRecord
		var firstObs, lastObs string

		err := rows.Scan(&pat.Pattern, &pat.Context, &pat.Frequency, &firstObs, &lastObs, &pat.Confidence, &pat.Actionable)
		if err != nil {
			return nil, err
		}

		pat.FirstObserved, _ = time.Parse("2006-01-02 15:04:05", firstObs)
		pat.LastObserved, _ = time.Parse("2006-01-02 15:04:05", lastObs)

		patterns = append(patterns, pat)
	}

	return patterns, nil
}

// GetResearchSummary gets aggregated research summary
func (r *ResearchDB) GetResearchSummary(strategyID string) (*ResearchSummary, error) {
	summary := &ResearchSummary{}

	// Count hypotheses by status
	err := r.db.QueryRow(`
		SELECT 
			COUNT(*) as total,
			SUM(CASE WHEN status = 'confirmed' THEN 1 ELSE 0 END) as confirmed,
			SUM(CASE WHEN status = 'rejected' THEN 1 ELSE 0 END) as rejected,
			SUM(CASE WHEN status = 'active' THEN 1 ELSE 0 END) as active
		FROM hypotheses
		WHERE strategy_id = ?
	`, strategyID).Scan(&summary.TotalHypotheses, &summary.Confirmed, &summary.Rejected, &summary.Active)

	if err != nil {
		return nil, err
	}

	// Count iterations
	err = r.db.QueryRow(`
		SELECT COUNT(*) FROM iterations WHERE strategy_id = ?
	`, strategyID).Scan(&summary.TotalIterations)

	if err != nil {
		return nil, err
	}

	// Average improvement
	err = r.db.QueryRow(`
		SELECT AVG(improvement_pct) FROM iterations WHERE strategy_id = ?
	`, strategyID).Scan(&summary.AverageImprovement)

	if err != nil {
		return nil, err
	}

	// Count insights
	err = r.db.QueryRow(`
		SELECT COUNT(*) FROM insights WHERE strategy_id = ?
	`, strategyID).Scan(&summary.TotalInsights)

	if err != nil {
		return nil, err
	}

	// Count patterns
	err = r.db.QueryRow(`
		SELECT COUNT(*) FROM patterns WHERE strategy_id = ?
	`, strategyID).Scan(&summary.PatternsFound)

	if err != nil {
		return nil, err
	}

	return summary, nil
}

// Record types
type IterationRecord struct {
	IterationID string
	CodeVersion string
	Timestamp   time.Time
	Metrics     map[string]interface{}
	ChangesMade []string
	Rationale   string
	Improvement float64
}

type HypothesisRecord struct {
	ID             string
	Description    string
	Timestamp      time.Time
	Status         string
	RelatedCode    string
	Evaluation     map[string]interface{}
	LessonsLearned []string
}

type InsightRecord struct {
	Insight    string
	Category   string
	Confidence float64
	Timestamp  time.Time
}

type PatternRecord struct {
	Pattern       string
	Context       string
	Frequency     int
	FirstObserved time.Time
	LastObserved  time.Time
	Confidence    float64
	Actionable    bool
}

type ResearchSummary struct {
	TotalHypotheses    int
	Confirmed          int
	Rejected           int
	Active             int
	TotalIterations    int
	AverageImprovement float64
	TotalInsights      int
	PatternsFound      int
}
