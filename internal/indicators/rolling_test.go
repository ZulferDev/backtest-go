package indicators

import (
	"math"
	"testing"
)

func TestRollingMin(t *testing.T) {
	data := []float64{5.0, 3.0, 8.0, 2.0, 7.0, 1.0, 6.0}
	window := 3
	
	rw, err := NewRollingWindow(data, window)
	if err != nil {
		t.Fatalf("Failed to create rolling window: %v", err)
	}
	
	result, err := rw.RollingMin()
	if err != nil {
		t.Fatalf("RollingMin failed: %v", err)
	}
	
	expected := []float64{3.0, 2.0, 2.0, 1.0, 1.0}
	
	if len(result) != len(expected) {
		t.Fatalf("Expected length %d, got %d", len(expected), len(result))
	}
	
	for i, v := range result {
		if v != expected[i] {
			t.Errorf("Index %d: expected %.1f, got %.1f", i, expected[i], v)
		}
	}
}

func TestRollingMax(t *testing.T) {
	data := []float64{5.0, 3.0, 8.0, 2.0, 7.0, 1.0, 6.0}
	window := 3
	
	rw, err := NewRollingWindow(data, window)
	if err != nil {
		t.Fatalf("Failed to create rolling window: %v", err)
	}
	
	result, err := rw.RollingMax()
	if err != nil {
		t.Fatalf("RollingMax failed: %v", err)
	}
	
	expected := []float64{8.0, 8.0, 8.0, 7.0, 7.0}
	
	if len(result) != len(expected) {
		t.Fatalf("Expected length %d, got %d", len(expected), len(result))
	}
	
	for i, v := range result {
		if v != expected[i] {
			t.Errorf("Index %d: expected %.1f, got %.1f", i, expected[i], v)
		}
	}
}

func TestRollingMean(t *testing.T) {
	data := []float64{2.0, 4.0, 6.0, 8.0, 10.0}
	window := 3
	
	rw, err := NewRollingWindow(data, window)
	if err != nil {
		t.Fatalf("Failed to create rolling window: %v", err)
	}
	
	result, err := rw.RollingMean()
	if err != nil {
		t.Fatalf("RollingMean failed: %v", err)
	}
	
	expected := []float64{4.0, 6.0, 8.0}
	
	if len(result) != len(expected) {
		t.Fatalf("Expected length %d, got %d", len(expected), len(result))
	}
	
	for i, v := range result {
		if math.Abs(v-expected[i]) > 0.0001 {
			t.Errorf("Index %d: expected %.1f, got %.1f", i, expected[i], v)
		}
	}
}

func TestRollingStdDev(t *testing.T) {
	data := []float64{2.0, 4.0, 4.0, 4.0, 5.0, 5.0, 7.0, 9.0}
	window := 4
	
	rw, err := NewRollingWindow(data, window)
	if err != nil {
		t.Fatalf("Failed to create rolling window: %v", err)
	}
	
	result, err := rw.RollingStdDev()
	if err != nil {
		t.Fatalf("RollingStdDev failed: %v", err)
	}
	
	if len(result) != len(data)-window+1 {
		t.Fatalf("Expected length %d, got %d", len(data)-window+1, len(result))
	}
	
	// For window [2, 4, 4, 4]: mean=3.5, variance=0.75, stddev≈0.866
	expectedFirst := 0.866
	if math.Abs(result[0]-expectedFirst) > 0.01 {
		t.Errorf("First value: expected ~%.3f, got %.3f", expectedFirst, result[0])
	}
	
	t.Logf("Rolling StdDev results: %v", result)
}

func TestRollingMedian(t *testing.T) {
	data := []float64{1.0, 3.0, 5.0, 7.0, 9.0}
	window := 3
	
	rw, err := NewRollingWindow(data, window)
	if err != nil {
		t.Fatalf("Failed to create rolling window: %v", err)
	}
	
	result, err := rw.RollingMedian()
	if err != nil {
		t.Fatalf("RollingMedian failed: %v", err)
	}
	
	expected := []float64{3.0, 5.0, 7.0}
	
	if len(result) != len(expected) {
		t.Fatalf("Expected length %d, got %d", len(expected), len(result))
	}
	
	for i, v := range result {
		if v != expected[i] {
			t.Errorf("Index %d: expected %.1f, got %.1f", i, expected[i], v)
		}
	}
}

func TestRollingSum(t *testing.T) {
	data := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	window := 3
	
	rw, err := NewRollingWindow(data, window)
	if err != nil {
		t.Fatalf("Failed to create rolling window: %v", err)
	}
	
	result, err := rw.RollingSum()
	if err != nil {
		t.Fatalf("RollingSum failed: %v", err)
	}
	
	expected := []float64{6.0, 9.0, 12.0}
	
	if len(result) != len(expected) {
		t.Fatalf("Expected length %d, got %d", len(expected), len(result))
	}
	
	for i, v := range result {
		if v != expected[i] {
			t.Errorf("Index %d: expected %.1f, got %.1f", i, expected[i], v)
		}
	}
}

func TestRollingVar(t *testing.T) {
	data := []float64{2.0, 4.0, 4.0, 4.0, 5.0}
	window := 3
	
	rw, err := NewRollingWindow(data, window)
	if err != nil {
		t.Fatalf("Failed to create rolling window: %v", err)
	}
	
	result, err := rw.RollingVar()
	if err != nil {
		t.Fatalf("RollingVar failed: %v", err)
	}
	
	if len(result) != 3 {
		t.Fatalf("Expected length 3, got %d", len(result))
	}
	
	t.Logf("Rolling Variance results: %v", result)
}

func TestRollingWindowErrors(t *testing.T) {
	data := []float64{1.0, 2.0, 3.0}
	
	// Test zero window
	_, err := NewRollingWindow(data, 0)
	if err == nil {
		t.Error("Expected error for zero window")
	}
	
	// Test negative window
	_, err = NewRollingWindow(data, -1)
	if err == nil {
		t.Error("Expected error for negative window")
	}
	
	// Test window larger than data
	_, err = NewRollingWindow(data, 10)
	if err == nil {
		t.Error("Expected error for window larger than data")
	}
}

func TestRollingSimpleFunctions(t *testing.T) {
	data := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	window := 3
	
	// Test RollingMinSimple
	mins, err := RollingMinSimple(data, window)
	if err != nil {
		t.Errorf("RollingMinSimple failed: %v", err)
	}
	if len(mins) != 3 {
		t.Errorf("Expected 3 results, got %d", len(mins))
	}
	
	// Test RollingMaxSimple
	maxs, err := RollingMaxSimple(data, window)
	if err != nil {
		t.Errorf("RollingMaxSimple failed: %v", err)
	}
	if len(maxs) != 3 {
		t.Errorf("Expected 3 results, got %d", len(maxs))
	}
	
	// Test RollingStdDevSimple
	stds, err := RollingStdDevSimple(data, window)
	if err != nil {
		t.Errorf("RollingStdDevSimple failed: %v", err)
	}
	if len(stds) != 3 {
		t.Errorf("Expected 3 results, got %d", len(stds))
	}
	
	// Test RollingMeanSimple
	means, err := RollingMeanSimple(data, window)
	if err != nil {
		t.Errorf("RollingMeanSimple failed: %v", err)
	}
	if len(means) != 3 {
		t.Errorf("Expected 3 results, got %d", len(means))
	}
}

func TestRollingWithPriceData(t *testing.T) {
	// Simulate price data
	prices := []float64{50000, 50100, 49900, 50200, 50300, 49800, 50400, 50500}
	window := 5
	
	rw, err := NewRollingWindow(prices, window)
	if err != nil {
		t.Fatalf("Failed to create rolling window: %v", err)
	}
	
	// Test rolling min (support level)
	mins, err := rw.RollingMin()
	if err != nil {
		t.Fatalf("RollingMin failed: %v", err)
	}
	t.Logf("Rolling Min (Support): %v", mins)
	
	// Test rolling max (resistance level)
	maxs, err := rw.RollingMax()
	if err != nil {
		t.Fatalf("RollingMax failed: %v", err)
	}
	t.Logf("Rolling Max (Resistance): %v", maxs)
	
	// Test rolling stddev (volatility)
	stds, err := rw.RollingStdDev()
	if err != nil {
		t.Fatalf("RollingStdDev failed: %v", err)
	}
	t.Logf("Rolling StdDev (Volatility): %v", stds)
	
	// Verify we have correct number of results
	expectedLen := len(prices) - window + 1
	if len(mins) != expectedLen || len(maxs) != expectedLen || len(stds) != expectedLen {
		t.Error("Inconsistent result lengths")
	}
}
