package storage

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	// DefaultMaxUploadSize is 10 Megabytes (10MB).
	DefaultMaxUploadSize = 10 * 1024 * 1024
)

var (
	slugSanitizeRegex = regexp.MustCompile(`[^a-z0-9\-]+`)
	slugHyphenRegex   = regexp.MustCompile(`\-+`)
)

// LocalMediaStorage manages local filesystem storage for uploaded media assets.
type LocalMediaStorage struct {
	baseDir  string
	baseURL  string
	maxBytes int64
}

// NewLocalMediaStorage instantiates a new local media storage provider.
func NewLocalMediaStorage(baseDir, baseURL string) *LocalMediaStorage {
	if baseDir == "" {
		baseDir = "uploads"
	}
	if baseURL == "" {
		baseURL = "/uploads"
	}

	return &LocalMediaStorage{
		baseDir:  filepath.Clean(baseDir),
		baseURL:  strings.TrimSuffix(baseURL, "/"),
		maxBytes: DefaultMaxUploadSize,
	}
}

// SetMaxUploadSize overrides the default maximum file size limit.
func (s *LocalMediaStorage) SetMaxUploadSize(maxBytes int64) {
	if maxBytes > 0 {
		s.maxBytes = maxBytes
	}
}

// BaseDir returns the root storage directory path on the local filesystem.
func (s *LocalMediaStorage) BaseDir() string {
	return s.baseDir
}

// PublicURL returns the public HTTP URL for an asset.
func (s *LocalMediaStorage) PublicURL(filename string) string {
	base := filepath.Base(filename)
	return fmt.Sprintf("%s/%s", s.baseURL, base)
}

// Save validates, stores, and returns metadata for an uploaded media asset.
func (s *LocalMediaStorage) Save(ctx context.Context, input SaveMediaInput) (*MediaInfo, error) {
	if input.Reader == nil {
		return nil, ErrEmptyFile
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(s.baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory %s: %w", s.baseDir, err)
	}

	// Read first 512 bytes to sniff real MIME type
	header := make([]byte, 512)
	n, err := io.ReadFull(input.Reader, header)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, fmt.Errorf("failed to read media header: %w", err)
	}
	if n == 0 {
		return nil, ErrEmptyFile
	}
	header = header[:n]

	// Determine MIME type and canonical extension
	contentType, ext, err := s.sniffMIMEType(header)
	if err != nil {
		return nil, err
	}

	// Generate safe, collision-resistant filename: {slug}-{timestamp}-{rand}{ext}
	slug := s.sanitizeSlug(input.Slug)
	filename, err := s.generateFilename(slug, ext)
	if err != nil {
		return nil, fmt.Errorf("failed to generate unique filename: %w", err)
	}

	// Prepare temp file path in the same directory for atomic rename
	tempPath := filepath.Join(s.baseDir, fmt.Sprintf(".%s.tmp", filename))
	finalPath := filepath.Join(s.baseDir, filename)

	tmpFile, err := os.OpenFile(tempPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	// Cleanup temp file on any early exit or error
	cleanup := true
	defer func() {
		if cleanup {
			_ = tmpFile.Close()
			_ = os.Remove(tempPath)
		}
	}()

	// Write sniffed header bytes first
	written, err := tmpFile.Write(header)
	if err != nil {
		return nil, fmt.Errorf("failed to write header: %w", err)
	}

	// Copy remainder of stream, guarding against files larger than maxBytes
	limitReader := io.LimitReader(input.Reader, s.maxBytes-int64(written)+1)
	copied, err := io.Copy(tmpFile, limitReader)
	if err != nil {
		return nil, fmt.Errorf("failed to write media content: %w", err)
	}

	totalSize := int64(written) + copied
	if totalSize > s.maxBytes {
		return nil, ErrFileTooLarge
	}

	if err := tmpFile.Close(); err != nil {
		return nil, fmt.Errorf("failed to flush temp file: %w", err)
	}

	// Atomically rename into final target file
	if err := os.Rename(tempPath, finalPath); err != nil {
		return nil, fmt.Errorf("failed to finalize media file: %w", err)
	}

	cleanup = false

	return &MediaInfo{
		URL:         s.PublicURL(filename),
		Filename:    filename,
		ContentType: contentType,
		Size:        totalSize,
		CreatedAt:   time.Now().UTC(),
	}, nil
}

// Exists checks if a given media file exists in storage.
func (s *LocalMediaStorage) Exists(ctx context.Context, filename string) (bool, error) {
	base := filepath.Base(filename)
	if base == "." || base == "/" || base == "" {
		return false, nil
	}

	target := filepath.Join(s.baseDir, base)
	stat, err := os.Stat(target)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return !stat.IsDir(), nil
}

// Delete removes a media file from storage.
func (s *LocalMediaStorage) Delete(ctx context.Context, filename string) error {
	base := filepath.Base(filename)
	if base == "." || base == "/" || base == "" {
		return ErrFileNotFound
	}

	target := filepath.Join(s.baseDir, base)
	err := os.Remove(target)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrFileNotFound
		}
		return err
	}

	return nil
}

// sanitizeSlug sanitizes an arbitrary user string into a safe, URL-friendly prefix.
func (s *LocalMediaStorage) sanitizeSlug(slug string) string {
	cleaned := strings.ToLower(strings.TrimSpace(slug))
	cleaned = strings.ReplaceAll(cleaned, "/", "-")
	cleaned = strings.ReplaceAll(cleaned, "\\", "-")
	cleaned = strings.ReplaceAll(cleaned, ".", "-")
	cleaned = strings.ReplaceAll(cleaned, "_", "-")
	cleaned = strings.ReplaceAll(cleaned, " ", "-")
	cleaned = slugSanitizeRegex.ReplaceAllString(cleaned, "")
	cleaned = slugHyphenRegex.ReplaceAllString(cleaned, "-")
	cleaned = strings.Trim(cleaned, "-")

	if len(cleaned) > 40 {
		cleaned = cleaned[:40]
		cleaned = strings.Trim(cleaned, "-")
	}

	if cleaned == "" {
		return "event"
	}
	return cleaned
}

// sniffMIMEType verifies whether header bytes match supported image formats (JPEG, PNG, WebP, GIF).
func (s *LocalMediaStorage) sniffMIMEType(header []byte) (contentType string, extension string, err error) {
	if len(header) < 12 {
		return "", "", ErrInvalidMediaType
	}

	// 1. Explicit magic byte check for WebP (RIFF....WEBP)
	if len(header) >= 12 && bytes.Equal(header[0:4], []byte("RIFF")) && bytes.Equal(header[8:12], []byte("WEBP")) {
		return "image/webp", ".webp", nil
	}

	// 2. Explicit magic byte check for PNG (\x89PNG\r\n\x1a\n)
	if len(header) >= 8 && bytes.Equal(header[0:8], []byte("\x89PNG\r\n\x1a\n")) {
		return "image/png", ".png", nil
	}

	// 3. Explicit magic byte check for JPEG (\xff\xd8\xff)
	if len(header) >= 3 && bytes.Equal(header[0:3], []byte("\xff\xd8\xff")) {
		return "image/jpeg", ".jpg", nil
	}

	// 4. Explicit magic byte check for GIF (GIF87a / GIF89a)
	if len(header) >= 6 && (bytes.Equal(header[0:6], []byte("GIF87a")) || bytes.Equal(header[0:6], []byte("GIF89a"))) {
		return "image/gif", ".gif", nil
	}

	// 5. Standard library sniffing fallback
	detected := http.DetectContentType(header)
	switch {
	case strings.HasPrefix(detected, "image/jpeg"):
		return "image/jpeg", ".jpg", nil
	case strings.HasPrefix(detected, "image/png"):
		return "image/png", ".png", nil
	case strings.HasPrefix(detected, "image/webp"):
		return "image/webp", ".webp", nil
	case strings.HasPrefix(detected, "image/gif"):
		return "image/gif", ".gif", nil
	default:
		return "", "", ErrInvalidMediaType
	}
}

// generateFilename creates a collision-resistant filename: {slug}-{timestamp}-{rand8}{ext}
func (s *LocalMediaStorage) generateFilename(slug, ext string) (string, error) {
	randBytes := make([]byte, 4)
	if _, err := rand.Read(randBytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}

	timestamp := time.Now().Unix()
	randHex := hex.EncodeToString(randBytes)

	return fmt.Sprintf("%s-%d-%s%s", slug, timestamp, randHex, ext), nil
}
