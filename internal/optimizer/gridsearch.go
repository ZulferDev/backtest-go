package optimizer

import (
	"fmt"
	"math"
)

// ParameterRange defines a parameter search space
type ParameterRange struct {
	Name   string
	Type   string      // "int", "float", "bool", "string"
	Min    float64     // For int/float types
	Max    float64     // For int/float types
	Step   float64     // For int/float types
	Values []interface{} // For discrete values (bool, string, or specific values)
}

// ParameterSet represents a specific combination of parameters
type ParameterSet map[string]interface{}

// GridSearch generates all parameter combinations
type GridSearch struct {
	ranges []ParameterRange
}

// NewGridSearch creates a new grid search
func NewGridSearch(ranges []ParameterRange) *GridSearch {
	return &GridSearch{
		ranges: ranges,
	}
}

// Generate creates all parameter combinations
func (g *GridSearch) Generate() ([]ParameterSet, error) {
	if len(g.ranges) == 0 {
		return nil, fmt.Errorf("no parameter ranges defined")
	}

	// Validate ranges
	for _, r := range g.ranges {
		if err := g.validateRange(r); err != nil {
			return nil, fmt.Errorf("invalid range %s: %w", r.Name, err)
		}
	}

	// Generate all combinations
	var combinations []ParameterSet
	current := make(ParameterSet)
	g.generateRecursive(0, current, &combinations)

	return combinations, nil
}

// validateRange checks if a parameter range is valid
func (g *GridSearch) validateRange(r ParameterRange) error {
	switch r.Type {
	case "int", "float":
		if r.Min > r.Max {
			return fmt.Errorf("min (%v) cannot be greater than max (%v)", r.Min, r.Max)
		}
		if r.Step <= 0 {
			return fmt.Errorf("step must be positive, got %v", r.Step)
		}
	case "bool", "string":
		if len(r.Values) == 0 {
			return fmt.Errorf("values cannot be empty for type %s", r.Type)
		}
	default:
		return fmt.Errorf("unsupported type: %s", r.Type)
	}
	return nil
}

// generateRecursive recursively generates all combinations
func (g *GridSearch) generateRecursive(rangeIdx int, current ParameterSet, combinations *[]ParameterSet) {
	if rangeIdx == len(g.ranges) {
		// Make a copy of current combination
		combo := make(ParameterSet)
		for k, v := range current {
			combo[k] = v
		}
		*combinations = append(*combinations, combo)
		return
	}

	r := g.ranges[rangeIdx]
	values := g.getValuesForRange(r)

	for _, val := range values {
		current[r.Name] = val
		g.generateRecursive(rangeIdx+1, current, combinations)
	}
}

// getValuesForRange returns all values for a parameter range
func (g *GridSearch) getValuesForRange(r ParameterRange) []interface{} {
	switch r.Type {
	case "int":
		return g.generateIntRange(r)
	case "float":
		return g.generateFloatRange(r)
	case "bool", "string":
		return r.Values
	default:
		return []interface{}{}
	}
}

// generateIntRange generates integer values
func (g *GridSearch) generateIntRange(r ParameterRange) []interface{} {
	var values []interface{}
	for val := r.Min; val <= r.Max; val += r.Step {
		values = append(values, int(val))
	}
	return values
}

// generateFloatRange generates float values
func (g *GridSearch) generateFloatRange(r ParameterRange) []interface{} {
	var values []interface{}
	steps := int(math.Round((r.Max - r.Min) / r.Step))
	for i := 0; i <= steps; i++ {
		val := r.Min + (float64(i) * r.Step)
		if val <= r.Max {
			values = append(values, val)
		}
	}
	return values
}

// EstimateSize calculates total number of combinations
func (g *GridSearch) EstimateSize() int {
	if len(g.ranges) == 0 {
		return 0
	}

	size := 1
	for _, r := range g.ranges {
		count := g.countValues(r)
		size *= count
	}
	return size
}

// countValues counts number of values in a range
func (g *GridSearch) countValues(r ParameterRange) int {
	switch r.Type {
	case "int":
		return int((r.Max-r.Min)/r.Step) + 1
	case "float":
		return int(math.Round((r.Max-r.Min)/r.Step)) + 1
	case "bool", "string":
		return len(r.Values)
	default:
		return 0
	}
}

// RandomSearch generates random parameter combinations
type RandomSearch struct {
	ranges      []ParameterRange
	numSamples  int
	seed        int64
	currentIdx  int
}

// NewRandomSearch creates a new random search
func NewRandomSearch(ranges []ParameterRange, numSamples int) *RandomSearch {
	return &RandomSearch{
		ranges:     ranges,
		numSamples: numSamples,
	}
}

// Generate creates random parameter combinations
func (r *RandomSearch) Generate() ([]ParameterSet, error) {
	if len(r.ranges) == 0 {
		return nil, fmt.Errorf("no parameter ranges defined")
	}

	// For now, return a subset of grid search
	// Full random implementation would use actual randomization
	grid := NewGridSearch(r.ranges)
	allCombos, err := grid.Generate()
	if err != nil {
		return nil, err
	}

	// Take first N samples (simplified - would use random sampling in production)
	if r.numSamples > len(allCombos) {
		return allCombos, nil
	}

	return allCombos[:r.numSamples], nil
}
