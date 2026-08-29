package cache

import (
	"os"
	"testing"
	"github.com/ZulferDev/backtest-go/pkg/data"
)

func TestJSONStorage(t *testing.T) {
	tmpDir := "./tmp_cache_test"
	defer os.RemoveAll(tmpDir)

	storage := NewJSONStorage(tmpDir)

	testData := []data.OHLCV{
		{Timestamp: 1000, Open: 1.0, High: 1.5, Low: 0.9, Close: 1.2, Volume: 100},
		{Timestamp: 2000, Open: 1.2, High: 1.8, Low: 1.1, Close: 1.6, Volume: 150},
	}

	err := storage.Save("BTCUSDT", "1h", testData)
	if err != nil {
		t.Fatalf("failed to save cache: %v", err)
	}

	loaded, err := storage.Load("BTCUSDT", "1h")
	if err != nil {
		t.Fatalf("failed to load cache: %v", err)
	}

	if len(loaded) != len(testData) {
		t.Errorf("expected len %d, got %d", len(testData), len(loaded))
	}

	if loaded[0].Close != testData[0].Close {
		t.Errorf("data mismatch. expected Close %f, got %f", testData[0].Close, loaded[0].Close)
	}
}
