package validator

import (
	"strings"
	"testing"
)

func TestValidator_UnsafeImports(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{
			name: "safe imports",
			code: `package main
import "fmt"
func main() {}`,
			wantErr: false,
		},
		{
			name: "unsafe os import",
			code: `package main
import "os"
func main() {}`,
			wantErr: true,
		},
		{
			name: "unsafe net import",
			code: `package main
import "net"
func main() {}`,
			wantErr: true,
		},
		{
			name: "unsafe syscall import",
			code: `package main
import "syscall"
func main() {}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			err := v.ValidateFile("test.go", []byte(tt.code))
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidator_Goroutines(t *testing.T) {
	code := `package main
func main() {
	go func() {}
}`

	v := NewValidator()
	err := v.ValidateFile("test.go", []byte(code))
	if err == nil {
		t.Error("Expected error for goroutine usage")
	}

	errors := v.GetErrors()
	if len(errors) == 0 {
		t.Error("Expected validation errors")
	}

	if !strings.Contains(errors[0].Message, "goroutine") {
		t.Errorf("Expected goroutine error, got: %s", errors[0].Message)
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
