package validator

import (
	"strings"
	"testing"
)

func TestValidator_UnsafeImports(t *testing.T) {
	tests := []struct {
		name        string
		code        string
		wantErrors  bool
	}{
		{
			name: "safe imports",
			code: `package main
import "fmt"
func main() {}`,
			wantErrors: false,
		},
		{
			name: "unsafe os import",
			code: `package main
import "os"
func main() {}`,
			wantErrors: true,
		},
		{
			name: "unsafe net import",
			code: `package main
import "net"
func main() {}`,
			wantErrors: true,
		},
		{
			name: "unsafe syscall import",
			code: `package main
import "syscall"
func main() {}`,
			wantErrors: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			err := v.ValidateFile("test.go", []byte(tt.code))
			if err != nil {
				t.Fatalf("ValidateFile() unexpected error: %v", err)
			}
			
			hasErrors := len(v.GetErrors()) > 0
			if hasErrors != tt.wantErrors {
				t.Errorf("Expected validation errors=%v, got errors=%v (%v)", tt.wantErrors, hasErrors, v.GetErrors())
			}
		})
	}
}

func TestValidator_Goroutines(t *testing.T) {
	code := `package main

func main() {
	go func() {}()
}
`

	v := NewValidator()
	err := v.ValidateFile("test.go", []byte(code))
	
	if err != nil {
		t.Fatalf("ValidateFile() unexpected error: %v", err)
	}
	
	// Should have validation errors
	errors := v.GetErrors()
	t.Logf("ValidateFile returned err: %v, errors count: %d", err, len(errors))
	
	if len(errors) == 0 {
		t.Fatal("Expected at least one validation error")
	}

	// Check if any error mentions goroutines
	found := false
	for _, e := range errors {
		t.Logf("Error: %s (rule: %s)", e.Message, e.Rule)
		if strings.Contains(e.Message, "goroutine") {
			found = true
		}
	}
	
	if !found {
		t.Errorf("No goroutine error found. Errors: %+v", errors)
	}
}

func TestValidator_SafeCode(t *testing.T) {
	code := `package strategies

import (
	"github.com/ZulferDev/backtest-go/pkg/sdk"
	"github.com/ZulferDev/backtest-go/internal/indicators"
)

type MyStrategy struct {}

func (s *MyStrategy) Init(ctx sdk.InitContext) error {
	return nil
}

func (s *MyStrategy) OnBar(ctx sdk.BarContext, bar sdk.OHLCV) error {
	return nil
}`

	v := NewValidator()
	err := v.ValidateFile("strategy.go", []byte(code))
	if err != nil {
		t.Errorf("Safe code should pass validation, got: %v", err)
	}
}

func TestExtractStrategyInfo(t *testing.T) {
	code := `package strategies

type SMACrossover struct {
	fastPeriod int
	slowPeriod int
}`

	info, err := ExtractStrategyInfo("strategy.go", []byte(code))
	if err != nil {
		t.Fatalf("ExtractStrategyInfo() failed: %v", err)
	}

	if info.PackageName != "strategies" {
		t.Errorf("PackageName = %v, want strategies", info.PackageName)
	}

	if info.StrategyName != "SMACrossover" {
		t.Errorf("StrategyName = %v, want SMACrossover", info.StrategyName)
	}
}

func TestGenerateTestFile(t *testing.T) {
	info := &StrategyInfo{
		PackageName:  "strategies",
		StrategyName: "TestStrategy",
		Filename:     "test.go",
	}

	testCode := GenerateTestFile(info)

	// Check generated code contains expected elements
	if !strings.Contains(testCode, "package strategies") {
		t.Error("Generated test should have correct package")
	}

	if !strings.Contains(testCode, "func TestTestStrategy_Init") {
		t.Error("Generated test should have Init test")
	}

	if !strings.Contains(testCode, "func TestTestStrategy_OnBar") {
		t.Error("Generated test should have OnBar test")
	}

	if !strings.Contains(testCode, "mockBarContext") {
		t.Error("Generated test should have mock context")
	}
}
