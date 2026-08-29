package signal

import "fmt"

type Timeframe string

const (
	TF1m  Timeframe = "1m"
	TF5m  Timeframe = "5m"
	TF15m Timeframe = "15m"
	TF1h  Timeframe = "1h"
	TF4h  Timeframe = "4h"
	TF1d  Timeframe = "1d"
)

func TimeframeToMinutes(tf Timeframe) (int, error) {
	switch tf {
	case TF1m:
		return 1, nil
	case TF5m:
		return 5, nil
	case TF15m:
		return 15, nil
	case TF1h:
		return 60, nil
	case TF4h:
		return 240, nil
	case TF1d:
		return 1440, nil
	default:
		return 0, fmt.Errorf("unknown timeframe: %s", tf)
	}
}

func TimeframeToSeconds(tf Timeframe) (int64, error) {
	minutes, err := TimeframeToMinutes(tf)
	if err != nil {
		return 0, err
	}
	return int64(minutes * 60), nil
}

func AlignTimestamp(timestamp int64, tf Timeframe) (int64, error) {
	seconds, err := TimeframeToSeconds(tf)
	if err != nil {
		return 0, err
	}
	return (timestamp / seconds) * seconds, nil
}

type OHLCV struct {
	Timestamp int64
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    float64
}

func AggregateToHigherTimeframe(bars []OHLCV, targetTF Timeframe) ([]OHLCV, error) {
	if len(bars) == 0 {
		return []OHLCV{}, nil
	}
	targetSeconds, err := TimeframeToSeconds(targetTF)
	if err != nil {
		return nil, err
	}
	var aggregated []OHLCV
	var currentBar *OHLCV
	for _, bar := range bars {
		alignedTime := (bar.Timestamp / targetSeconds) * targetSeconds
		if currentBar == nil || currentBar.Timestamp != alignedTime {
			if currentBar != nil {
				aggregated = append(aggregated, *currentBar)
			}
			currentBar = &OHLCV{
				Timestamp: alignedTime,
				Open:      bar.Open,
				High:      bar.High,
				Low:       bar.Low,
				Close:     bar.Close,
				Volume:    bar.Volume,
			}
		} else {
			if bar.High > currentBar.High {
				currentBar.High = bar.High
			}
			if bar.Low < currentBar.Low {
				currentBar.Low = bar.Low
			}
			currentBar.Close = bar.Close
			currentBar.Volume += bar.Volume
		}
	}
	if currentBar != nil {
		aggregated = append(aggregated, *currentBar)
	}
	return aggregated, nil
}
