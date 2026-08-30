package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ZulferDev/backtest-go/internal/backtest"
	"github.com/ZulferDev/backtest-go/internal/datafetcher"
	"github.com/ZulferDev/backtest-go/internal/validator"
	"github.com/ZulferDev/backtest-go/pkg/sdk"
)

// runLintPhase validates strategy code with AST checker
func runLintPhase(strategyID string, version int) error {
	strategyDir := filepath.Join(researchLogsDir, strategyID)
	strategyFile := filepath.Join(strategyDir, fmt.Sprintf("strategy_v%d.go", version))

	// Check if strategy file exists
	if _, err := os.Stat(strategyFile); os.IsNotExist(err) {
		return fmt.Errorf("strategy file not found: %s", strategyFile)
	}

	// Read strategy file
	src, err := os.ReadFile(strategyFile)
	if err != nil {
		return fmt.Errorf("failed to read strategy file: %w", err)
	}

	// Run AST validator
	validationErrors, err := validator.ValidateStrategy(strategyFile, src)

	// Log results
	logPath := filepath.Join(strategyDir, "validation_errors.log")
	logEntry := fmt.Sprintf("\n=== LINT v%d ===\n", version)
	
	if len(validationErrors) > 0 {
		logEntry += "Validation errors found:\n"
		for _, verr := range validationErrors {
			logEntry += fmt.Sprintf("  %s: %s\n", verr.Pos, verr.Message)
		}
	} else {
		logEntry += "✓ No validation errors\n"
	}
	
	logFile, _ := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if logFile != nil {
		logFile.WriteString(logEntry)
		logFile.Close()
	}

	// Return error if validation failed
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if len(validationErrors) > 0 {
		return fmt.Errorf("found %d validation error(s) - check validation_errors.log", len(validationErrors))
	}

	return nil
}

// runTestPhase runs unit tests for strategy
func runTestPhase(strategyID string, version int) error {
	strategyDir := filepath.Join(researchLogsDir, strategyID)
	strategyFile := filepath.Join(strategyDir, fmt.Sprintf("strategy_v%d.go", version))

	// Check if strategy file exists
	if _, err := os.Stat(strategyFile); os.IsNotExist(err) {
		return fmt.Errorf("strategy file not found: %s", strategyFile)
	}

	// Try to compile the strategy
	cmd := exec.Command("go", "build", "-o", "/dev/null", strategyFile)
	output, err := cmd.CombinedOutput()

	// Log output
	logPath := filepath.Join(strategyDir, "validation_errors.log")
	logEntry := fmt.Sprintf("\n=== TEST v%d ===\n%s\n", version, string(output))
	
	logFile, _ := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if logFile != nil {
		logFile.WriteString(logEntry)
		logFile.Close()
	}

	if err != nil {
		return fmt.Errorf("compilation failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// runBacktestPhase executes historical simulation
func runBacktestPhase(strategyID string, version int, dataFile string, initialCash float64) (string, error) {
	strategyDir := filepath.Join(researchLogsDir, strategyID)
	strategyFile := filepath.Join(strategyDir, fmt.Sprintf("strategy_v%d.go", version))
	resultsFile := filepath.Join(strategyDir, fmt.Sprintf("results_v%d.json", version))

	// Check if strategy file exists
	if _, err := os.Stat(strategyFile); os.IsNotExist(err) {
		return "", fmt.Errorf("strategy file not found: %s", strategyFile)
	}

	// Check if data file exists
	if _, err := os.Stat(dataFile); os.IsNotExist(err) {
		return "", fmt.Errorf("data file not found: %s", dataFile)
	}

	// Run backtest (this would call the actual backtest engine)
	// For now, we'll create a placeholder implementation
	fmt.Printf("  → Loading data from: %s\n", dataFile)
	fmt.Printf("  → Initial cash: $%.2f\n", initialCash)
	fmt.Printf("  → Running backtest...\n")

	// TODO: Integrate with actual backtest engine
	// cmd := exec.Command("go", "run", "cmd/backtest/main.go",
	//     "--strategy", strategyFile,
	//     "--data", dataFile,
	//     "--cash", fmt.Sprintf("%.2f", initialCash),
	//     "--output", resultsFile)
	// 
	// if err := cmd.Run(); err != nil {
	//     return "", fmt.Errorf("backtest failed: %w", err)
	// }

	// For now, create a placeholder results file
	placeholderResults := map[string]interface{}{
		"initial_cash": initialCash,
		"final_equity": initialCash * 1.1, // 10% return placeholder
		"total_return": 0.1,
		"sharpe_ratio": 1.5,
		"max_drawdown": 0.08,
		"total_trades": 50,
		"win_rate":     0.52,
		"note":         "Placeholder results - integrate with actual backtest engine",
	}

	if err := saveJSON(resultsFile, placeholderResults); err != nil {
		return "", fmt.Errorf("failed to save results: %w", err)
	}

	return resultsFile, nil
}

// runAnalyzePhase evaluates results and generates analysis
func runAnalyzePhase(strategyID string, version int, resultsFile string) (string, bool, error) {
	strategyDir := filepath.Join(researchLogsDir, strategyID)
	analysisFile := filepath.Join(strategyDir, fmt.Sprintf("analysis_v%d.json", version))

	// Check if results file exists
	if _, err := os.Stat(resultsFile); os.IsNotExist(err) {
		return "", false, fmt.Errorf("results file not found: %s", resultsFile)
	}

	// Load results
	var results map[string]interface{}
	if err := loadJSON(resultsFile, &results); err != nil {
		return "", false, fmt.Errorf("failed to load results: %w", err)
	}

	// TODO: This is where AI would analyze results
	// For now, create a placeholder analysis
	sharpeRatio := getFloat(results, "sharpe_ratio", 0.0)
	maxDrawdown := getFloat(results, "max_drawdown", 0.0)
	winRate := getFloat(results, "win_rate", 0.0)
	totalReturn := getFloat(results, "total_return", 0.0)

	analysis := Analysis{
		StrategyID: strategyID,
		Version:    fmt.Sprintf("v%d", version),
		BacktestResults: AnalysisResults{
			TotalReturn:  totalReturn,
			SharpeRatio:  sharpeRatio,
			MaxDrawdown:  maxDrawdown,
			WinRate:      winRate,
			TotalTrades:  int(getFloat(results, "total_trades", 0)),
		},
		HypothesisValidation: HypothesisValidation{
			ThesisConfirmed: sharpeRatio > 1.5 && maxDrawdown < 0.15,
			SuccessCriteriaMet: map[string]CriteriaCheck{
				"sharpe_ratio": {
					Target: 1.5,
					Actual: sharpeRatio,
					Met:    sharpeRatio > 1.5,
				},
				"max_drawdown": {
					Target: 0.15,
					Actual: maxDrawdown,
					Met:    maxDrawdown < 0.15,
				},
			},
			Evidence: []string{
				fmt.Sprintf("Sharpe ratio: %.2f", sharpeRatio),
				fmt.Sprintf("Max drawdown: %.2f%%", maxDrawdown*100),
			},
		},
		IdentifiedWeaknesses: []string{
			"Placeholder analysis - AI would identify real weaknesses",
		},
		NextIterationPlan: IterationPlan{
			Focus:               "Improve risk management",
			Rationale:           "Placeholder - AI would provide real rationale",
			ExpectedImprovement: "Reduce drawdown by 20%",
			ImplementationSteps: []string{
				"Add stop loss logic",
				"Implement position sizing",
			},
		},
	}

	// Save analysis
	if err := saveJSON(analysisFile, analysis); err != nil {
		return "", false, fmt.Errorf("failed to save analysis: %w", err)
	}

	// Determine if we should continue iterating
	shouldContinue := !analysis.HypothesisValidation.ThesisConfirmed

	return analysisFile, shouldContinue, nil
}

// Helper function to safely get float from map
func getFloat(m map[string]interface{}, key string, defaultVal float64) float64 {
	if val, ok := m[key]; ok {
		if f, ok := val.(float64); ok {
			return f
		}
	}
	return defaultVal
}
