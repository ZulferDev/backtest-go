package codegen

import "fmt"

type StrategySpec struct {
	Name        string
	Description string
	Timeframe   string
}

func SystemPrompt() string {
	return `You are an expert Go trading strategy developer.

Write a complete strategy implementing sdk.Strategy interface.

RULES:
- ONLY safe imports: github.com/ZulferDev/backtest-go/internal/{indicators,risk,signal}
- NO: os, net, syscall, unsafe, goroutines
- Must be deterministic

Write ONLY Go code.`
}

func BuildPrompt(spec StrategySpec) string {
	return SystemPrompt() + fmt.Sprintf("\n\nStrategy: %s\nDescription: %s\n", spec.Name, spec.Description)
}
