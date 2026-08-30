package sdk

import "errors"

// Common errors
var (
	ErrInvalidQuantity = errors.New("invalid quantity")
	ErrPositionExists  = errors.New("position already exists")
	ErrNoPosition      = errors.New("no position to close")
)
