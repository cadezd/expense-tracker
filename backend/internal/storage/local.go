package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/cadezd/expense-tracker/internal/common"
	"github.com/google/uuid"
)

type LocalStorage struct {
	baseDir            string
	maxFileSizeInBytes int64
	allowedMIMETypes   map[string]string
}

func NewLocalStorage(baseDir string, maxFileSizeInBytes int64) *LocalStorage {
	return &LocalStorage{
		baseDir:            baseDir,
		maxFileSizeInBytes: maxFileSizeInBytes,
		allowedMIMETypes: map[string]string{
			"application/pdf": ".pdf",
			"image/jpeg":      ".jpg",
			"image/png":       ".png",
		},
	}
}

func (ls *LocalStorage) Save(
	ctx context.Context,
	originalFilename string,
	reader io.Reader,
) (*StoredFile, error) {
	buff := make([]byte, 512)

	n, err := reader.Read(buff)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("read file header: %w", err)
	}
	if n == 0 {
		return nil, ErrEmptyFile
	}

	mimeType := http.DetectContentType(buff[:n])
	extension, ok := ls.allowedMIMETypes[mimeType]
	if !ok {
		return nil, ErrUnsupportedMIMEType
	}

	storedFilename := uuid.NewString() + extension

	now := time.Now()
	relativeDir := filepath.Join(
		"uploads",
		fmt.Sprintf("%04d", now.Year()),
		fmt.Sprintf("%02d", int(now.Month())),
	)

	absoluteDir := filepath.Join(ls.baseDir, relativeDir)
	if err := os.MkdirAll(absoluteDir, 0755); err != nil {
		return nil, fmt.Errorf("create upload dir: %w", err)
	}

	relativePath := filepath.Join(relativeDir, storedFilename)
	absolutePath := filepath.Join(ls.baseDir, relativePath)

	file, err := os.OpenFile(
		absolutePath,
		os.O_WRONLY|os.O_CREATE|os.O_EXCL,
		0644,
	)
	if err != nil {
		return nil, fmt.Errorf("create stored file: %w", err)
	}
	defer file.Close()

	limitedReader := io.LimitReader(
		io.MultiReader(bytes.NewReader(buff[:n]), reader),
		ls.maxFileSizeInBytes+1, // So we can check if the file is too large
	)

	// Here we use custom context aware copy function
	size, err := common.Copy(ctx, file, limitedReader)
	if err != nil {
		_ = os.Remove(absolutePath)
		return nil, fmt.Errorf("write stored file: %w", err)
	}

	if size > ls.maxFileSizeInBytes {
		_ = os.Remove(absolutePath)
		return nil, ErrFileTooLarge
	}

	return &StoredFile{
		OriginalFilename: originalFilename,
		StoredFilename:   storedFilename,
		RelativePath:     relativePath,
		MIMEType:         mimeType,
		Size:             size,
	}, nil
}

func (ls *LocalStorage) Open(
	ctx context.Context,
	relativePath string,
) (io.ReadCloser, error) {
	_ = ctx

	absolutePath, err := ls.resolveAbsolutePath(relativePath)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(absolutePath)
	if err != nil {
		return nil, err
	}

	return file, nil
}

func (ls *LocalStorage) Delete(
	ctx context.Context,
	relativePath string,
) error {
	absolutePath, err := ls.resolveAbsolutePath(relativePath)
	if err != nil {
		return err
	}

	if err := os.Remove(absolutePath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return fmt.Errorf("delete file: %w", err)
	}

	return nil
}

func (ls *LocalStorage) resolveAbsolutePath(relativePath string) (string, error) {
	cleanedPath := filepath.Clean(relativePath)
	if cleanedPath == "." {
		return "", ErrInvalidPath
	}
	if filepath.IsAbs(cleanedPath) {
		return "", ErrInvalidPath
	}

	absoluteBase, err := filepath.Abs(ls.baseDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve base dir: %w", err)
	}

	absolutePath, err := filepath.Abs(filepath.Join(absoluteBase, cleanedPath))
	if err != nil {
		return "", fmt.Errorf("failed to resolve file path: %w", err)
	}

	return absolutePath, nil
}
