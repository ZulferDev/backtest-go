package bybit

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
		BaseURL:    "https://api.bybit.com",
	}
}

type bybitResponse struct {
	RetCode int `json:"retCode"`
	Result  struct {
		List [][]string `json:"list"`
	} `json:"result"`
}

// FetchKlines fetches raw Klines with retry logic
func (c *Client) FetchKlines(ctx context.Context, symbol string, interval string, limit int, startTime, endTime int64) ([]data.OHLCV, error) {
	url := fmt.Sprintf("%s/v5/market/kline?category=linear&symbol=%s&interval=%s&limit=%d", c.BaseURL, symbol, interval, limit)
	if startTime > 0 {
		url += fmt.Sprintf("&start=%d", startTime)
	}
	if endTime > 0 {
		url += fmt.Sprintf("&end=%d", endTime)
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
			lastErr = fmt.Errorf("bybit api status code: %d", resp.StatusCode)
			continue
		}

		var res bybitResponse
		err = json.NewDecoder(resp.Body).Decode(&res)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}

		if res.RetCode != 0 {
			lastErr = fmt.Errorf("bybit retcode: %d", res.RetCode)
			continue
		}

		// Bybit returns newest first, so we reverse it to chronological order (oldest first)
		list := res.Result.List
		var result []data.OHLCV
		for i := len(list) - 1; i >= 0; i-- {
			item := list[i]
			ts, _ := strconv.ParseInt(item[0], 10, 64)
			openP, _ := strconv.ParseFloat(item[1], 64)
			highP, _ := strconv.ParseFloat(item[2], 64)
			lowP, _ := strconv.ParseFloat(item[3], 64)
			closeP, _ := strconv.ParseFloat(item[4], 64)
			vol, _ := strconv.ParseFloat(item[5], 64)

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
