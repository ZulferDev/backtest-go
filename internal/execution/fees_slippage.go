package execution

import (
	"fmt"
)

// FeeModel defines interface for fee calculation
type FeeModel interface {
	Calculate(price, quantity float64, side string) float64
}

// FixedPercentageFee implements simple percentage-based fee
type FixedPercentageFee struct {
	Percentage float64 // e.g., 0.001 for 0.1%
}

func NewFixedPercentageFee(percentage float64) *FixedPercentageFee {
	return &FixedPercentageFee{Percentage: percentage}
}

func (f *FixedPercentageFee) Calculate(price, quantity float64, side string) float64 {
	notional := price * quantity
	return notional * f.Percentage
}

// TieredFee implements tiered fee structure based on volume
type TieredFee struct {
	Tiers []FeeTier
}

type FeeTier struct {
	MaxVolume float64 // Maximum 30-day volume for this tier
	MakerFee  float64 // Maker fee percentage
	TakerFee  float64 // Taker fee percentage
}

func NewTieredFee(tiers []FeeTier) *TieredFee {
	return &TieredFee{Tiers: tiers}
}

func (t *TieredFee) Calculate(price, quantity float64, side string) float64 {
	// Simplified: assume taker fee (aggressive order)
	// In real implementation, would track 30-day volume
	notional := price * quantity
	
	// Use first tier as default
	if len(t.Tiers) == 0 {
		return 0
	}
	
	fee := t.Tiers[0].TakerFee
	return notional * fee
}

// MakerTakerFee differentiates between maker and taker orders
type MakerTakerFee struct {
	MakerFee float64
	TakerFee float64
}

func NewMakerTakerFee(makerFee, takerFee float64) *MakerTakerFee {
	return &MakerTakerFee{
		MakerFee: makerFee,
		TakerFee: takerFee,
	}
}

func (m *MakerTakerFee) Calculate(price, quantity float64, side string) float64 {
	notional := price * quantity
	// Market orders are always taker
	// Limit orders could be maker if they add liquidity
	// For backtest, assume market orders (taker)
	return notional * m.TakerFee
}

// SlippageModel defines interface for slippage calculation
type SlippageModel interface {
	Apply(price, quantity float64, side string) float64
}

// FixedSlippage applies constant slippage
type FixedSlippage struct {
	BasisPoints float64 // e.g., 10 for 0.1% (10 basis points)
}

func NewFixedSlippage(basisPoints float64) *FixedSlippage {
	return &FixedSlippage{BasisPoints: basisPoints}
}

func (f *FixedSlippage) Apply(price, quantity float64, side string) float64 {
	slippagePercent := f.BasisPoints / 10000.0
	
	if side == "buy" || side == "long" {
		// Buy slippage increases price
		return price * (1.0 + slippagePercent)
	}
	// Sell slippage decreases price
	return price * (1.0 - slippagePercent)
}

// VolumeBasedSlippage calculates slippage based on order size
type VolumeBasedSlippage struct {
	BaseSlippage    float64 // Base slippage in basis points
	VolumeImpact    float64 // Additional slippage per unit volume
	ReferenceVolume float64 // Reference volume for impact calculation
}

func NewVolumeBasedSlippage(baseSlippage, volumeImpact, referenceVolume float64) *VolumeBasedSlippage {
	return &VolumeBasedSlippage{
		BaseSlippage:    baseSlippage,
		VolumeImpact:    volumeImpact,
		ReferenceVolume: referenceVolume,
	}
}

func (v *VolumeBasedSlippage) Apply(price, quantity float64, side string) float64 {
	// Calculate volume ratio
	volumeRatio := quantity / v.ReferenceVolume
	
	// Total slippage = base + volume impact
	totalSlippageBps := v.BaseSlippage + (v.VolumeImpact * volumeRatio)
	slippagePercent := totalSlippageBps / 10000.0
	
	if side == "buy" || side == "long" {
		return price * (1.0 + slippagePercent)
	}
	return price * (1.0 - slippagePercent)
}

// SpreadBasedSlippage uses bid-ask spread for slippage
type SpreadBasedSlippage struct {
	SpreadBasisPoints float64 // Typical spread in basis points
	SpreadMultiplier  float64 // Multiplier for aggressive orders
}

func NewSpreadBasedSlippage(spreadBps, multiplier float64) *SpreadBasedSlippage {
	return &SpreadBasedSlippage{
		SpreadBasisPoints: spreadBps,
		SpreadMultiplier:  multiplier,
	}
}

func (s *SpreadBasedSlippage) Apply(price, quantity float64, side string) float64 {
	// Spread as percentage
	spread := s.SpreadBasisPoints / 10000.0
	effectiveSpread := spread * s.SpreadMultiplier
	
	if side == "buy" || side == "long" {
		// Pay the ask (mid + half spread)
		return price * (1.0 + effectiveSpread/2.0)
	}
	// Receive the bid (mid - half spread)
	return price * (1.0 - effectiveSpread/2.0)
}

// ExecutionSimulator combines fee and slippage models
type ExecutionSimulator struct {
	FeeModel      FeeModel
	SlippageModel SlippageModel
}

func NewExecutionSimulator(feeModel FeeModel, slippageModel SlippageModel) *ExecutionSimulator {
	return &ExecutionSimulator{
		FeeModel:      feeModel,
		SlippageModel: slippageModel,
	}
}

// SimulateExecution returns fill price and total cost including fees
func (e *ExecutionSimulator) SimulateExecution(price, quantity float64, side string) (fillPrice, totalCost, fee float64, err error) {
	if quantity <= 0 {
		return 0, 0, 0, fmt.Errorf("quantity must be positive")
	}
	
	// Apply slippage to get fill price
	fillPrice = price
	if e.SlippageModel != nil {
		fillPrice = e.SlippageModel.Apply(price, quantity, side)
	}
	
	// Calculate notional value
	notional := fillPrice * quantity
	
	// Calculate fee
	fee = 0
	if e.FeeModel != nil {
		fee = e.FeeModel.Calculate(fillPrice, quantity, side)
	}
	
	// Total cost includes notional + fee
	totalCost = notional + fee
	
	return fillPrice, totalCost, fee, nil
}

// Common presets for popular exchanges
func BinanceSpotFeeModel() FeeModel {
	return NewMakerTakerFee(0.001, 0.001) // 0.1% maker/taker
}

func BybitFuturesFeeModel() FeeModel {
	return NewMakerTakerFee(0.0002, 0.0006) // 0.02% maker, 0.06% taker
}

func DefaultSlippageModel() SlippageModel {
	return NewFixedSlippage(5.0) // 5 basis points (0.05%)
}

func AggressiveSlippageModel() SlippageModel {
	return NewVolumeBasedSlippage(10.0, 5.0, 1.0) // 10 bps base + 5 bps per unit
}
