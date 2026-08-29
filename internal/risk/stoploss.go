package risk

import (
	"fmt"
	"math"
)

type StopLoss struct {
	Price float64
	Type  string
}

func FixedStopLoss(stopPrice float64) (*StopLoss, error) {
	if stopPrice <= 0 {
		return nil, fmt.Errorf("stop price must be positive")
	}
	return &StopLoss{Price: stopPrice, Type: "fixed"}, nil
}

func PercentStopLoss(entryPrice, percent float64, side string) (*StopLoss, error) {
	if entryPrice <= 0 || percent <= 0 || percent >= 1 {
		return nil, fmt.Errorf("invalid parameters")
	}
	if side != "long" && side != "short" {
		return nil, fmt.Errorf("side must be 'long' or 'short'")
	}
	var stopPrice float64
	if side == "long" {
		stopPrice = entryPrice * (1 - percent)
	} else {
		stopPrice = entryPrice * (1 + percent)
	}
	return &StopLoss{Price: stopPrice, Type: "percent"}, nil
}

func ATRStopLoss(entryPrice, atr, multiplier float64, side string) (*StopLoss, error) {
	if entryPrice <= 0 || atr <= 0 || multiplier <= 0 {
		return nil, fmt.Errorf("invalid parameters")
	}
	if side != "long" && side != "short" {
		return nil, fmt.Errorf("side must be 'long' or 'short'")
	}
	var stopPrice float64
	if side == "long" {
		stopPrice = entryPrice - (atr * multiplier)
	} else {
		stopPrice = entryPrice + (atr * multiplier)
	}
	if stopPrice <= 0 {
		return nil, fmt.Errorf("invalid stop price")
	}
	return &StopLoss{Price: stopPrice, Type: "atr"}, nil
}

type TrailingStop struct {
	InitialPrice float64
	CurrentPrice float64
	TrailAmount  float64
	HighestPrice float64
	Side         string
}

func NewTrailingStop(entryPrice, trailAmount float64, side string) (*TrailingStop, error) {
	if entryPrice <= 0 || trailAmount <= 0 {
		return nil, fmt.Errorf("invalid parameters")
	}
	if side != "long" && side != "short" {
		return nil, fmt.Errorf("side must be 'long' or 'short'")
	}
	var initialStop float64
	if side == "long" {
		initialStop = entryPrice - trailAmount
	} else {
		initialStop = entryPrice + trailAmount
	}
	return &TrailingStop{
		InitialPrice: initialStop,
		CurrentPrice: initialStop,
		TrailAmount:  trailAmount,
		HighestPrice: entryPrice,
		Side:         side,
	}, nil
}

func (t *TrailingStop) Update(currentPrice float64) {
	if currentPrice <= 0 {
		return
	}
	if t.Side == "long" {
		if currentPrice > t.HighestPrice {
			t.HighestPrice = currentPrice
			newStop := t.HighestPrice - t.TrailAmount
			if newStop > t.CurrentPrice {
				t.CurrentPrice = newStop
			}
		}
	} else {
		if currentPrice < t.HighestPrice || t.HighestPrice == t.InitialPrice {
			t.HighestPrice = currentPrice
			newStop := t.HighestPrice + t.TrailAmount
			if newStop < t.CurrentPrice || t.CurrentPrice == t.InitialPrice {
				t.CurrentPrice = newStop
			}
		}
	}
}

func (t *TrailingStop) IsTriggered(currentPrice float64) bool {
	if t.Side == "long" {
		return currentPrice <= t.CurrentPrice
	}
	return currentPrice >= t.CurrentPrice
}

func CalculateRiskReward(entryPrice, stopPrice, targetPrice float64, side string) (float64, error) {
	if entryPrice <= 0 || stopPrice <= 0 || targetPrice <= 0 {
		return 0, fmt.Errorf("all prices must be positive")
	}
	var risk, reward float64
	if side == "long" {
		risk = entryPrice - stopPrice
		reward = targetPrice - entryPrice
	} else {
		risk = stopPrice - entryPrice
		reward = entryPrice - targetPrice
	}
	if risk <= 0 || reward <= 0 {
		return 0, fmt.Errorf("invalid risk/reward")
	}
	return math.Abs(reward / risk), nil
}
