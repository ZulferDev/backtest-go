package risk

import (
	"fmt"
	"math"
)

// FixedFractional implements fixed fractional position sizing
type FixedFractional struct {
	RiskPercent float64
}

func NewFixedFractional(riskPercent float64) (*FixedFractional, error) {
	if riskPercent <= 0 || riskPercent > 1 {
		return nil, fmt.Errorf("riskPercent must be between 0 and 1")
	}
	return &FixedFractional{RiskPercent: riskPercent}, nil
}

func (f *FixedFractional) CalculateSize(equity, entryPrice, stopPrice float64) (float64, error) {
	if equity <= 0 || entryPrice <= 0 || stopPrice <= 0 {
		return 0, fmt.Errorf("all values must be positive")
	}
	riskAmount := equity * f.RiskPercent
	priceRisk := math.Abs(entryPrice - stopPrice)
	if priceRisk == 0 {
		return 0, fmt.Errorf("price risk cannot be zero")
	}
	return riskAmount / priceRisk, nil
}

// KellyCriterion implements Kelly Criterion position sizing
type KellyCriterion struct {
	WinRate       float64
	AvgWin        float64
	AvgLoss       float64
	Fraction      float64
	MaxAllocation float64
}

func NewKellyCriterion(winRate, avgWin, avgLoss, fraction, maxAllocation float64) (*KellyCriterion, error) {
	if winRate < 0 || winRate > 1 {
		return nil, fmt.Errorf("winRate must be between 0 and 1")
	}
	if avgWin <= 0 || avgLoss <= 0 {
		return nil, fmt.Errorf("avgWin and avgLoss must be positive")
	}
	if fraction <= 0 || fraction > 1 {
		return nil, fmt.Errorf("fraction must be between 0 and 1")
	}
	if maxAllocation <= 0 || maxAllocation > 1 {
		return nil, fmt.Errorf("maxAllocation must be between 0 and 1")
	}
	return &KellyCriterion{winRate, avgWin, avgLoss, fraction, maxAllocation}, nil
}

func (k *KellyCriterion) CalculateKellyPercent() float64 {
	lossRate := 1 - k.WinRate
	kellyPercent := (k.WinRate*k.AvgWin - lossRate*k.AvgLoss) / k.AvgWin
	kellyPercent *= k.Fraction
	if kellyPercent > k.MaxAllocation {
		kellyPercent = k.MaxAllocation
	}
	if kellyPercent < 0 {
		kellyPercent = 0
	}
	return kellyPercent
}

func (k *KellyCriterion) CalculateSize(equity, price float64) (float64, error) {
	if equity <= 0 || price <= 0 {
		return 0, fmt.Errorf("equity and price must be positive")
	}
	kellyPercent := k.CalculateKellyPercent()
	return (equity * kellyPercent) / price, nil
}
