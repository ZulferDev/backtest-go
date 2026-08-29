package validator

import (
	"testing"
	"github.com/ZulferDev/backtest-go/pkg/data"
)

func TestValidateOHLCV(t *testing.T) {
	tests := []struct {
		name    string
		bar     data.OHLCV
		expectErr bool
	}{
		{"Valid Bar", data.OHLCV{Timestamp: 1, Open: 10, High: 15, Low: 8, Close: 12, Volume: 100}, false},
		{"Negative Price", data.OHLCV{Timestamp: 1, Open: -10, High: 15, Low: 8, Close: 12, Volume: 100}, true},
		{"Negative Volume", data.OHLCV{Timestamp: 1, Open: 10, High: 15, Low: 8, Close: 12, Volume: -100}, true},
		{"Invalid High", data.OHLCV{Timestamp: 1, Open: 10, High: 9, Low: 8, Close: 12, Volume: 100}, true},
		{"Invalid Low", data.OHLCV{Timestamp: 1, Open: 10, High: 15, Low: 11, Close: 12, Volume: 100}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOHLCV(tt.bar)
			if (err != nil) != tt.expectErr {
				t.Errorf("expected error: %v, got: %v", tt.expectErr, err)
			}
		})
	}
}
