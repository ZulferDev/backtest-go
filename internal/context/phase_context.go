package context

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// PhaseContext manages isolated context for each AI lifecycle phase
type PhaseContext struct {
	StrategyID  string                 `json:"strategy_id"`
	Phase       string                 `json:"phase"` // CONCEIVE, WRITE, LINT, TEST, BACKTEST, ANALYZE
	Version     int                    `json:"version"`
	InputFiles  []string               `json:"input_files"`  // Files to read for this phase
	OutputFiles []string               `json:"output_files"` // Files to write in this phase
	Metadata    map[string]interface{} `json:"metadata"`
}

// ContextManager handles phase isolation and focused prompts
type ContextManager struct {
	baseDir string
}

// NewContextManager creates a new context manager
func NewContextManager(baseDir string) *ContextManager {
	return &ContextManager{baseDir: baseDir}
}

// GetPhaseContext returns isolated context for a specific phase
func (cm *ContextManager) GetPhaseContext(strategyID, phase string, version int) (*PhaseContext, error) {
	ctx := &PhaseContext{
		StrategyID:  strategyID,
		Phase:       phase,
		Version:     version,
		InputFiles:  []string{},
		OutputFiles: []string{},
		Metadata:    make(map[string]interface{}),
	}

	strategyDir := filepath.Join(cm.baseDir, strategyID)

	// Define phase-specific input/output files
	switch phase {
	case "CONCEIVE":
		// CONCEIVE phase: No inputs, only output hypothesis
		ctx.OutputFiles = []string{
			filepath.Join(strategyDir, "hypothesis.md"),
		}

	case "WRITE":
		// WRITE phase: Read hypothesis, write strategy code
		ctx.InputFiles = []string{
			filepath.Join(strategyDir, "hypothesis.md"),
		}
		ctx.OutputFiles = []string{
			filepath.Join(strategyDir, fmt.Sprintf("strategy_v%d.go", version)),
		}

	case "LINT":
		// LINT phase: Read strategy code, write validation log
		ctx.InputFiles = []string{
			filepath.Join(strategyDir, fmt.Sprintf("strategy_v%d.go", version)),
		}
		ctx.OutputFiles = []string{
			filepath.Join(strategyDir, "validation_errors.log"),
		}

	case "TEST":
		// TEST phase: Read strategy code, write test results
		ctx.InputFiles = []string{
			filepath.Join(strategyDir, fmt.Sprintf("strategy_v%d.go", version)),
		}
		ctx.OutputFiles = []string{
			filepath.Join(strategyDir, "validation_errors.log"),
		}

	case "BACKTEST":
		// BACKTEST phase: Read strategy code, write results
		ctx.InputFiles = []string{
			filepath.Join(strategyDir, fmt.Sprintf("strategy_v%d.go", version)),
		}
		ctx.OutputFiles = []string{
			filepath.Join(strategyDir, fmt.Sprintf("results_v%d.json", version)),
		}

	case "ANALYZE":
		// ANALYZE phase: Read hypothesis, results, memory, write analysis
		ctx.InputFiles = []string{
			filepath.Join(strategyDir, "hypothesis.md"),
			filepath.Join(strategyDir, fmt.Sprintf("results_v%d.json", version)),
			filepath.Join(strategyDir, "memory.json"),
		}
		ctx.OutputFiles = []string{
			filepath.Join(strategyDir, fmt.Sprintf("analysis_v%d.json", version)),
			filepath.Join(strategyDir, "memory.json"),
		}

	default:
		return nil, fmt.Errorf("unknown phase: %s", phase)
	}

	return ctx, nil
}

// ReadPhaseInputs reads all input files for a phase
func (cm *ContextManager) ReadPhaseInputs(ctx *PhaseContext) (map[string]string, error) {
	inputs := make(map[string]string)

	for _, filepath := range ctx.InputFiles {
		// Skip if file doesn't exist (optional input)
		if _, err := os.Stat(filepath); os.IsNotExist(err) {
			continue
		}

		data, err := os.ReadFile(filepath)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", filepath, err)
		}

		inputs[filepath] = string(data)
	}

	return inputs, nil
}

// ValidatePhaseOutputs checks if all required output files were created
func (cm *ContextManager) ValidatePhaseOutputs(ctx *PhaseContext) error {
	for _, filepath := range ctx.OutputFiles {
		if _, err := os.Stat(filepath); os.IsNotExist(err) {
			return fmt.Errorf("required output file not created: %s", filepath)
		}
	}
	return nil
}

// GenerateFocusedPrompt creates phase-specific prompt with only relevant context
func (cm *ContextManager) GenerateFocusedPrompt(ctx *PhaseContext, inputs map[string]string) (string, error) {
	prompt := ""

	switch ctx.Phase {
	case "CONCEIVE":
		prompt = cm.generateConceivePrompt()

	case "WRITE":
		hypothesis, ok := inputs[filepath.Join(cm.baseDir, ctx.StrategyID, "hypothesis.md")]
		if !ok {
			return "", fmt.Errorf("hypothesis.md not found in inputs")
		}
		prompt = cm.generateWritePrompt(hypothesis)

	case "ANALYZE":
		hypothesis := inputs[filepath.Join(cm.baseDir, ctx.StrategyID, "hypothesis.md")]
		results := inputs[filepath.Join(cm.baseDir, ctx.StrategyID, fmt.Sprintf("results_v%d.json", ctx.Version))]
		memory := inputs[filepath.Join(cm.baseDir, ctx.StrategyID, "memory.json")]
		prompt = cm.generateAnalyzePrompt(hypothesis, results, memory, ctx.Version)

	default:
		return "", fmt.Errorf("no prompt generator for phase: %s", ctx.Phase)
	}

	return prompt, nil
}

func (cm *ContextManager) generateConceivePrompt() string {
	return `You are a quantitative researcher. Formulate a trading hypothesis.

Context:
- Market: Cryptocurrency (BTCUSDT)
- Timeframe: 1 hour
- Available indicators: SMA, EMA, RSI, MACD, ATR, Bollinger Bands
- Risk management: Position sizing, stop loss, trailing stop

Task:
Write a hypothesis.md file following this structure:

# Hypothesis: [Strategy Name]

## Market Observation
- [Key observation 1]
- [Key observation 2]

## Core Thesis
[2-3 sentence explanation of your edge]

## Expected Edge
- Entry: [When to buy/sell]
- Exit: [When to close]
- Risk: [How to manage risk]

## Success Criteria
- Sharpe Ratio > 1.5
- Max Drawdown < 15%
- Win Rate > 45%

Output: hypothesis.md content only, no code blocks.`
}

func (cm *ContextManager) generateWritePrompt(hypothesis string) string {
	return fmt.Sprintf(`You are a Go developer implementing a trading strategy.

Hypothesis to implement:
---
%s
---

Context:
- Implement interface: Strategy (Init, OnBar methods)
- Use SDK methods: ctx.Buy(), ctx.Sell(), ctx.CloseAll()
- Use indicators: ctx.Indicator.SMA(), ctx.Indicator.RSI(), etc.

Constraints:
- No imports beyond SDK and indicators packages
- No goroutines, no channels
- Deterministic logic only
- Add // RATIONALE: comments explaining logic

Task:
Write complete strategy.go file implementing the hypothesis.

Output: Go code only, no markdown fences.`, hypothesis)
}

func (cm *ContextManager) generateAnalyzePrompt(hypothesis, results, memory string, version int) string {
	return fmt.Sprintf(`You are a quantitative analyst evaluating backtest results.

Hypothesis:
---
%s
---

Backtest Results (v%d):
---
%s
---

Previous Learning:
---
%s
---

Task:
1. Validate if hypothesis was confirmed or rejected
2. Identify specific weaknesses (e.g., "losses during trend", "late entries")
3. Propose ONE focused improvement for next iteration
4. Update memory.json with learnings

Output: analysis.json in this format:
{
  "strategy_id": "...",
  "version": %d,
  "hypothesis_validation": {
    "thesis_confirmed": true/false,
    "evidence": ["..."],
    "contradictions": ["..."]
  },
  "identified_weaknesses": ["..."],
  "next_iteration_plan": {
    "focus": "...",
    "rationale": "...",
    "expected_improvement": "..."
  }
}`, hypothesis, version, results, memory, version)
}

// SavePhaseContext saves context state for debugging
func (cm *ContextManager) SavePhaseContext(ctx *PhaseContext) error {
	strategyDir := filepath.Join(cm.baseDir, ctx.StrategyID)
	contextFile := filepath.Join(strategyDir, fmt.Sprintf("context_%s_v%d.json", ctx.Phase, ctx.Version))

	data, err := json.MarshalIndent(ctx, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(contextFile, data, 0644)
}

// LoadPhaseContext loads saved context state
func (cm *ContextManager) LoadPhaseContext(strategyID, phase string, version int) (*PhaseContext, error) {
	strategyDir := filepath.Join(cm.baseDir, strategyID)
	contextFile := filepath.Join(strategyDir, fmt.Sprintf("context_%s_v%d.json", phase, version))

	data, err := os.ReadFile(contextFile)
	if err != nil {
		return nil, err
	}

	var ctx PhaseContext
	if err := json.Unmarshal(data, &ctx); err != nil {
		return nil, err
	}

	return &ctx, nil
}

// CleanupPhaseContext removes temporary context files
func (cm *ContextManager) CleanupPhaseContext(strategyID string, version int) error {
	strategyDir := filepath.Join(cm.baseDir, strategyID)
	
	phases := []string{"CONCEIVE", "WRITE", "LINT", "TEST", "BACKTEST", "ANALYZE"}
	for _, phase := range phases {
		contextFile := filepath.Join(strategyDir, fmt.Sprintf("context_%s_v%d.json", phase, version))
		os.Remove(contextFile) // Ignore errors
	}
	
	return nil
}
