package main

import (
	"context"
	"fmt"
	"backtest-go/internal/exchange/binance"
	"backtest-go/internal/exchange/bybit"
	"backtest-go/internal/normalizer"
	"backtest-go/internal/validator"
)

func main() {
	ctx := context.Background()

	fmt.Println("Testing Binance Client...")
	binanceClient := binance.NewClient()
	binanceBars, err := binanceClient.FetchKlines(ctx, "BTCUSDT", "1h", 5, 0, 0)
	if err != nil { panic(err) }
	
	for _, bar := range binanceBars {
		if err := validator.ValidateOHLCV(bar); err != nil {
			fmt.Printf("Validation error Binance: %v\n", err)
		}
	}
	fmt.Printf("Binance successfully fetched and validated %d bars. First ts: %d\n", len(binanceBars), binanceBars[0].Timestamp)

	fmt.Println("\nTesting Bybit Client...")
	bybitClient := bybit.NewClient()
	bybitBars, err := bybitClient.FetchKlines(ctx, "BTCUSDT", "60", 5, 0, 0)
	if err != nil { panic(err) }
	
	bybitBars = normalizer.ToChronological(bybitBars)

	for _, bar := range bybitBars {
		if err := validator.ValidateOHLCV(bar); err != nil {
			fmt.Printf("Validation error Bybit: %v\n", err)
		}
	}
	fmt.Printf("Bybit successfully fetched and validated %d bars. First ts: %d\n", len(bybitBars), bybitBars[0].Timestamp)
}
