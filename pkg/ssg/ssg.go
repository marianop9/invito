package ssg

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"invitation/pkg/calendar"
	"invitation/pkg/domain"
	"invitation/pkg/renderer"
	"invitation/pkg/storage"
	"invitation/web"
)

// Config configures static site generation and file bundling.
type Config struct {
	OutputDir          string // Directory where static files will be exported (e.g. "_demo")
	SeedDir            string // Seed directory containing invitation JSON files (default "seed")
	IncludeLandingPage bool   // If true, renders the index.html landing page at root (default true)
	CleanOutputDir     bool   // If true, clears the output directory before exporting (default true)
	CreateZip          bool   // If true, creates a .zip archive of the exported bundle
	ZipPath            string // Optional custom path for the zip archive (default "<OutputDir>.zip")
}

// Result summarizes the static site generation output.
type Result struct {
	OutputDir    string
	ZipPath      string
	Invitations  []string
	FilesWritten []string
	TotalBytes   int64
}

// Generator manages the static site export process.
type Generator struct {
	config   Config
	renderer *renderer.Renderer
	store    storage.Store
}

// New creates a new static site Generator with the given configuration.
func New(cfg Config) (*Generator, error) {
	if cfg.OutputDir == "" {
		cfg.OutputDir = "_demo"
	}
	if cfg.SeedDir == "" {
		cfg.SeedDir = "seed"
	}

	rnd, err := renderer.New()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize renderer: %w", err)
	}

	store, err := storage.NewSQLiteStore(":memory:", cfg.SeedDir)
	if err != nil {
		return nil, fmt.Errorf("failed to load seed invitations from %s: %w", cfg.SeedDir, err)
	}

	return &Generator{
		config:   cfg,
		renderer: rnd,
		store:    store,
	}, nil
}

// NewWithComponents creates a Generator with pre-configured store and renderer.
func NewWithComponents(cfg Config, store storage.Store, rnd *renderer.Renderer) *Generator {
	if cfg.OutputDir == "" {
		cfg.OutputDir = "_demo"
	}
	return &Generator{
		config:   cfg,
		renderer: rnd,
		store:    store,
	}
}

// ExportAll exports all invitations, embedded static assets, calendar files,
// and optionally the landing page and a zip archive.
func (g *Generator) ExportAll() (*Result, error) {
	res := &Result{
		OutputDir:    g.config.OutputDir,
		Invitations:  make([]string, 0),
		FilesWritten: make([]string, 0),
	}

	// 1. Prepare clean output directory
	if g.config.CleanOutputDir {
		if err := os.RemoveAll(g.config.OutputDir); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to clean output directory %s: %w", g.config.OutputDir, err)
		}
	}
	if err := os.MkdirAll(g.config.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory %s: %w", g.config.OutputDir, err)
	}

	// 2. Copy embedded static assets (CSS, JS, images)
	if err := g.copyEmbeddedStatic(res); err != nil {
		return nil, fmt.Errorf("failed to copy static assets: %w", err)
	}

	// 3. Render and export each invitation
	invitations, err := g.store.ListInvitations()
	if err != nil {
		return nil, fmt.Errorf("failed to list invitations: %w", err)
	}
	for _, inv := range invitations {
		if err := g.exportSingleInvitation(inv, res); err != nil {
			return nil, fmt.Errorf("failed to export invitation %s: %w", inv.Slug, err)
		}
		res.Invitations = append(res.Invitations, inv.Slug)
	}

	// 4. Render landing page index.html if requested
	if g.config.IncludeLandingPage {
		if err := g.exportLandingPage(invitations, res); err != nil {
			return nil, fmt.Errorf("failed to export landing page: %w", err)
		}
	}

	// 5. Create ZIP bundle if requested
	if g.config.CreateZip {
		zipPath := g.config.ZipPath
		if zipPath == "" {
			zipPath = g.config.OutputDir + ".zip"
		}
		if err := CreateZipArchive(g.config.OutputDir, zipPath); err != nil {
			return nil, fmt.Errorf("failed to create zip bundle: %w", err)
		}
		res.ZipPath = zipPath
	}

	return res, nil
}

// ExportInvitation exports a single invitation with all static assets to the configured output directory.
func (g *Generator) ExportInvitation(slug string) (*Result, error) {
	inv, err := g.store.GetInvitation(slug)
	if err != nil {
		return nil, fmt.Errorf("invitation not found: %w", err)
	}

	res := &Result{
		OutputDir:    g.config.OutputDir,
		Invitations:  []string{inv.Slug},
		FilesWritten: make([]string, 0),
	}

	// Prepare output directory
	if g.config.CleanOutputDir {
		if err := os.RemoveAll(g.config.OutputDir); err != nil && !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to clean output directory: %w", err)
		}
	}
	if err := os.MkdirAll(g.config.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Copy static assets
	if err := g.copyEmbeddedStatic(res); err != nil {
		return nil, fmt.Errorf("failed to copy static assets: %w", err)
	}

	// Export invitation
	if err := g.exportSingleInvitation(inv, res); err != nil {
		return nil, fmt.Errorf("failed to export invitation: %w", err)
	}

	// Render landing page if enabled
	if g.config.IncludeLandingPage {
		if err := g.exportLandingPage([]*domain.Invitation{inv}, res); err != nil {
			return nil, fmt.Errorf("failed to export landing page: %w", err)
		}
	}

	// Create ZIP archive if requested
	if g.config.CreateZip {
		zipPath := g.config.ZipPath
		if zipPath == "" {
			zipPath = g.config.OutputDir + ".zip"
		}
		if err := CreateZipArchive(g.config.OutputDir, zipPath); err != nil {
			return nil, fmt.Errorf("failed to create zip bundle: %w", err)
		}
		res.ZipPath = zipPath
	}

	return res, nil
}

// copyEmbeddedStatic copies all files under web/static to <OutputDir>/static.
func (g *Generator) copyEmbeddedStatic(res *Result) error {
	return fs.WalkDir(web.Files, "static", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		content, err := web.Files.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read embedded file %s: %w", path, err)
		}

		targetPath := filepath.Join(g.config.OutputDir, path)
		if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
			return fmt.Errorf("failed to create parent dir for %s: %w", targetPath, err)
		}

		if err := os.WriteFile(targetPath, content, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", targetPath, err)
		}

		res.FilesWritten = append(res.FilesWritten, path)
		res.TotalBytes += int64(len(content))
		return nil
	})
}

// exportSingleInvitation renders HTML and generates .ics for an invitation into <OutputDir>/i/<slug>/.
func (g *Generator) exportSingleInvitation(inv *domain.Invitation, res *Result) error {
	invDir := filepath.Join(g.config.OutputDir, "i", inv.Slug)
	if err := os.MkdirAll(invDir, 0755); err != nil {
		return fmt.Errorf("failed to create invitation dir %s: %w", invDir, err)
	}

	// 1. Render HTML
	var htmlBuf bytes.Buffer
	if err := g.renderer.RenderInvitation(&htmlBuf, inv); err != nil {
		return fmt.Errorf("failed to render HTML for %s: %w", inv.Slug, err)
	}

	htmlPath := filepath.Join(invDir, "index.html")
	if err := os.WriteFile(htmlPath, htmlBuf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", htmlPath, err)
	}

	relHTMLPath := filepath.Join("i", inv.Slug, "index.html")
	res.FilesWritten = append(res.FilesWritten, relHTMLPath)
	res.TotalBytes += int64(htmlBuf.Len())

	// 2. Generate and write iCalendar .ics
	icsContent := calendar.GenerateICS(inv)
	icsPath := filepath.Join(invDir, "calendar.ics")
	if err := os.WriteFile(icsPath, []byte(icsContent), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", icsPath, err)
	}

	relICSPath := filepath.Join("i", inv.Slug, "calendar.ics")
	res.FilesWritten = append(res.FilesWritten, relICSPath)
	res.TotalBytes += int64(len(icsContent))

	return nil
}

// exportLandingPage renders the index showcase landing page into <OutputDir>/index.html.
func (g *Generator) exportLandingPage(invitations []*domain.Invitation, res *Result) error {
	data := map[string]any{
		"Invitations": invitations,
		"CurrentYear": time.Now().Year(),
	}

	var buf bytes.Buffer
	if err := g.renderer.RenderIndex(&buf, data); err != nil {
		return fmt.Errorf("failed to render landing page: %w", err)
	}

	indexPath := filepath.Join(g.config.OutputDir, "index.html")
	if err := os.WriteFile(indexPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write landing page %s: %w", indexPath, err)
	}

	res.FilesWritten = append(res.FilesWritten, "index.html")
	res.TotalBytes += int64(buf.Len())

	return nil
}

// CreateZipArchive bundles a source directory into a .zip file.
func CreateZipArchive(sourceDir, targetZip string) error {
	// Ensure parent directory of zip exists
	if err := os.MkdirAll(filepath.Dir(targetZip), 0755); err != nil {
		return fmt.Errorf("failed to create zip target parent dir: %w", err)
	}

	zipFile, err := os.Create(targetZip)
	if err != nil {
		return fmt.Errorf("failed to create zip file %s: %w", targetZip, err)
	}
	defer zipFile.Close()

	w := zip.NewWriter(zipFile)
	defer w.Close()

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return fmt.Errorf("failed to compute relative path for %s: %w", path, err)
		}

		// Use forward slashes inside ZIP archive for cross-platform compatibility
		zipEntryName := filepath.ToSlash(relPath)

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return fmt.Errorf("failed to create zip header for %s: %w", path, err)
		}
		header.Name = zipEntryName
		header.Method = zip.Deflate

		writer, err := w.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("failed to create entry %s in zip: %w", zipEntryName, err)
		}

		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("failed to open file %s for zipping: %w", path, err)
		}
		defer file.Close()

		if _, err := io.Copy(writer, file); err != nil {
			return fmt.Errorf("failed to write %s to zip: %w", zipEntryName, err)
		}

		return nil
	})
}
