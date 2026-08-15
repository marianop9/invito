package ssg

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSSG_ExportAll(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "invito-ssg-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	generator, err := New(Config{
		OutputDir:          tempDir,
		SeedDir:            "../../seed",
		IncludeLandingPage: true,
		CleanOutputDir:     true,
		CreateZip:          false,
	})
	if err != nil {
		t.Fatalf("failed to initialize SSG: %v", err)
	}

	result, err := generator.ExportAll()
	if err != nil {
		t.Fatalf("ExportAll returned error: %v", err)
	}

	if len(result.Invitations) == 0 {
		t.Fatalf("expected at least 1 exported invitation, got 0")
	}

	if result.TotalBytes <= 0 {
		t.Errorf("expected TotalBytes > 0, got %d", result.TotalBytes)
	}

	// 1. Verify Landing Page
	indexPath := filepath.Join(tempDir, "index.html")
	indexContent, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("failed to read exported index.html: %v", err)
	}
	if !strings.Contains(string(indexContent), "Structured Event Invitations with Go") {
		t.Errorf("index.html missing expected title content")
	}
	if !strings.Contains(string(indexContent), "/i/sarah-and-alex-wedding") {
		t.Errorf("index.html missing wedding invitation link")
	}

	// 2. Verify Static Assets (CSS, JS, Images)
	expectedStaticFiles := []string{
		filepath.Join(tempDir, "static", "css", "base.css"),
		filepath.Join(tempDir, "static", "css", "invitation.css"),
		filepath.Join(tempDir, "static", "css", "themes", "botanical-elegance.css"),
		filepath.Join(tempDir, "static", "css", "themes", "golden-sunset.css"),
		filepath.Join(tempDir, "static", "js", "invitation.js"),
		filepath.Join(tempDir, "static", "img", "demo-hero.webp"),
		filepath.Join(tempDir, "static", "img", "demo-carousel-1.webp"),
	}
	for _, sf := range expectedStaticFiles {
		info, err := os.Stat(sf)
		if err != nil {
			t.Errorf("expected static file missing: %s", sf)
		} else if info.Size() == 0 {
			t.Errorf("static file is empty: %s", sf)
		}
	}

	// 3. Verify Wedding Invitation HTML & Calendar ICS
	weddingHTMLPath := filepath.Join(tempDir, "i", "sarah-and-alex-wedding", "index.html")
	weddingHTML, err := os.ReadFile(weddingHTMLPath)
	if err != nil {
		t.Fatalf("failed to read wedding index.html: %v", err)
	}
	weddingKeywords := []string{
		"Sarah &amp; Alex",
		"botanical-elegance",
		"demo-hero.webp",
		"carousel-track",
		"RSVP",
	}
	for _, kw := range weddingKeywords {
		if !strings.Contains(string(weddingHTML), kw) {
			t.Errorf("wedding HTML missing keyword %q", kw)
		}
	}

	weddingICSPath := filepath.Join(tempDir, "i", "sarah-and-alex-wedding", "calendar.ics")
	weddingICS, err := os.ReadFile(weddingICSPath)
	if err != nil {
		t.Fatalf("failed to read wedding calendar.ics: %v", err)
	}
	icsStr := string(weddingICS)
	if !strings.Contains(icsStr, "BEGIN:VCALENDAR") || !strings.Contains(icsStr, "END:VCALENDAR") {
		t.Errorf("invalid iCalendar file structure: %s", icsStr)
	}
	if !strings.Contains(icsStr, "UID:sarah-and-alex-wedding@invito.app") {
		t.Errorf("missing UID in calendar.ics: %s", icsStr)
	}

	// 4. Verify Birthday Invitation HTML
	birthdayHTMLPath := filepath.Join(tempDir, "i", "lucas-30th-birthday-sunset-fiesta", "index.html")
	birthdayHTML, err := os.ReadFile(birthdayHTMLPath)
	if err != nil {
		t.Fatalf("failed to read birthday index.html: %v", err)
	}
	if !strings.Contains(string(birthdayHTML), "Lucas is 30") {
		t.Errorf("birthday HTML missing title")
	}
}

func TestSSG_ExportAll_WithZip(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "invito-ssg-zip-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	outDir := filepath.Join(tempDir, "bundle")
	zipPath := filepath.Join(tempDir, "bundle.zip")

	generator, err := New(Config{
		OutputDir:          outDir,
		SeedDir:            "../../seed",
		IncludeLandingPage: true,
		CleanOutputDir:     true,
		CreateZip:          true,
		ZipPath:            zipPath,
	})
	if err != nil {
		t.Fatalf("failed to initialize SSG: %v", err)
	}

	result, err := generator.ExportAll()
	if err != nil {
		t.Fatalf("ExportAll returned error: %v", err)
	}

	if result.ZipPath != zipPath {
		t.Errorf("expected ZipPath %q, got %q", zipPath, result.ZipPath)
	}

	zipInfo, err := os.Stat(zipPath)
	if err != nil {
		t.Fatalf("failed to stat generated zip: %v", err)
	}
	if zipInfo.Size() == 0 {
		t.Fatalf("generated zip file is empty")
	}

	// Inspect ZIP archive contents
	zipReader, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("failed to open zip reader: %v", err)
	}
	defer zipReader.Close()

	foundIndex := false
	foundCSS := false
	foundWedding := false
	foundICS := false

	for _, file := range zipReader.File {
		name := filepath.ToSlash(file.Name)
		if name == "index.html" {
			foundIndex = true
		}
		if name == "static/css/invitation.css" {
			foundCSS = true
		}
		if name == "i/sarah-and-alex-wedding/index.html" {
			foundWedding = true
			rc, err := file.Open()
			if err != nil {
				t.Fatalf("failed to open zipped wedding index.html: %v", err)
			}
			content, _ := io.ReadAll(rc)
			rc.Close()
			if !strings.Contains(string(content), "Sarah &amp; Alex") {
				t.Errorf("zipped wedding index.html content corrupted")
			}
		}
		if name == "i/sarah-and-alex-wedding/calendar.ics" {
			foundICS = true
		}
	}

	if !foundIndex {
		t.Errorf("zip archive missing index.html")
	}
	if !foundCSS {
		t.Errorf("zip archive missing static/css/invitation.css")
	}
	if !foundWedding {
		t.Errorf("zip archive missing i/sarah-and-alex-wedding/index.html")
	}
	if !foundICS {
		t.Errorf("zip archive missing i/sarah-and-alex-wedding/calendar.ics")
	}
}

func TestSSG_ExportSingleInvitation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "invito-ssg-single-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	generator, err := New(Config{
		OutputDir:          tempDir,
		SeedDir:            "../../seed",
		IncludeLandingPage: false,
		CleanOutputDir:     true,
		CreateZip:          false,
	})
	if err != nil {
		t.Fatalf("failed to initialize SSG: %v", err)
	}

	result, err := generator.ExportInvitation("sarah-and-alex-wedding")
	if err != nil {
		t.Fatalf("ExportInvitation error: %v", err)
	}

	if len(result.Invitations) != 1 || result.Invitations[0] != "sarah-and-alex-wedding" {
		t.Errorf("unexpected exported invitations list: %v", result.Invitations)
	}

	// Verify wedding index exists
	weddingHTMLPath := filepath.Join(tempDir, "i", "sarah-and-alex-wedding", "index.html")
	if _, err := os.Stat(weddingHTMLPath); err != nil {
		t.Errorf("wedding HTML missing in single export: %v", err)
	}

	// Verify landing page was NOT rendered
	indexPath := filepath.Join(tempDir, "index.html")
	if _, err := os.Stat(indexPath); !os.IsNotExist(err) {
		t.Errorf("expected index.html to not exist when IncludeLandingPage=false")
	}

	// Test non-existent slug
	_, err = generator.ExportInvitation("unknown-event-xyz")
	if err == nil {
		t.Errorf("expected error when exporting unknown slug, got nil")
	}
}

func TestSSG_CleanOutputDir(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "invito-ssg-clean-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a dummy stale file
	staleFile := filepath.Join(tempDir, "stale_old_file.txt")
	if err := os.WriteFile(staleFile, []byte("stale"), 0644); err != nil {
		t.Fatalf("failed to write stale file: %v", err)
	}

	generator, err := New(Config{
		OutputDir:          tempDir,
		SeedDir:            "../../seed",
		IncludeLandingPage: true,
		CleanOutputDir:     true,
	})
	if err != nil {
		t.Fatalf("failed to initialize SSG: %v", err)
	}

	if _, err := generator.ExportAll(); err != nil {
		t.Fatalf("ExportAll error: %v", err)
	}

	// Check stale file was cleaned
	if _, err := os.Stat(staleFile); !os.IsNotExist(err) {
		t.Errorf("expected stale file to be removed when CleanOutputDir=true")
	}
}
