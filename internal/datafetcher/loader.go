package datafetcher

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/ZulferDev/backtest-go/pkg/data"
)

// LoadOHLCVFromFile loads OHLCV data from a JSON or CSV file
func LoadOHLCVFromFile(filePath string) ([]data.OHLCV, error) {
	if len(filePath) == 0 {
		return nil, fmt.Errorf("file path is empty")
	}

	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	// Determine file type by extension
	switch {
	case hasExtension(filePath, ".json"):
		return loadFromJSON(f)
	case hasExtension(filePath, ".csv"):
		return loadFromCSV(f)
	default:
		// Try JSON first, then CSV
		d, err := loadFromJSON(f)
		if err == nil {
			return d, nil
		}
		// Reset file pointer and try CSV
		f.Seek(0, 0)
		return loadFromCSV(f)
	}
}

func hasExtension(path, ext string) bool {
	return len(path) >= len(ext) && path[len(path)-len(ext):] == ext
}

// loadFromJSON loads OHLCV data from JSON format
// Expected format: [{"timestamp": 1234567890, "open": 100, "high": 105, "low": 95, "close": 102, "volume": 1000}, ...]
func loadFromJSON(r io.Reader) ([]data.OHLCV, error) {
	var raw []map[string]interface{}
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&raw); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	result := make([]data.OHLCV, 0, len(raw))
	for i, item := range raw {
		bar, err := parseJSONBar(item, i)
		if err != nil {
			return nil, err
		}
		result = append(result, bar)
	}

	return result, nil
}

func parseJSONBar(item map[string]interface{}, index int) (data.OHLCV, error) {
	ts, err := getInt64(item, "timestamp")
	if err != nil {
		return data.OHLCV{}, fmt.Errorf("bar %d: invalid timestamp: %w", index, err)
	}

	open, err := getFloat64(item, "open")
	if err != nil {
		return data.OHLCV{}, fmt.Errorf("bar %d: invalid open price: %w", index, err)
	}

	high, err := getFloat64(item, "high")
	if err != nil {
		return data.OHLCV{}, fmt.Errorf("bar %d: invalid high price: %w", index, err)
	}

	low, err := getFloat64(item, "low")
	if err != nil {
		return data.OHLCV{}, fmt.Errorf("bar %d: invalid low price: %w", index, err)
	}

	close, err := getFloat64(item, "close")
	if err != nil {
		return data.OHLCV{}, fmt.Errorf("bar %d: invalid close price: %w", index, err)
	}

	vol, err := getFloat64(item, "volume")
	if err != nil {
		vol = 0 // Volume is optional
	}

	return data.OHLCV{
		Timestamp: ts,
		Open:      open,
		High:      high,
		Low:       low,
		Close:     close,
		Volume:    vol,
	}, nil
}

// loadFromCSV loads OHLCV data from CSV format
// Expected format: timestamp,open,high,low,close,volume
func loadFromCSV(r io.Reader) ([]data.OHLCV, error) {
	reader := csv.NewReader(r)
	reader.TrimLeadingSpace = true

	result := []data.OHLCV{}
	lineNum := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV line %d: %w", lineNum+1, err)
		}

		lineNum++

		// Skip header row if present
		if lineNum == 1 && record[0] == "timestamp" {
			continue
		}

		if len(record) < 5 {
			return nil, fmt.Errorf("line %d: expected at least 5 columns, got %d", lineNum, len(record))
		}

		ts, err := strconv.ParseInt(record[0], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid timestamp: %w", lineNum, err)
		}

		open, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid open price: %w", lineNum, err)
		}

		high, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid high price: %w", lineNum, err)
		}

		low, err := strconv.ParseFloat(record[3], 64)
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid low price: %w", lineNum, err)
		}

		close, err := strconv.ParseFloat(record[4], 64)
		if err != nil {
			return nil, fmt.Errorf("line %d: invalid close price: %w", lineNum, err)
		}

		vol := 0.0
		if len(record) >= 6 {
			vol, err = strconv.ParseFloat(record[5], 64)
			if err != nil {
				vol = 0 // Volume is optional
			}
		}

		result = append(result, data.OHLCV{
			Timestamp: ts,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    vol,
		})
	}

	return result, nil
}

// Helper functions for type-safe parsing

func getInt64(m map[string]interface{}, key string) (int64, error) {
	v, ok := m[key]
	if !ok {
		return 0, fmt.Errorf("missing key '%s'", key)
	}

	switch val := v.(type) {
	case float64:
		return int64(val), nil
	case int:
		return int64(val), nil
	case int64:
		return val, nil
	case string:
		return strconv.ParseInt(val, 10, 64)
	default:
		return 0, fmt.Errorf("unexpected type for '%s': %T", key, v)
	}
}

func getFloat64(m map[string]interface{}, key string) (float64, error) {
	v, ok := m[key]
	if !ok {
		return 0, fmt.Errorf("missing key '%s'", key)
	}

	switch val := v.(type) {
	case float64:
		return val, nil
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case string:
		return strconv.ParseFloat(val, 64)
	default:
		return 0, fmt.Errorf("unexpected type for '%s': %T", key, v)
	}
}

// GenerateSampleData creates sample OHLCV data for testing
func GenerateSampleData(numBars int, startPrice float64) []data.OHLCV {
	result := make([]data.OHLCV, numBars)
	currentTime := time.Now().Add(-time.Duration(numBars) * time.Hour).Unix()

	price := startPrice
	for i := 0; i < numBars; i++ {
		change := (float64(i%10-5) / 100.0) * price // Simulate price movement
		open := price
		close := price + change
		high := max(open, close) * 1.02
		low := min(open, close) * 0.98

		result[i] = data.OHLCV{
			Timestamp: currentTime + int64(i)*3600,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    float64(1000 + i*10),
		}

		price = close
	}

	return result
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
