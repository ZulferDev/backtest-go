package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"github.com/ZulferDev/backtest-go/pkg/data"
)

type Client struct {
	HTTPClient *http.Client
	BaseURL    string
}

func NewClient() *Client {
	return &Client{
		HTTPClient: &http.Client{Timeout: 15 * time.Second},
		BaseURL:    "https://api.binance.com",
	}
}

// FetchKlines fetches raw Klines with context timeout and exponential backoff retry
func (c *Client) FetchKlines(ctx context.Context, symbol string, interval string, limit int, startTime, endTime int64) ([]data.OHLCV, error) {
	url := fmt.Sprintf("%s/api/v3/klines?symbol=%s&interval=%s&limit=%d", c.BaseURL, symbol, interval, limit)
	if startTime > 0 {
		url += fmt.Sprintf("&startTime=%d", startTime)
	}
	if endTime > 0 {
		url += fmt.Sprintf("&endTime=%d", endTime)
	}

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(1<<attempt) * 200 * time.Millisecond):
			}
		}

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, err
		}

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			lastErr = fmt.Errorf("binance api status code: %d", resp.StatusCode)
			continue
		}

		var raw [][]interface{}
		err = json.NewDecoder(resp.Body).Decode(&raw)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}

		var result []data.OHLCV
		for _, item := range raw {
			ts := int64(item[0].(float64))
			openP, _ := strconv.ParseFloat(item[1].(string), 64)
			highP, _ := strconv.ParseFloat(item[2].(string), 64)
			lowP, _ := strconv.ParseFloat(item[3].(string), 64)
			closeP, _ := strconv.ParseFloat(item[4].(string), 64)
			vol, _ := strconv.ParseFloat(item[5].(string), 64)

			result = append(result, data.OHLCV{
				Timestamp: ts,
				Open:      openP,
				High:      highP,
				Low:       lowP,
				Close:     closeP,
				Volume:    vol,
			})
		}
		return result, nil
	}
	return nil, fmt.Errorf("failed after 3 retries: %w", lastErr)
}
