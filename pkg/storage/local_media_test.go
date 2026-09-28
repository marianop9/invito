package storage

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalMediaStorage(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "invito-media-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	storage := NewLocalMediaStorage(tempDir, "/uploads")
	ctx := context.Background()

	// 1x1 8-bit PNG sample
	validPNG := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D,
		0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4, 0x89, 0x00, 0x00, 0x00,
		0x0A, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00, 0x00, 0x00, 0x00, 0x49,
		0x45, 0x4E, 0x44, 0xAE, 0x42, 0x60, 0x82,
	}

	// Minimal JPEG sample header
	validJPEG := []byte{
		0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01,
		0x01, 0x01, 0x00, 0x48, 0x00, 0x48, 0x00, 0x00, 0xFF, 0xDB, 0x00, 0x43,
	}

	// Minimal WebP sample
	validWebP := []byte("RIFF\x1a\x00\x00\x00WEBPVP8 \x0e\x00\x00\x00\x30\x01\x00\x9d\x01\x2a\x01\x00\x01\x00")

	// Minimal GIF sample
	validGIF := []byte("GIF89a\x01\x00\x01\x00\x80\x00\x00\xff\xff\xff\x00\x00\x00!\xf9\x04\x01\x00\x00\x00\x00,\x00\x00\x00\x00\x01\x00\x01\x00\x00\x02\x02D\x01\x00;")

	t.Run("Save valid PNG image", func(t *testing.T) {
		info, err := storage.Save(ctx, SaveMediaInput{
			Slug:         "wedding-preview",
			OriginalName: "photo.png",
			Reader:       bytes.NewReader(validPNG),
			Size:         int64(len(validPNG)),
		})
		if err != nil {
			t.Fatalf("unexpected error saving PNG: %v", err)
		}

		if info.ContentType != "image/png" {
			t.Errorf("expected ContentType image/png, got %s", info.ContentType)
		}
		if !strings.HasPrefix(info.Filename, "wedding-preview-") || !strings.HasSuffix(info.Filename, ".png") {
			t.Errorf("unexpected filename format: %s", info.Filename)
		}
		if info.URL != "/uploads/"+info.Filename {
			t.Errorf("unexpected URL: %s", info.URL)
		}
		if info.Size != int64(len(validPNG)) {
			t.Errorf("expected size %d, got %d", len(validPNG), info.Size)
		}

		exists, err := storage.Exists(ctx, info.Filename)
		if err != nil || !exists {
			t.Errorf("expected saved file to exist on disk")
		}
	})

	t.Run("Save valid JPEG image", func(t *testing.T) {
		info, err := storage.Save(ctx, SaveMediaInput{
			Slug:         "birthday_party",
			OriginalName: "sunset.jpg",
			Reader:       bytes.NewReader(validJPEG),
		})
		if err != nil {
			t.Fatalf("unexpected error saving JPEG: %v", err)
		}

		if info.ContentType != "image/jpeg" {
			t.Errorf("expected ContentType image/jpeg, got %s", info.ContentType)
		}
		if !strings.HasPrefix(info.Filename, "birthday-party-") || !strings.HasSuffix(info.Filename, ".jpg") {
			t.Errorf("unexpected filename: %s", info.Filename)
		}
	})

	t.Run("Save valid WebP image", func(t *testing.T) {
		info, err := storage.Save(ctx, SaveMediaInput{
			Slug:         "lucas-fiesta",
			OriginalName: "cover.webp",
			Reader:       bytes.NewReader(validWebP),
		})
		if err != nil {
			t.Fatalf("unexpected error saving WebP: %v", err)
		}

		if info.ContentType != "image/webp" {
			t.Errorf("expected ContentType image/webp, got %s", info.ContentType)
		}
		if !strings.HasSuffix(info.Filename, ".webp") {
			t.Errorf("expected .webp extension, got %s", info.Filename)
		}
	})

	t.Run("Save valid GIF image", func(t *testing.T) {
		info, err := storage.Save(ctx, SaveMediaInput{
			OriginalName: "animation.gif",
			Reader:       bytes.NewReader(validGIF),
		})
		if err != nil {
			t.Fatalf("unexpected error saving GIF: %v", err)
		}

		if info.ContentType != "image/gif" {
			t.Errorf("expected ContentType image/gif, got %s", info.ContentType)
		}
		// Empty slug defaults to "event"
		if !strings.HasPrefix(info.Filename, "event-") || !strings.HasSuffix(info.Filename, ".gif") {
			t.Errorf("expected event- prefix and .gif suffix, got %s", info.Filename)
		}
	})

	t.Run("Collision-resistant filenames for identical content", func(t *testing.T) {
		info1, err := storage.Save(ctx, SaveMediaInput{
			Slug:   "dup-test",
			Reader: bytes.NewReader(validPNG),
		})
		if err != nil {
			t.Fatalf("save 1 failed: %v", err)
		}

		info2, err := storage.Save(ctx, SaveMediaInput{
			Slug:   "dup-test",
			Reader: bytes.NewReader(validPNG),
		})
		if err != nil {
			t.Fatalf("save 2 failed: %v", err)
		}

		if info1.Filename == info2.Filename {
			t.Errorf("expected different filenames for duplicate saves, got %s", info1.Filename)
		}
	})

	t.Run("Reject unsupported media type", func(t *testing.T) {
		textData := []byte("Hello world, this is a plain text file, not an image!")
		_, err := storage.Save(ctx, SaveMediaInput{
			Slug:   "malicious",
			Reader: bytes.NewReader(textData),
		})
		if err != ErrInvalidMediaType {
			t.Errorf("expected ErrInvalidMediaType, got %v", err)
		}

		pdfData := []byte("%PDF-1.4\n%...\nHello PDF")
		_, err = storage.Save(ctx, SaveMediaInput{
			Slug:   "document",
			Reader: bytes.NewReader(pdfData),
		})
		if err != ErrInvalidMediaType {
			t.Errorf("expected ErrInvalidMediaType for PDF, got %v", err)
		}
	})

	t.Run("Reject empty file", func(t *testing.T) {
		_, err := storage.Save(ctx, SaveMediaInput{
			Slug:   "empty",
			Reader: bytes.NewReader([]byte{}),
		})
		if err != ErrEmptyFile {
			t.Errorf("expected ErrEmptyFile, got %v", err)
		}
	})

	t.Run("Enforce maximum upload size limit", func(t *testing.T) {
		smallStorage := NewLocalMediaStorage(tempDir, "/uploads")
		smallStorage.SetMaxUploadSize(50) // 50 bytes limit

		_, err := smallStorage.Save(ctx, SaveMediaInput{
			Slug:   "too-large",
			Reader: bytes.NewReader(validPNG), // 67 bytes > 50 bytes
		})
		if err != ErrFileTooLarge {
			t.Errorf("expected ErrFileTooLarge, got %v", err)
		}
	})

	t.Run("Sanitize slug with path traversal attacks", func(t *testing.T) {
		info, err := storage.Save(ctx, SaveMediaInput{
			Slug:   "../../etc/passwd!@#$",
			Reader: bytes.NewReader(validPNG),
		})
		if err != nil {
			t.Fatalf("unexpected error saving with malicious slug: %v", err)
		}

		// Ensure no path traversal in filename
		if strings.Contains(info.Filename, "..") || strings.Contains(info.Filename, "/") || strings.Contains(info.Filename, "\\") {
			t.Errorf("filename contains path separators: %s", info.Filename)
		}
		// Slug should be sanitized cleanly
		if !strings.HasPrefix(info.Filename, "etc-passwd-") {
			t.Errorf("expected sanitized slug prefix 'etc-passwd-', got: %s", info.Filename)
		}

		// Verify file was written inside tempDir and not outside
		expectedPath := filepath.Join(tempDir, info.Filename)
		if _, err := os.Stat(expectedPath); err != nil {
			t.Errorf("file not written to expected target path: %v", err)
		}
	})

	t.Run("Delete existing file", func(t *testing.T) {
		info, err := storage.Save(ctx, SaveMediaInput{
			Slug:   "to-delete",
			Reader: bytes.NewReader(validPNG),
		})
		if err != nil {
			t.Fatalf("failed to save file: %v", err)
		}

		if err := storage.Delete(ctx, info.Filename); err != nil {
			t.Errorf("expected successful deletion, got %v", err)
		}

		exists, err := storage.Exists(ctx, info.Filename)
		if err != nil || exists {
			t.Errorf("expected file to not exist after deletion")
		}

		// Deleting non-existent file returns ErrFileNotFound
		if err := storage.Delete(ctx, info.Filename); err != ErrFileNotFound {
			t.Errorf("expected ErrFileNotFound for non-existent file, got %v", err)
		}
	})

	t.Run("Save valid MP3 audio with ID3 tag", func(t *testing.T) {
		validMP3ID3 := []byte("ID3\x04\x00\x00\x00\x00\x00\x23TIT2\x00\x00\x00\x07\x00\x00\x00Wedding")
		info, err := storage.Save(ctx, SaveMediaInput{
			Slug:         "wedding-waltz",
			OriginalName: "waltz.mp3",
			Reader:       bytes.NewReader(validMP3ID3),
		})
		if err != nil {
			t.Fatalf("unexpected error saving MP3: %v", err)
		}
		if info.ContentType != "audio/mpeg" {
			t.Errorf("expected ContentType audio/mpeg, got %s", info.ContentType)
		}
		if !strings.HasPrefix(info.Filename, "wedding-waltz-") || !strings.HasSuffix(info.Filename, ".mp3") {
			t.Errorf("unexpected filename format: %s", info.Filename)
		}
	})

	t.Run("Save valid MP3 audio with raw MPEG sync frame", func(t *testing.T) {
		validMP3Raw := []byte{0xff, 0xfb, 0x90, 0x64, 0x00, 0x00, 0x00, 0x00, 0x55, 0x55, 0x55, 0x55}
		info, err := storage.Save(ctx, SaveMediaInput{
			Slug:         "party-beats",
			OriginalName: "track.mp3",
			Reader:       bytes.NewReader(validMP3Raw),
		})
		if err != nil {
			t.Fatalf("unexpected error saving raw MP3: %v", err)
		}
		if info.ContentType != "audio/mpeg" {
			t.Errorf("expected ContentType audio/mpeg, got %s", info.ContentType)
		}
		if !strings.HasSuffix(info.Filename, ".mp3") {
			t.Errorf("expected .mp3 suffix, got %s", info.Filename)
		}
	})

	t.Run("Save valid M4A audio", func(t *testing.T) {
		validM4A := []byte{0x00, 0x00, 0x00, 0x20, 'f', 't', 'y', 'p', 'M', '4', 'A', ' ', 0x00, 0x00, 0x02, 0x00}
		info, err := storage.Save(ctx, SaveMediaInput{
			Slug:         "sunset-acoustic",
			OriginalName: "guitar.m4a",
			Reader:       bytes.NewReader(validM4A),
		})
		if err != nil {
			t.Fatalf("unexpected error saving M4A: %v", err)
		}
		if info.ContentType != "audio/mp4" {
			t.Errorf("expected ContentType audio/mp4, got %s", info.ContentType)
		}
		if !strings.HasSuffix(info.Filename, ".m4a") {
			t.Errorf("expected .m4a suffix, got %s", info.Filename)
		}
	})

	t.Run("Save valid OGG audio", func(t *testing.T) {
		validOGG := []byte("OggS\x00\x02\x00\x00\x00\x00\x00\x00\x00\x00\x01\x00\x00\x00\x00\x00")
		info, err := storage.Save(ctx, SaveMediaInput{
			Slug:         "ambient-sound",
			OriginalName: "nature.ogg",
			Reader:       bytes.NewReader(validOGG),
		})
		if err != nil {
			t.Fatalf("unexpected error saving OGG: %v", err)
		}
		if info.ContentType != "audio/ogg" {
			t.Errorf("expected ContentType audio/ogg, got %s", info.ContentType)
		}
		if !strings.HasSuffix(info.Filename, ".ogg") {
			t.Errorf("expected .ogg suffix, got %s", info.Filename)
		}
	})

	t.Run("Save valid WAV audio", func(t *testing.T) {
		validWAV := []byte("RIFF\x24\x00\x00\x00WAVEfmt \x10\x00\x00\x00\x01\x00\x02\x00")
		info, err := storage.Save(ctx, SaveMediaInput{
			Slug:         "chimes",
			OriginalName: "intro.wav",
			Reader:       bytes.NewReader(validWAV),
		})
		if err != nil {
			t.Fatalf("unexpected error saving WAV: %v", err)
		}
		if info.ContentType != "audio/wav" {
			t.Errorf("expected ContentType audio/wav, got %s", info.ContentType)
		}
		if !strings.HasSuffix(info.Filename, ".wav") {
			t.Errorf("expected .wav suffix, got %s", info.Filename)
		}
	})
}
