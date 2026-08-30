package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

const researchLogsDir = "research_logs"

// initializeStrategy creates directory structure for new strategy
func initializeStrategy(strategyID, hypothesisFile string) error {
	strategyDir := filepath.Join(researchLogsDir, strategyID)

	// Check if already exists
	if _, err := os.Stat(strategyDir); err == nil {
		return fmt.Errorf("strategy '%s' already exists", strategyID)
	}

	// Create directory
	if err := os.MkdirAll(strategyDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Copy hypothesis template or user-provided file
	if hypothesisFile != "" {
		// Copy user-provided hypothesis
		if err := copyFile(hypothesisFile, filepath.Join(strategyDir, "hypothesis.md")); err != nil {
			return fmt.Errorf("failed to copy hypothesis: %w", err)
		}
	} else {
		// Copy template
		templatePath := filepath.Join(researchLogsDir, ".template", "hypothesis.md")
		destPath := filepath.Join(strategyDir, "hypothesis.md")
		if err := copyFile(templatePath, destPath); err != nil {
			return fmt.Errorf("failed to copy hypothesis template: %w", err)
		}
	}

	// Initialize memory.json
	memory := Memory{
		StrategyID:       strategyID,
		StrategyLineage:  []StrategyVersion{},
		LearnedPatterns:  []string{},
		FailedApproaches: []string{},
		MarketInsights:   []string{},
	}

	memoryPath := filepath.Join(strategyDir, "memory.json")
	if err := saveJSON(memoryPath, memory); err != nil {
		return fmt.Errorf("failed to initialize memory.json: %w", err)
	}

	// Create validation_errors.log (empty)
	logPath := filepath.Join(strategyDir, "validation_errors.log")
	if err := os.WriteFile(logPath, []byte(""), 0644); err != nil {
		return fmt.Errorf("failed to create validation_errors.log: %w", err)
	}

	return nil
}

// runFullCycle executes complete strategy lifecycle
func runFullCycle(config RunConfig) error {
	strategyDir := filepath.Join(researchLogsDir, config.StrategyID)

	// Verify strategy exists
	if _, err := os.Stat(strategyDir); os.IsNotExist(err) {
		return fmt.Errorf("strategy '%s' not found. Run 'init' first", config.StrategyID)
	}

	fmt.Printf("Starting full cycle for strategy: %s\n", config.StrategyID)
	fmt.Printf("Max iterations: %d\n", config.MaxIterations)
	fmt.Println()

	// Load memory
	memory, err := loadMemory(config.StrategyID)
	if err != nil {
		return fmt.Errorf("failed to load memory: %w", err)
	}

	currentVersion := len(memory.StrategyLineage) + 1

	for iteration := 0; iteration < config.MaxIterations; iteration++ {
		fmt.Printf("=== Iteration %d (Version v%d) ===\n", iteration+1, currentVersion)

		// Phase 1: LINT (validate existing strategy code)
		fmt.Println("Phase 1: LINT - Validating strategy code...")
		if !config.SkipValidation {
			if err := runLintPhase(config.StrategyID, currentVersion); err != nil {
				fmt.Printf("❌ LINT failed: %v\n", err)
				fmt.Println("Fix validation errors before continuing.")
				return err
			}
			fmt.Println("✓ LINT passed")
		} else {
			fmt.Println("⚠ LINT skipped (--skip-validation)")
		}

		// Phase 2: TEST (run unit tests)
		fmt.Println("Phase 2: TEST - Running unit tests...")
		if err := runTestPhase(config.StrategyID, currentVersion); err != nil {
			fmt.Printf("❌ TEST failed: %v\n", err)
			return err
		}
		fmt.Println("✓ TEST passed")

		// Phase 3: BACKTEST (execute historical simulation)
		fmt.Println("Phase 3: BACKTEST - Running historical simulation...")
		resultsFile, err := runBacktestPhase(config.StrategyID, currentVersion, config.DataFile, config.InitialCash)
		if err != nil {
			fmt.Printf("❌ BACKTEST failed: %v\n", err)
			return err
		}
		fmt.Printf("✓ BACKTEST complete: %s\n", resultsFile)

		// Phase 4: ANALYZE (evaluate results and plan next iteration)
		fmt.Println("Phase 4: ANALYZE - Evaluating results...")
		analysisFile, shouldContinue, err := runAnalyzePhase(config.StrategyID, currentVersion, resultsFile)
		if err != nil {
			fmt.Printf("❌ ANALYZE failed: %v\n", err)
			return err
		}
		fmt.Printf("✓ ANALYZE complete: %s\n", analysisFile)

		// Update memory
		if err := updateMemoryFromAnalysis(config.StrategyID, currentVersion, analysisFile); err != nil {
			fmt.Printf("⚠ Failed to update memory: %v\n", err)
		}

		fmt.Println()

		// Check if we should continue iterating
		if !shouldContinue {
			fmt.Println("✓ Strategy meets success criteria. Stopping iteration.")
			break
		}

		if iteration < config.MaxIterations-1 {
			fmt.Println("→ Next iteration: refine strategy based on analysis")
			fmt.Println()
		}

		currentVersion++
	}

	fmt.Printf("\n✓ Full cycle complete for strategy: %s\n", config.StrategyID)
	fmt.Printf("Final version: v%d\n", currentVersion)

	return nil
}

// runPhase executes a single phase
func runPhase(config PhaseConfig) error {
	strategyDir := filepath.Join(researchLogsDir, config.StrategyID)

	// Verify strategy exists
	if _, err := os.Stat(strategyDir); os.IsNotExist(err) {
		return fmt.Errorf("strategy '%s' not found", config.StrategyID)
	}

	// Get current version
	memory, err := loadMemory(config.StrategyID)
	if err != nil {
		return fmt.Errorf("failed to load memory: %w", err)
	}
	currentVersion := len(memory.StrategyLineage) + 1

	fmt.Printf("Running phase: %s (Strategy: %s, Version: v%d)\n", config.Phase, config.StrategyID, currentVersion)

	switch config.Phase {
	case "LINT":
		return runLintPhase(config.StrategyID, currentVersion)
	case "TEST":
		return runTestPhase(config.StrategyID, currentVersion)
	case "BACKTEST":
		if config.DataFile == "" {
			return fmt.Errorf("--data is required for BACKTEST phase")
		}
		_, err := runBacktestPhase(config.StrategyID, currentVersion, config.DataFile, config.InitialCash)
		return err
	case "ANALYZE":
		resultsFile := filepath.Join(strategyDir, fmt.Sprintf("results_v%d.json", currentVersion))
		_, _, err := runAnalyzePhase(config.StrategyID, currentVersion, resultsFile)
		return err
	default:
		return fmt.Errorf("unknown phase: %s (valid: LINT, TEST, BACKTEST, ANALYZE)", config.Phase)
	}
}

// showHistory displays strategy iteration history
func showHistory(strategyID string, verbose bool) error {
	memory, err := loadMemory(strategyID)
	if err != nil {
		return err
	}

	fmt.Printf("Strategy: %s\n", strategyID)
	fmt.Printf("Total iterations: %d\n", len(memory.StrategyLineage))
	fmt.Println()

	if len(memory.StrategyLineage) == 0 {
		fmt.Println("No iterations yet.")
		return nil
	}

	for _, version := range memory.StrategyLineage {
		fmt.Printf("Version %s (%s)\n", version.Version, version.Date)
		fmt.Printf("  Changes: %s\n", version.Changes)

		if verbose {
			fmt.Printf("  Results:\n")
			fmt.Printf("    Sharpe Ratio: %.2f\n", version.BacktestResults.SharpeRatio)
			fmt.Printf("    Max Drawdown: %.2f%%\n", version.BacktestResults.MaxDrawdown*100)
			fmt.Printf("    Win Rate: %.2f%%\n", version.BacktestResults.WinRate*100)
			fmt.Printf("    Total Return: %.2f%%\n", version.BacktestResults.TotalReturn*100)
		}

		fmt.Printf("  Lesson: %s\n", version.Lesson)
		fmt.Println()
	}

	if len(memory.LearnedPatterns) > 0 {
		fmt.Println("Learned Patterns:")
		for _, pattern := range memory.LearnedPatterns {
			fmt.Printf("  • %s\n", pattern)
		}
		fmt.Println()
	}

	if len(memory.FailedApproaches) > 0 {
		fmt.Println("Failed Approaches:")
		for _, approach := range memory.FailedApproaches {
			fmt.Printf("  • %s\n", approach)
		}
	}

	return nil
}

// Utility functions

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

func saveJSON(path string, v interface{}) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func loadJSON(path string, v interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func loadMemory(strategyID string) (*Memory, error) {
	memoryPath := filepath.Join(researchLogsDir, strategyID, "memory.json")
	var memory Memory
	if err := loadJSON(memoryPath, &memory); err != nil {
		return nil, err
	}
	return &memory, nil
}

func updateMemoryFromAnalysis(strategyID string, version int, analysisFile string) error {
	// Load current memory
	memory, err := loadMemory(strategyID)
	if err != nil {
		return err
	}

	// Load analysis
	var analysis Analysis
	if err := loadJSON(analysisFile, &analysis); err != nil {
		return err
	}

	// Create new version entry
	newVersion := StrategyVersion{
		Version: fmt.Sprintf("v%d", version),
		Date:    time.Now().Format("2006-01-02"),
		Changes: analysis.NextIterationPlan.Focus,
		BacktestResults: BacktestSummary{
			SharpeRatio:  analysis.BacktestResults.SharpeRatio,
			MaxDrawdown:  analysis.BacktestResults.MaxDrawdown,
			WinRate:      analysis.BacktestResults.WinRate,
			TotalReturn:  analysis.BacktestResults.TotalReturn,
		},
		Lesson: analysis.NextIterationPlan.Rationale,
	}

	memory.StrategyLineage = append(memory.StrategyLineage, newVersion)

	// Update learned patterns and failed approaches from analysis
	// (This would be done by AI in real implementation)

	// Save updated memory
	memoryPath := filepath.Join(researchLogsDir, strategyID, "memory.json")
	return saveJSON(memoryPath, memory)
}
