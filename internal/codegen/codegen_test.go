package codegen

import (
	"strings"
	"testing"
)

func TestSystemPrompt(t *testing.T) {
	prompt := SystemPrompt()
	if !strings.Contains(prompt, "sdk.Strategy") {
		t.Error("Should mention sdk.Strategy")
	}
}

func TestBuildPrompt(t *testing.T) {
	spec := StrategySpec{Name: "Test", Description: "A test"}
	prompt := BuildPrompt(spec)
	if !strings.Contains(prompt, "Test") {
		t.Error("Should contain strategy name")
	}
}

func TestPipelineValidate(t *testing.T) {
	p := NewPipeline(t.TempDir())
	code := `package strategies
import "github.com/ZulferDev/backtest-go/pkg/sdk"
type T struct{}
func (t *T) Init(ctx sdk.InitContext) error { return nil }
func (t *T) OnBar(ctx sdk.BarContext, bar sdk.OHLCV) error { return nil }`
	errs, err := p.Validate(code, "test.go")
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	if len(errs) > 0 {
		t.Errorf("Should have no errors: %v", errs)
	}
}
