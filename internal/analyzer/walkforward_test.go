package analyzer

import (
"os"
"testing"
"time"
)

func TestWalkForwardOrchestrator(t *testing.T) {
config := WalkForwardConfig{
TotalPeriodDays:   365,
InSampleDays:      90,
OutSampleDays:     30,
StepDays:          30,
MinTradesRequired: 10,
}

orchestrator := NewWalkForwardOrchestrator(config)
startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

err := orchestrator.InitializeSegments(startDate)
if err != nil {
t.Fatalf("Failed to initialize segments: %v", err)
}

if len(orchestrator.segments) == 0 {
t.Fatal("No segments created")
}

t.Logf("Created %d segments", len(orchestrator.segments))
}

func TestUpdateSegmentMetrics(t *testing.T) {
config := WalkForwardConfig{
TotalPeriodDays: 365,
InSampleDays:    90,
OutSampleDays:   30,
StepDays:        30,
}

orchestrator := NewWalkForwardOrchestrator(config)
startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
orchestrator.InitializeSegments(startDate)

isMetrics := SummaryMetrics{
TotalTrades:  50,
WinRate:      60.0,
TotalPnL:     1000.0,
TotalReturn:  10.0,
ProfitFactor: 1.8,
SharpeRatio:  1.5,
SortinoRatio: 2.0,
MaxDrawdown:  8.0,
AverageWin:   50.0,
AverageLoss:  -25.0,
}

oosMetrics := SummaryMetrics{
TotalTrades:  15,
WinRate:      53.0,
TotalPnL:     250.0,
TotalReturn:  7.5,
ProfitFactor: 1.5,
SharpeRatio:  1.2,
SortinoRatio: 1.6,
MaxDrawdown:  10.0,
AverageWin:   45.0,
AverageLoss:  -28.0,
}

err := orchestrator.UpdateSegmentMetrics("WF-001", isMetrics, oosMetrics)
if err != nil {
t.Fatalf("Failed to update segment metrics: %v", err)
}

found := false
for _, seg := range orchestrator.segments {
if seg.SegmentID == "WF-001" {
found = true
if seg.Status != "completed" {
t.Errorf("Expected status 'completed', got '%s'", seg.Status)
}
if seg.InSampleMetrics.TotalReturn != 10.0 {
t.Errorf("Expected IS return 10.0, got %.2f", seg.InSampleMetrics.TotalReturn)
}
break
}
}

if !found {
t.Error("Segment WF-001 not found after update")
}
}

func TestAnalyzeWalkForward(t *testing.T) {
config := WalkForwardConfig{
TotalPeriodDays: 365,
InSampleDays:    90,
OutSampleDays:   30,
StepDays:        30,
}

orchestrator := NewWalkForwardOrchestrator(config)
startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
orchestrator.InitializeSegments(startDate)

segments := []struct {
id        string
isReturn  float64
oosReturn float64
isSharpe  float64
oosSharpe float64
}{
{"WF-001", 12.0, 9.0, 1.6, 1.3},
{"WF-002", 8.0, 6.5, 1.2, 1.0},
{"WF-003", 15.0, 11.0, 2.0, 1.5},
{"WF-004", 10.0, 8.0, 1.4, 1.1},
{"WF-005", 5.0, 3.0, 0.8, 0.5},
}

for _, s := range segments {
isMetrics := SummaryMetrics{
TotalTrades:  50,
WinRate:      58.0,
TotalReturn:  s.isReturn,
SharpeRatio:  s.isSharpe,
AverageWin:   45.0,
AverageLoss:  -25.0,
}
oosMetrics := SummaryMetrics{
TotalTrades:  15,
WinRate:      53.0,
TotalReturn:  s.oosReturn,
SharpeRatio:  s.oosSharpe,
AverageWin:   42.0,
AverageLoss:  -27.0,
}
orchestrator.UpdateSegmentMetrics(s.id, isMetrics, oosMetrics)
}

result, err := orchestrator.Analyze()
if err != nil {
t.Fatalf("Analysis failed: %v", err)
}

t.Logf("Overfitting Score: %.3f", result.OverfittingScore)
t.Logf("Risk Level: %s", result.RiskLevel)
t.Logf("Success Rate: %.1f%%", result.OverallMetrics.SuccessRate)

if result.OverallMetrics.TotalSegments != 5 {
t.Errorf("Expected 5 segments, got %d", result.OverallMetrics.TotalSegments)
}
}

func TestOverfittingScoreCalculation(t *testing.T) {
tests := []struct {
name              string
returnDegradation float64
sharpeDegradation float64
consistencyScore  float64
successRate       float64
expectedMaxScore  float64
}{
{"Low Risk", 10.0, 10.0, 2.0, 100.0, 0.3},
{"Medium Risk", 30.0, 35.0, 5.0, 60.0, 0.6},
{"High Risk", 60.0, 70.0, 10.0, 20.0, 1.0},
}

for _, tt := range tests {
t.Run(tt.name, func(t *testing.T) {
result := &WalkForwardResult{
OverallMetrics: WalkForwardMetrics{
ReturnDegradation: tt.returnDegradation,
SharpeDegradation: tt.sharpeDegradation,
ConsistencyScore:  tt.consistencyScore,
SuccessRate:       tt.successRate,
},
}

score := calculateOverfittingScore(result)
t.Logf("Calculated score: %.3f", score)

if score > tt.expectedMaxScore {
t.Errorf("Expected score <= %.2f, got %.3f", tt.expectedMaxScore, score)
}
})
}
}

func TestAssessOverfittingRisk(t *testing.T) {
tests := []struct {
score           float64
expectedRisk    string
expectedKeyword string
}{
{0.15, "low", "robust"},
{0.45, "medium", "Moderate"},
{0.75, "high", "curve-fitted"},
}

for _, tt := range tests {
result := &WalkForwardResult{OverfittingScore: tt.score}
risk, recommendation := assessOverfittingRisk(result)

if risk != tt.expectedRisk {
t.Errorf("For score %.2f, expected risk '%s', got '%s'", tt.score, tt.expectedRisk, risk)
}

if tt.expectedKeyword != "" && !stringContains(recommendation, tt.expectedKeyword) {
t.Errorf("Recommendation should contain '%s': %s", tt.expectedKeyword, recommendation)
}
}
}

func TestSaveAndLoadResults(t *testing.T) {
result := &WalkForwardResult{
Config: WalkForwardConfig{
TotalPeriodDays: 365,
InSampleDays:    90,
OutSampleDays:   30,
StepDays:        30,
},
OverfittingScore: 0.25,
RiskLevel:        "low",
Recommendation:   "Strategy shows robust performance",
OverallMetrics: WalkForwardMetrics{
AvgInSampleReturn:  10.0,
AvgOutSampleReturn: 8.5,
ReturnDegradation:  15.0,
SuccessRate:        80.0,
},
}

tmpFile := "/tmp/walkforward_test_results.json"
defer os.Remove(tmpFile)

err := result.SaveResults(tmpFile)
if err != nil {
t.Fatalf("Failed to save results: %v", err)
}

loaded, err := LoadResults(tmpFile)
if err != nil {
t.Fatalf("Failed to load results: %v", err)
}

if loaded.OverfittingScore != result.OverfittingScore {
t.Errorf("Expected overfitting score %.3f, got %.3f", result.OverfittingScore, loaded.OverfittingScore)
}
}

func TestGenerateSummaryReport(t *testing.T) {
result := &WalkForwardResult{
Config: WalkForwardConfig{
TotalPeriodDays: 365,
InSampleDays:    90,
OutSampleDays:   30,
StepDays:        30,
},
OverfittingScore: 0.22,
RiskLevel:        "low",
OverallMetrics: WalkForwardMetrics{
AvgInSampleReturn:  10.5,
AvgOutSampleReturn: 9.2,
ReturnDegradation:  12.4,
SuccessRate:        85.0,
TotalSegments:      6,
},
}

report := result.GenerateSummaryReport()

requiredKeywords := []string{
"Walk-Forward Analysis Summary",
"Configuration",
"Overall Performance",
"Robustness Metrics",
"Overfitting Score",
"Recommendation",
}

for _, keyword := range requiredKeywords {
if !stringContains(report, keyword) {
t.Errorf("Report missing required section: %s", keyword)
}
}
}

func TestInvalidConfig(t *testing.T) {
config := WalkForwardConfig{
InSampleDays:  0,
OutSampleDays: 30,
StepDays:      30,
}

orchestrator := NewWalkForwardOrchestrator(config)
err := orchestrator.InitializeSegments(time.Now())

if err == nil {
t.Error("Expected error for invalid config, got nil")
}
}

func stringContains(s, substr string) bool {
for i := 0; i <= len(s)-len(substr); i++ {
if s[i:i+len(substr)] == substr {
return true
}
}
return false
}
