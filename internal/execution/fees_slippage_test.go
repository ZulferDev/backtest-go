package execution

import (
	"testing"
)

func TestFixedPercentageFee(t *testing.T) {
	fee := NewFixedPercentageFee(0.001) // 0.1%
	
	result := fee.Calculate(50000, 1.0, "buy")
	expected := 50.0 // 50000 * 1.0 * 0.001
	
	if result != expected {
		t.Errorf("Expected fee %.2f, got %.2f", expected, result)
	}
}

func TestMakerTakerFee(t *testing.T) {
	fee := NewMakerTakerFee(0.0002, 0.0006) // 0.02% maker, 0.06% taker
	
	result := fee.Calculate(50000, 1.0, "buy")
	expected := 30.0 // 50000 * 1.0 * 0.0006 (taker)
	
	if result != expected {
		t.Errorf("Expected taker fee %.2f, got %.2f", expected, result)
	}
}

func TestTieredFee(t *testing.T) {
	tiers := []FeeTier{
		{MaxVolume: 50000, MakerFee: 0.001, TakerFee: 0.001},
		{MaxVolume: 500000, MakerFee: 0.0008, TakerFee: 0.0008},
	}
	
	fee := NewTieredFee(tiers)
	result := fee.Calculate(50000, 1.0, "buy")
	expected := 50.0 // First tier: 50000 * 0.001
	
	if result != expected {
		t.Errorf("Expected fee %.2f, got %.2f", expected, result)
	}
}

func TestFixedSlippage(t *testing.T) {
	slippage := NewFixedSlippage(10.0) // 10 basis points = 0.1%
	
	// Buy side (price increases)
	buyPrice := slippage.Apply(50000, 1.0, "buy")
	expectedBuy := 50050.0 // 50000 * 1.001
	
	if buyPrice != expectedBuy {
		t.Errorf("Expected buy price %.2f, got %.2f", expectedBuy, buyPrice)
	}
	
	// Sell side (price decreases)
	sellPrice := slippage.Apply(50000, 1.0, "sell")
	expectedSell := 49950.0 // 50000 * 0.999
	
	if sellPrice != expectedSell {
		t.Errorf("Expected sell price %.2f, got %.2f", expectedSell, sellPrice)
	}
}

func TestVolumeBasedSlippage(t *testing.T) {
	// Base 5 bps, +2 bps per unit volume, reference 1.0
	slippage := NewVolumeBasedSlippage(5.0, 2.0, 1.0)
	
	// Small order (1.0 unit)
	smallPrice := slippage.Apply(50000, 1.0, "buy")
	// Total slippage: 5 + (2 * 1.0/1.0) = 7 bps = 0.07%
	expectedSmall := 50035.0 // 50000 * 1.0007
	
	if smallPrice != expectedSmall {
		t.Errorf("Expected small order price %.2f, got %.2f", expectedSmall, smallPrice)
	}
	
	// Large order (5.0 units)
	largePrice := slippage.Apply(50000, 5.0, "buy")
	// Total slippage: 5 + (2 * 5.0/1.0) = 15 bps = 0.15%
	expectedLarge := 50075.0 // 50000 * 1.0015
	
	if largePrice != expectedLarge {
		t.Errorf("Expected large order price %.2f, got %.2f", expectedLarge, largePrice)
	}
}

func TestSpreadBasedSlippage(t *testing.T) {
	// 20 bps spread, 1.5x multiplier for market orders
	slippage := NewSpreadBasedSlippage(20.0, 1.5)
	
	// Buy: pay the ask (mid + half effective spread)
	// Effective spread = 20 * 1.5 = 30 bps
	// Half spread = 15 bps = 0.15%
	buyPrice := slippage.Apply(50000, 1.0, "buy")
	expectedBuy := 50075.0 // 50000 * 1.0015
	
	if buyPrice != expectedBuy {
		t.Errorf("Expected buy price %.2f, got %.2f", expectedBuy, buyPrice)
	}
	
	// Sell: receive the bid (mid - half effective spread)
	sellPrice := slippage.Apply(50000, 1.0, "sell")
	expectedSell := 49925.0 // 50000 * 0.9985
	
	if sellPrice != expectedSell {
		t.Errorf("Expected sell price %.2f, got %.2f", expectedSell, sellPrice)
	}
}

func TestExecutionSimulator(t *testing.T) {
	feeModel := NewFixedPercentageFee(0.001)      // 0.1% fee
	slippageModel := NewFixedSlippage(10.0)       // 10 bps slippage
	
	simulator := NewExecutionSimulator(feeModel, slippageModel)
	
	// Simulate buy order
	fillPrice, totalCost, fee, err := simulator.SimulateExecution(50000, 1.0, "buy")
	
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	expectedFillPrice := 50050.0  // 50000 * 1.001 (slippage)
	expectedFee := 50.05          // 50050 * 1.0 * 0.001
	expectedTotalCost := 50100.05 // 50050 + 50.05
	
	if fillPrice != expectedFillPrice {
		t.Errorf("Expected fill price %.2f, got %.2f", expectedFillPrice, fillPrice)
	}
	
	if fee != expectedFee {
		t.Errorf("Expected fee %.2f, got %.2f", expectedFee, fee)
	}
	
	if totalCost != expectedTotalCost {
		t.Errorf("Expected total cost %.2f, got %.2f", expectedTotalCost, totalCost)
	}
}

func TestExecutionSimulatorInvalidQuantity(t *testing.T) {
	simulator := NewExecutionSimulator(nil, nil)
	
	_, _, _, err := simulator.SimulateExecution(50000, 0, "buy")
	if err == nil {
		t.Error("Expected error for zero quantity")
	}
	
	_, _, _, err = simulator.SimulateExecution(50000, -1.0, "buy")
	if err == nil {
		t.Error("Expected error for negative quantity")
	}
}

func TestBinancePreset(t *testing.T) {
	fee := BinanceSpotFeeModel()
	result := fee.Calculate(50000, 1.0, "buy")
	expected := 50.0 // 0.1% taker fee
	
	if result != expected {
		t.Errorf("Binance fee: expected %.2f, got %.2f", expected, result)
	}
}

func TestBybitPreset(t *testing.T) {
	fee := BybitFuturesFeeModel()
	result := fee.Calculate(50000, 1.0, "buy")
	expected := 30.0 // 0.06% taker fee
	
	if result != expected {
		t.Errorf("Bybit fee: expected %.2f, got %.2f", expected, result)
	}
}

func TestDefaultSlippage(t *testing.T) {
	slippage := DefaultSlippageModel()
	buyPrice := slippage.Apply(50000, 1.0, "buy")
	expectedBuy := 50025.0 // 5 bps = 0.05%
	
	if buyPrice != expectedBuy {
		t.Errorf("Default slippage: expected %.2f, got %.2f", expectedBuy, buyPrice)
	}
}

func TestRealisticScenario(t *testing.T) {
	// Binance-like setup: 0.1% fee, 5 bps slippage
	feeModel := BinanceSpotFeeModel()
	slippageModel := DefaultSlippageModel()
	simulator := NewExecutionSimulator(feeModel, slippageModel)
	
	// Buy 0.5 BTC at $50000
	fillPrice, totalCost, fee, err := simulator.SimulateExecution(50000, 0.5, "buy")
	
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	
	t.Logf("Buy 0.5 BTC @ $50000:")
	t.Logf("  Fill price: $%.2f (slippage applied)", fillPrice)
	t.Logf("  Fee: $%.2f", fee)
	t.Logf("  Total cost: $%.2f", totalCost)
	
	// Verify reasonable values
	if fillPrice < 50000 {
		t.Error("Buy fill price should be >= market price")
	}
	
	if fee <= 0 {
		t.Error("Fee should be positive")
	}
	
	if totalCost <= fillPrice*0.5 {
		t.Error("Total cost should include notional + fee")
	}
}
