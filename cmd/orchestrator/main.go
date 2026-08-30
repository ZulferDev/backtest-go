package main

import (
	"flag"
	"fmt"
	"os"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "init":
		initCmd()
	case "run":
		runCmd()
	case "phase":
		phaseCmd()
	case "history":
		historyCmd()
	case "version":
		fmt.Printf("orchestrator v%s\n", version)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Orchestrator - AI Strategy Research Automation")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  orchestrator <command> [flags]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  init       Initialize new research session")
	fmt.Println("  run        Run full strategy lifecycle (CONCEIVE -> ANALYZE)")
	fmt.Println("  phase      Execute single phase (for debugging)")
	fmt.Println("  history    View strategy iteration history")
	fmt.Println("  version    Show version")
	fmt.Println("  help       Show this help message")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Initialize new strategy")
	fmt.Println("  orchestrator init --strategy-id rsi_mean_reversion")
	fmt.Println()
	fmt.Println("  # Run full cycle")
	fmt.Println("  orchestrator run --strategy-id rsi_mean_reversion --data data/BTCUSDT_1h.json")
	fmt.Println()
	fmt.Println("  # Run single phase")
	fmt.Println("  orchestrator phase --strategy-id rsi_mean_reversion --phase BACKTEST")
	fmt.Println()
	fmt.Println("  # View history")
	fmt.Println("  orchestrator history --strategy-id rsi_mean_reversion")
}

func initCmd() {
	fs := flag.NewFlagSet("init", flag.ExitOnError)
	strategyID := fs.String("strategy-id", "", "Strategy identifier (required)")
	hypothesisFile := fs.String("hypothesis-file", "", "Path to hypothesis.md file (optional)")

	fs.Parse(os.Args[2:])

	if *strategyID == "" {
		fmt.Println("Error: --strategy-id is required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	if err := initializeStrategy(*strategyID, *hypothesisFile); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ Strategy '%s' initialized in research_logs/%s/\n", *strategyID, *strategyID)
}

func runCmd() {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	strategyID := fs.String("strategy-id", "", "Strategy identifier (required)")
	dataFile := fs.String("data", "", "Path to OHLCV data file (required)")
	initialCash := fs.Float64("cash", 10000.0, "Initial cash for backtest")
	maxIterations := fs.Int("max-iterations", 5, "Maximum refinement iterations")
	skipValidation := fs.Bool("skip-validation", false, "Skip AST validation (dangerous)")

	fs.Parse(os.Args[2:])

	if *strategyID == "" || *dataFile == "" {
		fmt.Println("Error: --strategy-id and --data are required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	config := RunConfig{
		StrategyID:     *strategyID,
		DataFile:       *dataFile,
		InitialCash:    *initialCash,
		MaxIterations:  *maxIterations,
		SkipValidation: *skipValidation,
	}

	if err := runFullCycle(config); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func phaseCmd() {
	fs := flag.NewFlagSet("phase", flag.ExitOnError)
	strategyID := fs.String("strategy-id", "", "Strategy identifier (required)")
	phase := fs.String("phase", "", "Phase to run: LINT, TEST, BACKTEST, ANALYZE (required)")
	dataFile := fs.String("data", "", "Path to OHLCV data file (required for BACKTEST)")
	initialCash := fs.Float64("cash", 10000.0, "Initial cash for backtest")

	fs.Parse(os.Args[2:])

	if *strategyID == "" || *phase == "" {
		fmt.Println("Error: --strategy-id and --phase are required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	config := PhaseConfig{
		StrategyID:  *strategyID,
		Phase:       *phase,
		DataFile:    *dataFile,
		InitialCash: *initialCash,
	}

	if err := runPhase(config); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func historyCmd() {
	fs := flag.NewFlagSet("history", flag.ExitOnError)
	strategyID := fs.String("strategy-id", "", "Strategy identifier (required)")
	verbose := fs.Bool("verbose", false, "Show detailed results for each version")

	fs.Parse(os.Args[2:])

	if *strategyID == "" {
		fmt.Println("Error: --strategy-id is required")
		fs.PrintDefaults()
		os.Exit(1)
	}

	if err := showHistory(*strategyID, *verbose); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

type RunConfig struct {
	StrategyID     string
	DataFile       string
	InitialCash    float64
	MaxIterations  int
	SkipValidation bool
}

type PhaseConfig struct {
	StrategyID  string
	Phase       string
	DataFile    string
	InitialCash float64
}
