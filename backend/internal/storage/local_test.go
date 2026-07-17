package storage

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLocalStorage_SaveAndDelete(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	content := []byte("%PDF-1.4\nfake pdf content")

	localStorage := NewLocalStorage(tmpDir, 1024)
	savedFile, err := localStorage.Save(ctx, "racun.pdf", bytes.NewReader(content))
	r.NoError(err)
	r.Equal("application/pdf", savedFile.MIMEType)
	r.FileExists(filepath.Join(tmpDir, savedFile.RelativePath))

	err = localStorage.Delete(ctx, savedFile.RelativePath)
	r.NoError(err)
	r.NoFileExists(filepath.Join(tmpDir, savedFile.RelativePath))
}

func TestLocalStorage_UnsupportedMIMEType(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	content := []byte("this is just plain text, not a pdf or image")

	localStorage := NewLocalStorage(tmpDir, 1024)
	savedFile, err := localStorage.Save(ctx, "racun.pdf", bytes.NewReader(content))
	r.Error(err)
	r.Nil(savedFile)
	r.True(errors.Is(err, ErrUnsupportedMIMEType))
}

func TestLocalStorage_FileSizeTooLarget(t *testing.T) {
	r := require.New(t)
	ctx := context.Background()
	tmpDir := t.TempDir()

	content := []byte("%PDF-1.4\nthis content is longer than 10 bytes!!!")

	localStorage := NewLocalStorage(tmpDir, 10)
	savedFile, err := localStorage.Save(ctx, "racun.pdf", bytes.NewReader(content))
	r.Error(err)
	r.Nil(savedFile)
	r.True(errors.Is(err, ErrFileTooLarge))
}

func TestLocalStorage_ContextTimeout(t *testing.T) {
	r := require.New(t)
	ctx, cancel := context.WithTimeout(context.Background(), 0*time.Second)
	defer cancel()
	tmpDir := t.TempDir()

	content := []byte("%PDF-1.4\nthis content is longer than 10 bytes!!!")

	localStorage := NewLocalStorage(tmpDir, 10)
	savedFile, err := localStorage.Save(ctx, "racun.pdf", bytes.NewReader(content))
	r.Error(err)
	r.Nil(savedFile)
	r.True(errors.Is(err, context.DeadlineExceeded))
}

func TestLocalStorage_EmptyFile(t *testing.T) {
	r := require.New(t)
	ctx, cancel := context.WithTimeout(context.Background(), 0*time.Second)
	defer cancel()
	tmpDir := t.TempDir()

	content := []byte{}

	localStorage := NewLocalStorage(tmpDir, 10)
	savedFile, err := localStorage.Save(ctx, "racun.pdf", bytes.NewReader(content))
	r.Error(err)
	r.Nil(savedFile)
	r.True(errors.Is(err, ErrEmptyFile))
}
