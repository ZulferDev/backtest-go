package indicators

import (
	"fmt"
	"math"
)

// RollingWindow provides rolling window calculations for time series data
type RollingWindow struct {
	data   []float64
	window int
}

// NewRollingWindow creates a new rolling window calculator
func NewRollingWindow(data []float64, window int) (*RollingWindow, error) {
	if window <= 0 {
		return nil, fmt.Errorf("window must be positive")
	}
	if window > len(data) {
		return nil, fmt.Errorf("window (%d) cannot be larger than data length (%d)", window, len(data))
	}
	
	return &RollingWindow{
		data:   data,
		window: window,
	}, nil
}

// RollingMin calculates rolling minimum values
func (r *RollingWindow) RollingMin() ([]float64, error) {
	result := make([]float64, 0, len(r.data)-r.window+1)
	
	for i := r.window - 1; i < len(r.data); i++ {
		windowData := r.data[i-r.window+1 : i+1]
		min := windowData[0]
		
		for _, v := range windowData[1:] {
			if v < min {
				min = v
			}
		}
		
		result = append(result, min)
	}
	
	return result, nil
}

// RollingMax calculates rolling maximum values
func (r *RollingWindow) RollingMax() ([]float64, error) {
	result := make([]float64, 0, len(r.data)-r.window+1)
	
	for i := r.window - 1; i < len(r.data); i++ {
		windowData := r.data[i-r.window+1 : i+1]
		max := windowData[0]
		
		for _, v := range windowData[1:] {
			if v > max {
				max = v
			}
		}
		
		result = append(result, max)
	}
	
	return result, nil
}

// RollingStdDev calculates rolling standard deviation
func (r *RollingWindow) RollingStdDev() ([]float64, error) {
	result := make([]float64, 0, len(r.data)-r.window+1)
	
	for i := r.window - 1; i < len(r.data); i++ {
		windowData := r.data[i-r.window+1 : i+1]
		
		// Calculate mean
		sum := 0.0
		for _, v := range windowData {
			sum += v
		}
		mean := sum / float64(len(windowData))
		
		// Calculate variance
		variance := 0.0
		for _, v := range windowData {
			diff := v - mean
			variance += diff * diff
		}
		variance /= float64(len(windowData))
		
		// Standard deviation
		stdDev := math.Sqrt(variance)
		result = append(result, stdDev)
	}
	
	return result, nil
}

// RollingMean calculates rolling mean (average)
func (r *RollingWindow) RollingMean() ([]float64, error) {
	result := make([]float64, 0, len(r.data)-r.window+1)
	
	for i := r.window - 1; i < len(r.data); i++ {
		windowData := r.data[i-r.window+1 : i+1]
		
		sum := 0.0
		for _, v := range windowData {
			sum += v
		}
		mean := sum / float64(len(windowData))
		
		result = append(result, mean)
	}
	
	return result, nil
}

// RollingMedian calculates rolling median
func (r *RollingWindow) RollingMedian() ([]float64, error) {
	result := make([]float64, 0, len(r.data)-r.window+1)
	
	for i := r.window - 1; i < len(r.data); i++ {
		windowData := make([]float64, r.window)
		copy(windowData, r.data[i-r.window+1:i+1])
		
		// Sort to find median
		sortFloat64(windowData)
		
		var median float64
		mid := len(windowData) / 2
		if len(windowData)%2 == 0 {
			median = (windowData[mid-1] + windowData[mid]) / 2.0
		} else {
			median = windowData[mid]
		}
		
		result = append(result, median)
	}
	
	return result, nil
}

// RollingSum calculates rolling sum
func (r *RollingWindow) RollingSum() ([]float64, error) {
	result := make([]float64, 0, len(r.data)-r.window+1)
	
	for i := r.window - 1; i < len(r.data); i++ {
		windowData := r.data[i-r.window+1 : i+1]
		
		sum := 0.0
		for _, v := range windowData {
			sum += v
		}
		
		result = append(result, sum)
	}
	
	return result, nil
}

// RollingVar calculates rolling variance
func (r *RollingWindow) RollingVar() ([]float64, error) {
	result := make([]float64, 0, len(r.data)-r.window+1)
	
	for i := r.window - 1; i < len(r.data); i++ {
		windowData := r.data[i-r.window+1 : i+1]
		
		// Calculate mean
		sum := 0.0
		for _, v := range windowData {
			sum += v
		}
		mean := sum / float64(len(windowData))
		
		// Calculate variance
		variance := 0.0
		for _, v := range windowData {
			diff := v - mean
			variance += diff * diff
		}
		variance /= float64(len(windowData))
		
		result = append(result, variance)
	}
	
	return result, nil
}

// sortFloat64 sorts a slice of float64 in ascending order (simple bubble sort for small windows)
func sortFloat64(data []float64) {
	n := len(data)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if data[j] > data[j+1] {
				data[j], data[j+1] = data[j+1], data[j]
			}
		}
	}
}

// Helper functions for convenience

// RollingMinSimple is a convenience function for calculating rolling minimum
func RollingMinSimple(data []float64, window int) ([]float64, error) {
	rw, err := NewRollingWindow(data, window)
	if err != nil {
		return nil, err
	}
	return rw.RollingMin()
}

// RollingMaxSimple is a convenience function for calculating rolling maximum
func RollingMaxSimple(data []float64, window int) ([]float64, error) {
	rw, err := NewRollingWindow(data, window)
	if err != nil {
		return nil, err
	}
	return rw.RollingMax()
}

// RollingStdDevSimple is a convenience function for calculating rolling standard deviation
func RollingStdDevSimple(data []float64, window int) ([]float64, error) {
	rw, err := NewRollingWindow(data, window)
	if err != nil {
		return nil, err
	}
	return rw.RollingStdDev()
}

// RollingMeanSimple is a convenience function for calculating rolling mean
func RollingMeanSimple(data []float64, window int) ([]float64, error) {
	rw, err := NewRollingWindow(data, window)
	if err != nil {
		return nil, err
	}
	return rw.RollingMean()
}
