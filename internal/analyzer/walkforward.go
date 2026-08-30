package analyzer

import (
"encoding/json"
"fmt"
"os"
"time"
)

// WalkForwardConfig defines walk-forward testing parameters
type WalkForwardConfig struct {
TotalPeriodDays   int `json:"total_period_days"`   // Total backtest period in days
InSampleDays      int `json:"in_sample_days"`      // Training period length
OutSampleDays     int `json:"out_sample_days"`     // Testing period length
StepDays          int `json:"step_days"`           // Rolling window step size
MinTradesRequired int `json:"min_trades_required"` // Minimum trades per segment
}

// WalkForwardSegment represents a single IS/OOS segment
type WalkForwardSegment struct {
SegmentID       string         `json:"segment_id"`
InSampleStart   time.Time      `json:"in_sample_start"`
InSampleEnd     time.Time      `json:"in_sample_end"`
OutSampleStart  time.Time      `json:"out_sample_start"`
OutSampleEnd    time.Time      `json:"out_sample_end"`
InSampleMetrics SummaryMetrics `json:"in_sample_metrics"`
OutSampleMetrics SummaryMetrics `json:"out_sample_metrics"`
Status          string         `json:"status"` // "pending", "completed", "failed"
}

// WalkForwardResult contains complete walk-forward analysis results
type WalkForwardResult struct {
Config           WalkForwardConfig      `json:"config"`
Segments         []WalkForwardSegment   `json:"segments"`
OverallMetrics   WalkForwardMetrics     `json:"overall_metrics"`
OverfittingScore float64                `json:"overfitting_score"` // 0-1, lower is better
RiskLevel        string                 `json:"risk_level"`        // "low", "medium", "high"
Recommendation   string                 `json:"recommendation"`
}

// WalkForwardMetrics aggregates walk-forward performance statistics
type WalkForwardMetrics struct {
AvgInSampleReturn     float64 `json:"avg_in_sample_return"`
AvgOutSampleReturn    float64 `json:"avg_out_sample_return"`
ReturnDegradation     float64 `json:"return_degradation"` // % drop from IS to OOS
AvgInSampleSharpe     float64 `json:"avg_in_sample_sharpe"`
AvgOutSampleSharpe    float64 `json:"avg_out_sample_sharpe"`
SharpeDegradation     float64 `json:"sharpe_degradation"`
ConsistencyScore      float64 `json:"consistency_score"` // How consistent OOS results are
ProfitableSegments    int     `json:"profitable_segments"`
TotalSegments         int     `json:"total_segments"`
SuccessRate           float64 `json:"success_rate"` // % of segments where OOS was profitable
}

// NewWalkForwardOrchestrator creates a new walk-forward orchestrator
func NewWalkForwardOrchestrator(config WalkForwardConfig) *WalkForwardOrchestrator {
return &WalkForwardOrchestrator{
config:   config,
segments: []WalkForwardSegment{},
}
}

// WalkForwardOrchestrator manages walk-forward testing workflow
type WalkForwardOrchestrator struct {
config      WalkForwardConfig
segments    []WalkForwardSegment
}

// InitializeSegments creates walk-forward segments based on config
func (w *WalkForwardOrchestrator) InitializeSegments(startDate time.Time) error {
if w.config.InSampleDays <= 0 || w.config.OutSampleDays <= 0 || w.config.StepDays <= 0 {
return fmt.Errorf("invalid config: all period values must be positive")
}

currentStart := startDate
segmentNum := 0

for {
inSampleEnd := currentStart.AddDate(0, 0, w.config.InSampleDays)
outSampleStart := inSampleEnd
outSampleEnd := outSampleStart.AddDate(0, 0, w.config.OutSampleDays)

// Check if we've exceeded a reasonable end date (e.g., 2 years from start)
if outSampleEnd.After(startDate.AddDate(2, 0, 0)) {
break
}

segmentNum++
segment := WalkForwardSegment{
SegmentID:      fmt.Sprintf("WF-%03d", segmentNum),
InSampleStart:  currentStart,
InSampleEnd:    inSampleEnd,
OutSampleStart: outSampleStart,
OutSampleEnd:   outSampleEnd,
Status:         "pending",
}
w.segments = append(w.segments, segment)

// Move to next window
currentStart = currentStart.AddDate(0, 0, w.config.StepDays)

// Safety limit
if segmentNum > 50 {
break
}
}

if len(w.segments) == 0 {
return fmt.Errorf("no segments could be created with current configuration")
}

return nil
}

// UpdateSegmentMetrics updates metrics for a specific segment
func (w *WalkForwardOrchestrator) UpdateSegmentMetrics(segmentID string, isMetrics, oosMetrics SummaryMetrics) error {
for i, seg := range w.segments {
if seg.SegmentID == segmentID {
w.segments[i].InSampleMetrics = isMetrics
w.segments[i].OutSampleMetrics = oosMetrics
w.segments[i].Status = "completed"
return nil
}
}
return fmt.Errorf("segment %s not found", segmentID)
}

// Analyze performs complete walk-forward analysis
func (w *WalkForwardOrchestrator) Analyze() (*WalkForwardResult, error) {
if len(w.segments) == 0 {
return nil, fmt.Errorf("no segments to analyze")
}

result := &WalkForwardResult{
Config:   w.config,
Segments: w.segments,
}

var (
totalISReturn, totalOOSReturn     float64
totalISSharpe, totalAOSSharpe     float64
profitableSegments                int
oosReturns                        []float64
)

completedSegments := 0
for _, seg := range w.segments {
if seg.Status != "completed" {
continue
}
completedSegments++

totalISReturn += seg.InSampleMetrics.TotalReturn
totalOOSReturn += seg.OutSampleMetrics.TotalReturn
totalISSharpe += seg.InSampleMetrics.SharpeRatio
totalAOSSharpe += seg.OutSampleMetrics.SharpeRatio
oosReturns = append(oosReturns, seg.OutSampleMetrics.TotalReturn)

if seg.OutSampleMetrics.TotalReturn > 0 {
profitableSegments++
}
}

if completedSegments == 0 {
result.RiskLevel = "unknown"
result.Recommendation = "No completed segments to analyze"
return result, nil
}

// Calculate averages
result.OverallMetrics.AvgInSampleReturn = totalISReturn / float64(completedSegments)
result.OverallMetrics.AvgOutSampleReturn = totalOOSReturn / float64(completedSegments)
result.OverallMetrics.AvgInSampleSharpe = totalISSharpe / float64(completedSegments)
result.OverallMetrics.AvgOutSampleSharpe = totalAOSSharpe / float64(completedSegments)
result.OverallMetrics.ProfitableSegments = profitableSegments
result.OverallMetrics.TotalSegments = completedSegments
result.OverallMetrics.SuccessRate = float64(profitableSegments) / float64(completedSegments) * 100

// Calculate degradation
if result.OverallMetrics.AvgInSampleReturn != 0 {
result.OverallMetrics.ReturnDegradation = (result.OverallMetrics.AvgInSampleReturn - result.OverallMetrics.AvgOutSampleReturn) / result.OverallMetrics.AvgInSampleReturn * 100
}
if result.OverallMetrics.AvgInSampleSharpe != 0 {
result.OverallMetrics.SharpeDegradation = (result.OverallMetrics.AvgInSampleSharpe - result.OverallMetrics.AvgOutSampleSharpe) / result.OverallMetrics.AvgInSampleSharpe * 100
}

// Calculate consistency score (std dev of OOS returns)
result.OverallMetrics.ConsistencyScore = calculateStdDev(oosReturns)

// Calculate overfitting score (0-1, lower is better)
result.OverfittingScore = calculateOverfittingScore(result)

// Determine risk level and recommendation
result.RiskLevel, result.Recommendation = assessOverfittingRisk(result)

return result, nil
}

func calculateStdDev(values []float64) float64 {
if len(values) == 0 {
return 0
}

mean := 0.0
for _, v := range values {
mean += v
}
mean /= float64(len(values))

variance := 0.0
for _, v := range values {
diff := v - mean
variance += diff * diff
}
variance /= float64(len(values))

return sqrt(variance)
}

func sqrt(x float64) float64 {
if x <= 0 {
return 0
}
z := x
for i := 0; i < 10; i++ {
z = (z + x/z) / 2
}
return z
}

func calculateOverfittingScore(result *WalkForwardResult) float64 {
score := 0.0

// Factor 1: Return degradation (0-0.4)
degradationFactor := minFloat(result.OverallMetrics.ReturnDegradation/50.0, 1.0) * 0.4
score += degradationFactor

// Factor 2: Sharpe degradation (0-0.3)
sharpeDegradationFactor := minFloat(result.OverallMetrics.SharpeDegradation/50.0, 1.0) * 0.3
score += sharpeDegradationFactor

// Factor 3: Consistency penalty (0-0.2)
consistencyFactor := minFloat(result.OverallMetrics.ConsistencyScore/10.0, 1.0) * 0.2
score += consistencyFactor

// Factor 4: Success rate bonus (0-0.1, inverted)
successPenalty := (1.0 - result.OverallMetrics.SuccessRate/100.0) * 0.1
score += successPenalty

return minFloat(score, 1.0)
}

func assessOverfittingRisk(result *WalkForwardResult) (string, string) {
score := result.OverfittingScore

if score < 0.3 {
return "low", "Strategy shows robust performance across IS/OOS periods. Low overfitting risk detected."
} else if score < 0.6 {
return "medium", "Moderate degradation between IS/OOS. Consider additional validation or parameter stabilization."
} else {
return "high", "High overfitting risk! Significant performance drop in OOS. Strategy likely curve-fitted. Major revision needed."
}
}

func minFloat(a, b float64) float64 {
if a < b {
return a
}
return b
}

// SaveResults saves walk-forward results to JSON file
func (r *WalkForwardResult) SaveResults(filepath string) error {
data, err := json.MarshalIndent(r, "", "  ")
if err != nil {
return fmt.Errorf("failed to marshal results: %w", err)
}
return os.WriteFile(filepath, data, 0644)
}

// LoadResults loads walk-forward results from JSON file
func LoadResults(filepath string) (*WalkForwardResult, error) {
data, err := os.ReadFile(filepath)
if err != nil {
return nil, fmt.Errorf("failed to read results file: %w", err)
}

var result WalkForwardResult
if err := json.Unmarshal(data, &result); err != nil {
return nil, fmt.Errorf("failed to parse results JSON: %w", err)
}

return &result, nil
}

// GetSegmentReport generates detailed report for a specific segment
func (r *WalkForwardResult) GetSegmentReport(segmentID string) string {
for _, seg := range r.Segments {
if seg.SegmentID == segmentID {
report := fmt.Sprintf("=== Segment %s ===\n", segmentID)
report += fmt.Sprintf("In-Sample Period: %s to %s\n", seg.InSampleStart.Format("2006-01-02"), seg.InSampleEnd.Format("2006-01-02"))
report += fmt.Sprintf("Out-of-Sample Period: %s to %s\n", seg.OutSampleStart.Format("2006-01-02"), seg.OutSampleEnd.Format("2006-01-02"))
report += "\nIn-Sample Metrics:\n"
report += fmt.Sprintf("  Return: %.2f%% | Sharpe: %.3f | Win Rate: %.1f%%\n", seg.InSampleMetrics.TotalReturn, seg.InSampleMetrics.SharpeRatio, seg.InSampleMetrics.WinRate)
report += "\nOut-of-Sample Metrics:\n"
report += fmt.Sprintf("  Return: %.2f%% | Sharpe: %.3f | Win Rate: %.1f%%\n", seg.OutSampleMetrics.TotalReturn, seg.OutSampleMetrics.SharpeRatio, seg.OutSampleMetrics.WinRate)

returnDegradation := 0.0
if seg.InSampleMetrics.TotalReturn != 0 {
returnDegradation = (seg.InSampleMetrics.TotalReturn - seg.OutSampleMetrics.TotalReturn) / seg.InSampleMetrics.TotalReturn * 100
}
report += fmt.Sprintf("\nDegradation: %.1f%%\n", returnDegradation)
report += fmt.Sprintf("Status: %s\n", seg.Status)

return report
}
}
return fmt.Sprintf("Segment %s not found", segmentID)
}

// GenerateSummaryReport generates overall summary report
func (r *WalkForwardResult) GenerateSummaryReport() string {
report := "# Walk-Forward Analysis Summary\n\n"
report += "## Configuration\n\n"
report += fmt.Sprintf("- Total Period: %d days\n", r.Config.TotalPeriodDays)
report += fmt.Sprintf("- In-Sample Window: %d days\n", r.Config.InSampleDays)
report += fmt.Sprintf("- Out-of-Sample Window: %d days\n", r.Config.OutSampleDays)
report += fmt.Sprintf("- Step Size: %d days\n", r.Config.StepDays)
report += "\n"

report += "## Overall Performance\n\n"
report += "| Metric | In-Sample | Out-of-Sample | Degradation |\n"
report += "|--------|-----------|---------------|-------------|\n"
report += fmt.Sprintf("| Avg Return | %.2f%% | %.2f%% | %.1f%% |\n", 
r.OverallMetrics.AvgInSampleReturn, 
r.OverallMetrics.AvgOutSampleReturn,
r.OverallMetrics.ReturnDegradation)
report += fmt.Sprintf("| Avg Sharpe | %.3f | %.3f | %.1f%% |\n",
r.OverallMetrics.AvgInSampleSharpe,
r.OverallMetrics.AvgOutSampleSharpe,
r.OverallMetrics.SharpeDegradation)
report += "\n"

report += "## Robustness Metrics\n\n"
report += fmt.Sprintf("- Overfitting Score: %.3f (0=perfect, 1=severe overfitting)\n", r.OverfittingScore)
report += fmt.Sprintf("- Risk Level: %s\n", r.RiskLevel)
report += fmt.Sprintf("- Success Rate: %.1f%% (%d/%d segments profitable)\n", 
r.OverallMetrics.SuccessRate,
r.OverallMetrics.ProfitableSegments,
r.OverallMetrics.TotalSegments)
report += fmt.Sprintf("- Consistency Score: %.3f (lower is better)\n", r.OverallMetrics.ConsistencyScore)
report += "\n"

report += fmt.Sprintf("## Recommendation\n\n%s\n", r.Recommendation)

return report
}
