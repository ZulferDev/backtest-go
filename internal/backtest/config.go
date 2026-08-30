package backtest

import (
	"github.com/ZulferDev/backtest-go/internal/execution"
	"github.com/ZulferDev/backtest-go/pkg/data"
	"github.com/ZulferDev/backtest-go/pkg/sdk"
)

// EngineConfig holds configuration for backtest engine
type EngineConfig struct {
	FeeModel      execution.FeeModel
	SlippageModel execution.SlippageModel
	EnableFees    bool
	EnableSlippage bool
}

// DefaultEngineConfig returns default configuration (no fees/slippage)
func DefaultEngineConfig() EngineConfig {
	return EngineConfig{
		FeeModel:       nil,
		SlippageModel:  nil,
		EnableFees:     false,
		EnableSlippage: false,
	}
}

// RealisticEngineConfig returns realistic config with fees and slippage
func RealisticEngineConfig() EngineConfig {
	return EngineConfig{
		FeeModel:       execution.BinanceSpotFeeModel(),
		SlippageModel:  execution.DefaultSlippageModel(),
		EnableFees:     true,
		EnableSlippage: true,
	}
}

// NewEngineWithConfig creates a new backtest engine with custom config
func NewEngineWithConfig(strategy sdk.Strategy, historicalData []data.OHLCV, initialCash float64, config EngineConfig) *Engine {
	engine := NewEngine(strategy, historicalData, initialCash)
	
	// Add execution simulator if fees/slippage enabled
	if config.EnableFees || config.EnableSlippage {
		var feeModel execution.FeeModel
		var slippageModel execution.SlippageModel
		
		if config.EnableFees {
			feeModel = config.FeeModel
		}
		if config.EnableSlippage {
			slippageModel = config.SlippageModel
		}
		
		engine.executionSim = execution.NewExecutionSimulator(feeModel, slippageModel)
	}
	
	return engine
}

// applyExecutionCosts applies fees and slippage to order execution
func (e *Engine) applyExecutionCosts(price, quantity float64, side string) (fillPrice float64, fee float64) {
	if e.executionSim == nil {
		// No execution costs
		return price, 0
	}
	
	fillPrice, _, fee, err := e.executionSim.SimulateExecution(price, quantity, side)
	if err != nil {
		// Fallback to no costs if simulation fails
		return price, 0
	}
	
	return fillPrice, fee
}
