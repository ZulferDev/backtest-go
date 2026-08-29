package validator

import (
	"errors"
	"fmt"
	"github.com/ZulferDev/backtest-go/pkg/data"
)

var (
	ErrNegativePrice  = errors.New("price cannot be negative")
	ErrNegativeVolume = errors.New("volume cannot be negative")
	ErrInvalidOHLC    = errors.New("high must be >= open, close and >= low")
)

func ValidateOHLCV(bar data.OHLCV) error {
	if bar.Open < 0 || bar.High < 0 || bar.Low < 0 || bar.Close < 0 {
		return ErrNegativePrice
	}
	if bar.Volume < 0 {
		return ErrNegativeVolume
	}
	if bar.High < bar.Open || bar.High < bar.Close || bar.High < bar.Low {
		return fmt.Errorf("%w: High=%f, Open=%f, Close=%f, Low=%f", ErrInvalidOHLC, bar.High, bar.Open, bar.Close, bar.Low)
	}
	if bar.Low > bar.Open || bar.Low > bar.Close {
		return fmt.Errorf("%w: Low=%f, Open=%f, Close=%f", ErrInvalidOHLC, bar.Low, bar.Open, bar.Close)
	}
	return nil
}
