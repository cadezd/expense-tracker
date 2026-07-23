package storage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type failingReader struct {
	err error
}

func (fr failingReader) Read(_ []byte) (int, error) {
	return 0, fr.err
}

func TestLocalStorage_Save_Negative(t *testing.T) {
	headerReadErr := errors.New("boom")

	testCases := []struct {
		name        string
		ctx         func() context.Context
		reader      func() io.Reader
		maxSize     int64
		expectedErr error
	}{
		{
			name:        "read file header error",
			ctx:         context.Background,
			reader:      func() io.Reader { return failingReader{err: headerReadErr} },
			maxSize:     10,
			expectedErr: headerReadErr,
		},
		{
			name:        "empty file",
			ctx:         context.Background,
			reader:      func() io.Reader { return bytes.NewReader(nil) },
			maxSize:     10,
			expectedErr: ErrEmptyFile,
		},
		{
			name:        "unsupported mime type",
			ctx:         context.Background,
			reader:      func() io.Reader { return bytes.NewReader([]byte("plain text only")) },
			maxSize:     10,
			expectedErr: ErrUnsupportedMIMEType,
		},
		{
			name:        "file too large",
			ctx:         context.Background,
			reader:      func() io.Reader { return bytes.NewReader([]byte("%PDF-1.4\nthis content is longer than 10 bytes!!!")) },
			maxSize:     10,
			expectedErr: ErrFileTooLarge,
		},
		{
			name: "canceled context",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			},
			reader:      func() io.Reader { return bytes.NewReader([]byte("%PDF-1.4\nfake pdf content")) },
			maxSize:     1024,
			expectedErr: context.Canceled,
		},
		{
			name: "deadline exceeded",
			ctx: func() context.Context {
				ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
				cancel()
				return ctx
			},
			reader:      func() io.Reader { return bytes.NewReader([]byte("%PDF-1.4\nfake pdf content")) },
			maxSize:     1024,
			expectedErr: context.DeadlineExceeded,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)
			localStorage := NewLocalStorage(t.TempDir(), tc.maxSize)

			savedFile, err := localStorage.Save(tc.ctx(), "racun.pdf", tc.reader())
			r.Error(err)
			r.Nil(savedFile)
			r.ErrorIs(err, tc.expectedErr)
		})
	}
}

func TestLocalStorage_Save(t *testing.T) {
	testCases := []struct {
		name         string
		filename     string
		content      []byte
		expectedMime string
		expectedExt  string
	}{
		{
			name:         "pdf",
			filename:     "receipt.pdf",
			content:      []byte("%PDF-1.4\nfake pdf content"),
			expectedMime: "application/pdf",
			expectedExt:  ".pdf",
		},
		{
			name:         "jpeg",
			filename:     "receipt.jpg",
			content:      []byte{0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10, 'J', 'F', 'I', 'F', 0x00, 0x01, 0x02},
			expectedMime: "image/jpeg",
			expectedExt:  ".jpg",
		},
		{
			name:         "png",
			filename:     "receipt.png",
			content:      []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d, 'I', 'H', 'D', 'R'},
			expectedMime: "image/png",
			expectedExt:  ".png",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)
			tmpDir := t.TempDir()
			localStorage := NewLocalStorage(tmpDir, 1024)

			savedFile, err := localStorage.Save(context.Background(), tc.filename, bytes.NewReader(tc.content))
			r.NoError(err)
			r.NotNil(savedFile)
			r.Equal(tc.filename, savedFile.OriginalFilename)
			r.Equal(tc.expectedMime, savedFile.MIMEType)
			r.Equal(tc.expectedExt, filepath.Ext(savedFile.StoredFilename))
			r.Equal(tc.expectedExt, filepath.Ext(savedFile.RelativePath))
			r.Equal(int64(len(tc.content)), savedFile.Size)

			storedContent, err := os.ReadFile(filepath.Join(tmpDir, savedFile.RelativePath))
			r.NoError(err)
			r.Equal(tc.content, storedContent)
			r.FileExists(filepath.Join(tmpDir, savedFile.RelativePath))
		})
	}
}

func TestLocalStorage_Delete_Negative(t *testing.T) {
	testCases := []struct {
		name         string
		relativePath string
		expectedErr  error
	}{
		{
			name:         "absolute path",
			relativePath: "/tmp/outside.pdf",
			expectedErr:  ErrInvalidPath,
		},
		{
			name:         "empty path",
			relativePath: ".",
			expectedErr:  ErrInvalidPath,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)
			localStorage := NewLocalStorage(t.TempDir(), 1024)

			err := localStorage.Delete(context.Background(), tc.relativePath)
			r.Error(err)
			r.ErrorIs(err, tc.expectedErr)
		})
	}
}

func TestLocalStorage_Delete(t *testing.T) {
	testCases := []struct {
		name         string
		relativePath string
		prepare      func(t *testing.T, baseDir string)
	}{
		{
			name:         "deletes existing file",
			relativePath: filepath.Join("uploads", "2026", "07", "receipt.pdf"),
			prepare: func(t *testing.T, baseDir string) {
				t.Helper()
				absolutePath := filepath.Join(baseDir, filepath.Join("uploads", "2026", "07", "receipt.pdf"))
				r := require.New(t)
				r.NoError(os.MkdirAll(filepath.Dir(absolutePath), 0755))
				r.NoError(os.WriteFile(absolutePath, []byte("payload"), 0644))
			},
		},
		{
			name:         "missing file is ignored",
			relativePath: filepath.Join("uploads", "missing.pdf"),
			prepare:      func(t *testing.T, baseDir string) {},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)
			tmpDir := t.TempDir()
			localStorage := NewLocalStorage(tmpDir, 1024)

			tc.prepare(t, tmpDir)

			err := localStorage.Delete(context.Background(), tc.relativePath)
			r.NoError(err)
			r.NoFileExists(filepath.Join(tmpDir, tc.relativePath))
		})
	}
}

func TestLocalStorage_Open_Negative(t *testing.T) {
	testCases := []struct {
		name         string
		relativePath string
		expectedErr  error
	}{
		{
			name:         "missing file",
			relativePath: filepath.Join("uploads", "missing.pdf"),
			expectedErr:  os.ErrNotExist,
		},
		{
			name:         "absolute path",
			relativePath: "/tmp/outside.pdf",
			expectedErr:  ErrInvalidPath,
		},
		{
			name:         "empty path",
			relativePath: ".",
			expectedErr:  ErrInvalidPath,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := require.New(t)
			localStorage := NewLocalStorage(t.TempDir(), 1024)

			file, err := localStorage.Open(context.Background(), tc.relativePath)
			r.Error(err)
			r.Nil(file)
			r.ErrorIs(err, tc.expectedErr)
		})
	}
}

func TestLocalStorage_Open(t *testing.T) {
	r := require.New(t)
	tmpDir := t.TempDir()
	localStorage := NewLocalStorage(tmpDir, 1024)

	relativePath := filepath.Join("uploads", "2026", "07", "receipt.pdf")
	absolutePath := filepath.Join(tmpDir, relativePath)
	r.NoError(os.MkdirAll(filepath.Dir(absolutePath), 0755))
	content := []byte("%PDF-1.4\nfake pdf content")
	r.NoError(os.WriteFile(absolutePath, content, 0644))

	file, err := localStorage.Open(context.Background(), relativePath)
	r.NoError(err)
	defer file.Close()

	openedContent, err := io.ReadAll(file)
	r.NoError(err)
	r.Equal(content, openedContent)
}
