package context

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewContextManager(t *testing.T) {
	cm := NewContextManager("/tmp/research_logs")
	if cm == nil {
		t.Fatal("ContextManager should not be nil")
	}
	if cm.baseDir != "/tmp/research_logs" {
		t.Errorf("Expected baseDir '/tmp/research_logs', got '%s'", cm.baseDir)
	}
}

func TestGetPhaseContext(t *testing.T) {
	cm := NewContextManager("/tmp/research_logs")

	tests := []struct {
		name          string
		phase         string
		expectedInput int
		expectedOutput int
	}{
		{"CONCEIVE phase", "CONCEIVE", 0, 1},
		{"WRITE phase", "WRITE", 1, 1},
		{"LINT phase", "LINT", 1, 1},
		{"TEST phase", "TEST", 1, 1},
		{"BACKTEST phase", "BACKTEST", 1, 1},
		{"ANALYZE phase", "ANALYZE", 3, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, err := cm.GetPhaseContext("test_strategy", tt.phase, 1)
			if err != nil {
				t.Fatalf("GetPhaseContext failed: %v", err)
			}

			if ctx.Phase != tt.phase {
				t.Errorf("Expected phase '%s', got '%s'", tt.phase, ctx.Phase)
			}

			if len(ctx.InputFiles) != tt.expectedInput {
				t.Errorf("Expected %d input files, got %d", tt.expectedInput, len(ctx.InputFiles))
			}

			if len(ctx.OutputFiles) != tt.expectedOutput {
				t.Errorf("Expected %d output files, got %d", tt.expectedOutput, len(ctx.OutputFiles))
			}
		})
	}
}

func TestGetPhaseContextInvalidPhase(t *testing.T) {
	cm := NewContextManager("/tmp/research_logs")
	_, err := cm.GetPhaseContext("test_strategy", "INVALID_PHASE", 1)
	if err == nil {
		t.Error("Expected error for invalid phase")
	}
}

func TestReadPhaseInputs(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	cm := NewContextManager(tmpDir)

	strategyID := "test_strategy"
	strategyDir := filepath.Join(tmpDir, strategyID)
	os.MkdirAll(strategyDir, 0755)

	// Create test input file
	hypothesisFile := filepath.Join(strategyDir, "hypothesis.md")
	testContent := "# Test Hypothesis\n\nThis is a test."
	os.WriteFile(hypothesisFile, []byte(testContent), 0644)

	// Get WRITE phase context (requires hypothesis.md)
	ctx, err := cm.GetPhaseContext(strategyID, "WRITE", 1)
	if err != nil {
		t.Fatalf("GetPhaseContext failed: %v", err)
	}

	// Read inputs
	inputs, err := cm.ReadPhaseInputs(ctx)
	if err != nil {
		t.Fatalf("ReadPhaseInputs failed: %v", err)
	}

	if len(inputs) != 1 {
		t.Fatalf("Expected 1 input, got %d", len(inputs))
	}

	content, ok := inputs[hypothesisFile]
	if !ok {
		t.Errorf("Expected input file '%s' not found", hypothesisFile)
	}

	if content != testContent {
		t.Errorf("Content mismatch. Expected '%s', got '%s'", testContent, content)
	}
}

func TestReadPhaseInputsOptionalFiles(t *testing.T) {
	// Create temporary directory without input files
	tmpDir := t.TempDir()
	cm := NewContextManager(tmpDir)

	strategyID := "test_strategy"
	os.MkdirAll(filepath.Join(tmpDir, strategyID), 0755)

	ctx, _ := cm.GetPhaseContext(strategyID, "WRITE", 1)

	// Should not error even if files don't exist
	inputs, err := cm.ReadPhaseInputs(ctx)
	if err != nil {
		t.Errorf("ReadPhaseInputs should not error on missing optional files: %v", err)
	}

	if len(inputs) != 0 {
		t.Errorf("Expected 0 inputs when files don't exist, got %d", len(inputs))
	}
}

func TestValidatePhaseOutputs(t *testing.T) {
	tmpDir := t.TempDir()
	cm := NewContextManager(tmpDir)

	strategyID := "test_strategy"
	strategyDir := filepath.Join(tmpDir, strategyID)
	os.MkdirAll(strategyDir, 0755)

	ctx, _ := cm.GetPhaseContext(strategyID, "CONCEIVE", 1)

	// Validate without creating output - should fail
	err := cm.ValidatePhaseOutputs(ctx)
	if err == nil {
		t.Error("Expected error when output files don't exist")
	}

	// Create required output file
	hypothesisFile := filepath.Join(strategyDir, "hypothesis.md")
	os.WriteFile(hypothesisFile, []byte("test"), 0644)

	// Validate again - should succeed
	err = cm.ValidatePhaseOutputs(ctx)
	if err != nil {
		t.Errorf("ValidatePhaseOutputs failed after creating output: %v", err)
	}
}

func TestGenerateFocusedPrompt(t *testing.T) {
	tmpDir := t.TempDir()
	cm := NewContextManager(tmpDir)

	t.Run("CONCEIVE prompt", func(t *testing.T) {
		ctx, _ := cm.GetPhaseContext("test_strategy", "CONCEIVE", 1)
		inputs := make(map[string]string)

		prompt, err := cm.GenerateFocusedPrompt(ctx, inputs)
		if err != nil {
			t.Fatalf("GenerateFocusedPrompt failed: %v", err)
		}

		if len(prompt) == 0 {
			t.Error("Prompt should not be empty")
		}

		if !contains(prompt, "quantitative researcher") {
			t.Error("CONCEIVE prompt should mention 'quantitative researcher'")
		}
	})

	t.Run("WRITE prompt", func(t *testing.T) {
		strategyID := "test_strategy"
		strategyDir := filepath.Join(tmpDir, strategyID)
		os.MkdirAll(strategyDir, 0755)

		hypothesisFile := filepath.Join(strategyDir, "hypothesis.md")
		hypothesisContent := "# Test Hypothesis\n\nRSI mean reversion strategy."
		os.WriteFile(hypothesisFile, []byte(hypothesisContent), 0644)

		ctx, _ := cm.GetPhaseContext(strategyID, "WRITE", 1)
		inputs := map[string]string{
			hypothesisFile: hypothesisContent,
		}

		prompt, err := cm.GenerateFocusedPrompt(ctx, inputs)
		if err != nil {
			t.Fatalf("GenerateFocusedPrompt failed: %v", err)
		}

		if !contains(prompt, "RSI mean reversion") {
			t.Error("WRITE prompt should include hypothesis content")
		}

		if !contains(prompt, "Go developer") {
			t.Error("WRITE prompt should mention 'Go developer'")
		}
	})

	t.Run("ANALYZE prompt", func(t *testing.T) {
		strategyID := "test_strategy"
		strategyDir := filepath.Join(tmpDir, strategyID)
		os.MkdirAll(strategyDir, 0755)

		hypothesisFile := filepath.Join(strategyDir, "hypothesis.md")
		resultsFile := filepath.Join(strategyDir, "results_v1.json")
		memoryFile := filepath.Join(strategyDir, "memory.json")

		os.WriteFile(hypothesisFile, []byte("# Hypothesis"), 0644)
		os.WriteFile(resultsFile, []byte(`{"sharpe_ratio": 1.5}`), 0644)
		os.WriteFile(memoryFile, []byte(`{"insights": []}`), 0644)

		ctx, _ := cm.GetPhaseContext(strategyID, "ANALYZE", 1)
		inputs := map[string]string{
			hypothesisFile: "# Hypothesis",
			resultsFile:    `{"sharpe_ratio": 1.5}`,
			memoryFile:     `{"insights": []}`,
		}

		prompt, err := cm.GenerateFocusedPrompt(ctx, inputs)
		if err != nil {
			t.Fatalf("GenerateFocusedPrompt failed: %v", err)
		}

		if !contains(prompt, "quantitative analyst") {
			t.Error("ANALYZE prompt should mention 'quantitative analyst'")
		}

		if !contains(prompt, "sharpe_ratio") {
			t.Error("ANALYZE prompt should include results content")
		}
	})
}

func TestSaveAndLoadPhaseContext(t *testing.T) {
	tmpDir := t.TempDir()
	cm := NewContextManager(tmpDir)

	strategyID := "test_strategy"
	os.MkdirAll(filepath.Join(tmpDir, strategyID), 0755)

	// Create and save context
	ctx, _ := cm.GetPhaseContext(strategyID, "WRITE", 2)
	ctx.Metadata["test_key"] = "test_value"

	err := cm.SavePhaseContext(ctx)
	if err != nil {
		t.Fatalf("SavePhaseContext failed: %v", err)
	}

	// Load context
	loadedCtx, err := cm.LoadPhaseContext(strategyID, "WRITE", 2)
	if err != nil {
		t.Fatalf("LoadPhaseContext failed: %v", err)
	}

	if loadedCtx.StrategyID != strategyID {
		t.Errorf("Expected strategy_id '%s', got '%s'", strategyID, loadedCtx.StrategyID)
	}

	if loadedCtx.Phase != "WRITE" {
		t.Errorf("Expected phase 'WRITE', got '%s'", loadedCtx.Phase)
	}

	if loadedCtx.Version != 2 {
		t.Errorf("Expected version 2, got %d", loadedCtx.Version)
	}

	if loadedCtx.Metadata["test_key"] != "test_value" {
		t.Error("Metadata not preserved after load")
	}
}

func TestCleanupPhaseContext(t *testing.T) {
	tmpDir := t.TempDir()
	cm := NewContextManager(tmpDir)

	strategyID := "test_strategy"
	strategyDir := filepath.Join(tmpDir, strategyID)
	os.MkdirAll(strategyDir, 0755)

	// Create multiple phase contexts
	phases := []string{"CONCEIVE", "WRITE", "LINT"}
	for _, phase := range phases {
		ctx, _ := cm.GetPhaseContext(strategyID, phase, 1)
		cm.SavePhaseContext(ctx)
	}

	// Cleanup
	err := cm.CleanupPhaseContext(strategyID, 1)
	if err != nil {
		t.Fatalf("CleanupPhaseContext failed: %v", err)
	}

	// Verify files are removed
	for _, phase := range phases {
		contextFile := filepath.Join(strategyDir, "context_"+phase+"_v1.json")
		if _, err := os.Stat(contextFile); !os.IsNotExist(err) {
			t.Errorf("Context file %s should have been removed", contextFile)
		}
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr)))
}
