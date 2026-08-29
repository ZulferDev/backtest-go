package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func main() {
	url := "https://api.bybit.com/v5/market/kline?category=linear&symbol=BTCUSDT&interval=60&limit=5"

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

	fmt.Printf("Bybit API Status: %d\n", resp.StatusCode)
	fmt.Printf("Response Time: %s\n", duration)

	type BybitResponse struct {
		RetCode int `json:"retCode"`
		Result  struct {
			List [][]interface{} `json:"list"`
		} `json:"result"`
	}

	var result BybitResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		panic(err)
	}

	fmt.Printf("Retrieved %d candles\n", len(result.Result.List))
	if len(result.Result.List) > 0 {
		// Bybit returns newest first
		candles := result.Result.List
		fmt.Printf("Most recent candle timestamp: %v\n", candles[0][0])
		fmt.Printf("Most recent candle close price: %v\n", candles[0][4])
	}
}
