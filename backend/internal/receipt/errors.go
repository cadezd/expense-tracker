package receipt

import "errors"

var (
	ErrNotFound      = errors.New("receipt not found")
	ErrInvalidOffset = errors.New("invalid receipt offset")
	ErrInvalidLimit  = errors.New("invalid receipt limit")
)
