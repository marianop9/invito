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
	seedDir := flag.String("seed", "seed", "Path to seed directory containing invitation JSON files")
	flag.Parse()

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
