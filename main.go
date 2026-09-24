package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"invitation/pkg/server"
	"invitation/pkg/ssg"
	"invitation/pkg/storage"
)

func main() {
	defaultPort := os.Getenv("PORT")
	if defaultPort == "" {
		defaultPort = "8080"
	}

	port := flag.String("port", defaultPort, "HTTP server port")
	dbPath := flag.String("db", "invito.db", "Path to SQLite database file")
	uploadsDir := flag.String("uploads", "uploads", "Path to directory for user-uploaded media files")
	exportDir := flag.String("export", "", "Export invitations to static HTML/CSS bundle in specified directory (e.g. '_demo')")
	createZip := flag.Bool("zip", false, "When exporting, also bundle the static files into a .zip archive")
	seedDir := flag.String("seed", "seed", "Path to seed directory containing invitation JSON files")
	slug := flag.String("slug", "", "Optional single invitation slug to export (exports all if omitted)")
	noIndex := flag.Bool("no-index", false, "Skip generating the index.html landing page in static export")
	flag.Parse()

	// If export flag is provided, run SSG exporter and exit
	if *exportDir != "" {
		runSSGExport(*exportDir, *seedDir, *slug, *createZip, !*noIndex, *uploadsDir)
		return
	}

	// Initialize SQLite persistent storage
	sqliteStore, err := storage.NewSQLiteStore(*dbPath, *seedDir)
	if err != nil {
		log.Fatalf("Failed to initialize SQLite store: %v", err)
	}
	defer sqliteStore.Close()

	// Normal server mode
	srv, err := server.New(server.Config{
		Port:         *port,
		SeedDir:      *seedDir,
		DatabasePath: *dbPath,
		Store:        sqliteStore,
		UploadsDir:   *uploadsDir,
	})
	if err != nil {
		log.Fatalf("Failed to initialize server: %v", err)
	}

	httpServer := &http.Server{
		Addr:         ":" + *port,
		Handler:      srv.Router(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Channel to listen for interrupt/terminate signals for graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		fmt.Printf("✦ Invito server is listening at http://localhost:%s\n", *port)
		fmt.Printf("✦ Health status endpoint: http://localhost:%s/health\n", *port)
		fmt.Printf("✦ Uploaded media directory: %s (served at /uploads/)\n", *uploadsDir)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited cleanly.")
}

func runSSGExport(outputDir, seedDir, targetSlug string, createZip, includeLanding bool, uploadsDir string) {
	fmt.Printf("✦ Starting Invito Static Site Exporter (SSG)...\n")
	fmt.Printf("  • Target directory : %s\n", outputDir)
	fmt.Printf("  • Seed directory   : %s\n", seedDir)
	fmt.Printf("  • Uploads directory: %s\n", uploadsDir)
	fmt.Printf("  • Landing page     : %t\n", includeLanding)
	fmt.Printf("  • Create ZIP       : %t\n", createZip)

	generator, err := ssg.New(ssg.Config{
		OutputDir:          outputDir,
		SeedDir:            seedDir,
		UploadsDir:         uploadsDir,
		IncludeLandingPage: includeLanding,
		CleanOutputDir:     true,
		CreateZip:          createZip,
	})
	if err != nil {
		log.Fatalf("Failed to initialize SSG generator: %v", err)
	}

	var res *ssg.Result
	if targetSlug != "" {
		res, err = generator.ExportInvitation(targetSlug)
	} else {
		res, err = generator.ExportAll()
	}

	if err != nil {
		log.Fatalf("SSG export failed: %v", err)
	}

	fmt.Printf("\n✦ Static export completed successfully!\n")
	fmt.Printf("  • Exported invitations (%d):\n", len(res.Invitations))
	for _, slug := range res.Invitations {
		fmt.Printf("    - /i/%s/index.html (+ calendar.ics)\n", slug)
	}
	fmt.Printf("  • Total files written : %d\n", len(res.FilesWritten))
	fmt.Printf("  • Total payload size  : %.2f KB\n", float64(res.TotalBytes)/1024.0)
	if res.ZipPath != "" {
		fmt.Printf("  • ZIP archive created : %s\n", res.ZipPath)
	}
}
