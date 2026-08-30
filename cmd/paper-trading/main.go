package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ZulferDev/backtest-go/internal/paper"
	"github.com/ZulferDev/backtest-go/pkg/sdk"
)

func main() {
	symbol := flag.String("symbol", "btcusdt", "Trading symbol (e.g., btcusdt)")
	interval := flag.String("interval", "1m", "Candle interval (e.g., 1m, 5m, 15m, 1h)")
	initialCash := flag.Float64("cash", 10000.0, "Initial cash for paper trading")
	flag.Parse()

	log.Printf("Starting paper trading monitor for %s at %s interval", *symbol, *interval)
	log.Printf("Initial cash: $%.2f", *initialCash)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a simple monitoring strategy (buy and hold example)
	strategy := &MonitorStrategy{}

	// Create paper trading executor
	executor, err := paper.NewExecutor(ctx, strategy, *symbol, *interval, *initialCash)
	if err != nil {
		log.Fatalf("Failed to create executor: %v", err)
	}

	// Set up callbacks for monitoring
	executor.SetTradeCallback(func(trade paper.Trade) {
		log.Printf("TRADE EXECUTED: %s | Entry: %.2f | Exit: %.2f | PnL: %.2f | Fee: %.2f",
			trade.Side, trade.EntryPrice, trade.ExitPrice, trade.PnL, trade.Fee)
	})

	executor.SetStateUpdateCallback(func(state *paper.TradingState) {
		logState(state)
	})

	// Start paper trading
	if err := executor.Start(); err != nil {
		log.Fatalf("Failed to start executor: %v", err)
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Periodic state reporting
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-sigChan:
			log.Println("Received shutdown signal")
			executor.Stop()
			
			// Print final summary
			finalState := executor.GetState()
			printSummary(finalState)
			return

		case <-ticker.C:
			state := executor.GetState()
			printSummary(state)
		}
	}
}

// logState logs current trading state
func logState(state *paper.TradingState) {
	posInfo := "No position"
	if state.Position != nil {
		unrealizedPnL := 0.0
		if state.Position.Side == "long" {
			unrealizedPnL = (state.CurrentBar.Close - state.Position.EntryPrice) * state.Position.Size
		} else {
			unrealizedPnL = (state.Position.EntryPrice - state.CurrentBar.Close) * state.Position.Size
		}
		posInfo = fmt.Sprintf("%s %.4f @ %.2f (Unrealized PnL: %.2f)",
			state.Position.Side, state.Position.Size, state.Position.EntryPrice, unrealizedPnL)
	}

	log.Printf("BAR #%d | Price: %.2f | Equity: %.2f | %s",
		state.BarCount, state.CurrentBar.Close, state.Equity, posInfo)
}

// printSummary prints a summary of trading performance
func printSummary(state *paper.TradingState) {
	totalPnL := state.Equity - state.InitialCash
	returnPct := (totalPnL / state.InitialCash) * 100

	winCount := 0
	lossCount := 0
	totalWin := 0.0
	totalLoss := 0.0

	for _, trade := range state.Trades {
		if trade.PnL > 0 {
			winCount++
			totalWin += trade.PnL
		} else {
			lossCount++
			totalLoss += trade.PnL
		}
	}

	winRate := 0.0
	if len(state.Trades) > 0 {
		winRate = float64(winCount) / float64(len(state.Trades)) * 100
	}

	avgWin := 0.0
	if winCount > 0 {
		avgWin = totalWin / float64(winCount)
	}

	avgLoss := 0.0
	if lossCount > 0 {
		avgLoss = totalLoss / float64(lossCount)
	}

	log.Println("\n=== PAPER TRADING SUMMARY ===")
	log.Printf("Total Bars Processed: %d", state.BarCount)
	log.Printf("Initial Cash: $%.2f", state.InitialCash)
	log.Printf("Current Equity: $%.2f", state.Equity)
	log.Printf("Total PnL: $%.2f (%.2f%%)", totalPnL, returnPct)
	log.Printf("Total Trades: %d (Win: %d, Loss: %d)", len(state.Trades), winCount, lossCount)
	log.Printf("Win Rate: %.2f%%", winRate)
	log.Printf("Avg Win: $%.2f | Avg Loss: $%.2f", avgWin, avgLoss)
	
	if state.Position != nil {
		log.Printf("Current Position: %s %.4f @ %.2f", 
			state.Position.Side, state.Position.Size, state.Position.EntryPrice)
	} else {
		log.Println("Current Position: None")
	}
	log.Println("=============================\n")
}

// MonitorStrategy is a simple example strategy for monitoring
type MonitorStrategy struct{}

func (s *MonitorStrategy) Init(ctx sdk.InitContext) error {
	log.Println("Strategy initialized")
	return nil
}

func (s *MonitorStrategy) OnBar(ctx sdk.BarContext, bar sdk.OHLCV) error {
	// This is a passive monitoring strategy - no trades
	// Replace with actual strategy logic for active trading
	return nil
}

// Helper to print state as JSON
func printStateJSON(state *paper.TradingState) {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal state: %v", err)
		return
	}
	fmt.Println(string(data))
}
