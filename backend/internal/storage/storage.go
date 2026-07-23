package storage

import (
	"context"
	"errors"
	"io"
)

var (
	ErrFileTooLarge        = errors.New("file too large")
	ErrUnsupportedMIMEType = errors.New("unsupported mime type")
	ErrEmptyFile           = errors.New("empty file")
	ErrInvalidPath         = errors.New("invalid path")
)

type StoredFile struct {
	OriginalFilename string
	StoredFilename   string
	RelativePath     string
	MIMEType         string
	Size             int64
}

type Storage interface {
	Save(
		ctx context.Context,
		originalFilename string,
		reader io.Reader,
	) (*StoredFile, error)

	Open(
		ctx context.Context,
		relativePath string,
	) (io.ReadCloser, error)

	Delete(
		ctx context.Context,
		relativePath string,
	) error
}
