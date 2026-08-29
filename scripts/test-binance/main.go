package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// test binance API klines
func main() {
	url := "https://api.binance.com/api/v3/klines?symbol=BTCUSDT&interval=1h&limit=5"

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	duration := time.Since(start)

	fmt.Printf("Binance API Status: %d\n", resp.StatusCode)
	fmt.Printf("Response Time: %s\n", duration)

	var result [][]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		panic(err)
	}

	fmt.Printf("Retrieved %d candles\n", len(result))
	if len(result) > 0 {
		fmt.Printf("First candle timestamp: %v\n", result[0][0])
		fmt.Printf("First candle close price: %v\n", result[0][4])
	}
}
