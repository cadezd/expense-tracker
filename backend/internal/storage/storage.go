package storage

import (
	"context"
	"io"
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
