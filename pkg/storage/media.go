package storage

import (
	"context"
	"errors"
	"io"
	"time"
)

var (
	// ErrFileTooLarge is returned when an uploaded media item exceeds the allowed byte limit.
	ErrFileTooLarge = errors.New("file exceeds maximum allowed size of 10MB")

	// ErrInvalidMediaType is returned when an uploaded item is not a supported image format.
	ErrInvalidMediaType = errors.New("unsupported media type: only JPEG, PNG, WebP, and GIF images are allowed")

	// ErrEmptyFile is returned when an uploaded media item has 0 bytes.
	ErrEmptyFile = errors.New("uploaded file cannot be empty")

	// ErrFileNotFound is returned when a requested media item does not exist.
	ErrFileNotFound = errors.New("media file not found")
)

// MediaInfo holds metadata and access information for a persisted media asset.
type MediaInfo struct {
	URL         string    `json:"url"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"content_type"`
	Size        int64     `json:"size"`
	CreatedAt   time.Time `json:"created_at"`
}

// SaveMediaInput encapsulates parameters and payload for storing media assets.
type SaveMediaInput struct {
	Slug         string    // Optional event slug namespace or prefix
	OriginalName string    // Original filename from client
	ContentType  string    // Declared content type header (may be empty or untrusted)
	Reader       io.Reader // File data stream
	Size         int64     // Size in bytes (-1 if unknown)
}

// MediaStorage abstracts binary media persistence (e.g. local filesystem, S3, Cloudflare R2).
type MediaStorage interface {
	// Save validates, stores, and indexes the media payload.
	Save(ctx context.Context, input SaveMediaInput) (*MediaInfo, error)

	// Exists checks if a given media file exists in storage.
	Exists(ctx context.Context, filename string) (bool, error)

	// Delete removes a media file from storage.
	Delete(ctx context.Context, filename string) error

	// PublicURL returns the public URL or relative path for the specified filename.
	PublicURL(filename string) string
}
