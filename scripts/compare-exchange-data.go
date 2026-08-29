package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type OHLCV struct {
	Timestamp int64
	Close     float64
}

func fetchBinance() (OHLCV, error) {
	url := "https://api.binance.com/api/v3/klines?symbol=BTCUSDT&interval=1h&limit=1"
	resp, err := http.Get(url)
	if err != nil { return OHLCV{}, err }
	defer resp.Body.Close()

	var result [][]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil { return OHLCV{}, err }

	// Binance types: timestamp is float64 (from json), close is string
	ts := int64(result[0][0].(float64))
	closePrice, _ := strconv.ParseFloat(result[0][4].(string), 64)

	return OHLCV{Timestamp: ts, Close: closePrice}, nil
}

func fetchBybit() (OHLCV, error) {
	url := "https://api.bybit.com/v5/market/kline?category=linear&symbol=BTCUSDT&interval=60&limit=1"
	resp, err := http.Get(url)
	if err != nil { return OHLCV{}, err }
	defer resp.Body.Close()

	type Response struct {
		Result struct { List [][]string } `json:"result"`
	}
	var result Response
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil { return OHLCV{}, err }

	ts, _ := strconv.ParseInt(result.Result.List[0][0], 10, 64)
	closePrice, _ := strconv.ParseFloat(result.Result.List[0][4], 64)

	return OHLCV{Timestamp: ts, Close: closePrice}, nil
}

func main() {
	fmt.Println("Comparing Binance vs Bybit Latest 1h Close Price")

	binance, err := fetchBinance()
	if err != nil { panic(err) }
	fmt.Printf("Binance: Timestamp: %d, Close: %f\n", binance.Timestamp, binance.Close)

	bybit, err := fetchBybit()
	if err != nil { panic(err) }
	fmt.Printf("Bybit: Timestamp: %d, Close: %f\n", bybit.Timestamp, bybit.Close)

	diff := binance.Close - bybit.Close
	if diff < 0 { diff = -diff }

	fmt.Printf("Absolute Difference: %.2f USDT\n", diff)
}
