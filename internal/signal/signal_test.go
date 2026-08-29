package signal

import "testing"

func TestTimeframeToMinutes(t *testing.T) {
	tests := []struct {
		tf   Timeframe
		want int
	}{
		{TF1m, 1},
		{TF5m, 5},
		{TF15m, 15},
		{TF1h, 60},
		{TF4h, 240},
		{TF1d, 1440},
	}
	for _, tt := range tests {
		t.Run(string(tt.tf), func(t *testing.T) {
			got, err := TimeframeToMinutes(tt.tf)
			if err != nil {
				t.Errorf("TimeframeToMinutes() error = %v", err)
				return
			}
			if got != tt.want {
				t.Errorf("TimeframeToMinutes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAggregateToHigherTimeframe(t *testing.T) {
	bars := []OHLCV{
		{Timestamp: 1000, Open: 100, High: 105, Low: 99, Close: 103, Volume: 1000},
		{Timestamp: 2000, Open: 103, High: 108, Low: 102, Close: 107, Volume: 1500},
	}
	aggregated, err := AggregateToHigherTimeframe(bars, TF1h)
	if err != nil {
		t.Fatalf("AggregateToHigherTimeframe() error = %v", err)
	}
	if len(aggregated) == 0 {
		t.Error("Expected aggregated bars")
	}
}
