package main

import (
	"flag"
	"fmt"
	"log"

	"invitation/pkg/ssg"
)

func main() {
	targetDir := flag.String("target", "_deploy", "Directory to export static HTML/CSS bundle (default '_deploy/')")
	seedDir := flag.String("seed", "seed", "Path to seed directory containing invitation JSON files (default 'seed/')")
	uploadsDir := flag.String("uploads", "uploads", "Path to directory for user-uploaded media files (default 'uploads/')")
	slug := flag.String("slug", "", "Optional single invitation slug to export (exports all if omitted)")
	createZip := flag.Bool("zip", false, "When exporting, also bundle the static files into a .zip archive")
	noIndex := flag.Bool("no-index", false, "Skip generating the index.html landing page in static export")
	flag.Parse()

	runSSGExport(*targetDir, *seedDir, *slug, *createZip, !*noIndex, *uploadsDir)
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
