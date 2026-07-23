package storage

import "errors"

var (
	ErrFileTooLarge        = errors.New("file too large")
	ErrUnsupportedMIMEType = errors.New("unsupported mime type")
	ErrEmptyFile           = errors.New("empty file")
	ErrInvalidPath         = errors.New("invalid path")
)
