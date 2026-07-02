package receipt

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusUploaded   = "uploaded"
	StatusProcessing = "processing"
	StatusProcessed  = "processed"
	StatusFailed     = "failed"
)

type Receipt struct {
	ID               uuid.UUID `json:"id"`
	UserID           uuid.UUID `json:"user_id"`
	OriginalFilename string    `json:"original_filename"`
	StoredFilename   string    `json:"stored_filename"`
	StoragePath      string    `json:"-"`
	MimeType         string    `json:"mime_type"`
	FileSize         *int64    `json:"file_size"`
	Status           Status    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	ObjectVersion    int64     `json:"object_version"`
}
