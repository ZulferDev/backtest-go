package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"backtest-go/pkg/data"
)

type JSONStorage struct {
	BaseDir string
}

func NewJSONStorage(baseDir string) *JSONStorage {
	return &JSONStorage{BaseDir: baseDir}
}

func (s *JSONStorage) Save(symbol string, timeframe string, series []data.OHLCV) error {
	if err := os.MkdirAll(s.BaseDir, 0755); err != nil {
		return err
	}

	filePath := filepath.Join(s.BaseDir, symbol+"_"+timeframe+".json")
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(series)
}

func (s *JSONStorage) Load(symbol string, timeframe string) ([]data.OHLCV, error) {
	filePath := filepath.Join(s.BaseDir, symbol+"_"+timeframe+".json")
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var series []data.OHLCV
	err = json.NewDecoder(file).Decode(&series)
	return series, err
}
