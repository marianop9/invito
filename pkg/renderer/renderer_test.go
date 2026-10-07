package renderer

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"invitation/pkg/domain"
	"invitation/pkg/storage"
)

func TestRenderer(t *testing.T) {
	r, err := New()
	if err != nil {
		t.Fatalf("failed to initialize renderer: %v", err)
	}

	date := time.Date(2026, time.September, 19, 16, 0, 0, 0, time.UTC)
	inv := &domain.Invitation{
		Version:   "1.0",
		Slug:      "test-gala",
		Title:     "Annual Charity Gala",
		Subtitle:  "An evening of elegance",
		DateStart: date,
		Timezone:  "America/New_York",
		Location: domain.Location{
			Name:    "The Observatory",
			Address: "100 Skyline Blvd",
			MapURL:  "https://maps.google.com/?q=The+Observatory",
		},
		Theme: domain.ThemeConfig{
			ID: domain.ThemeMidnightSoiree,
		},
		Sections: []domain.Section{
			&domain.HeroSection{
				SectionType:   domain.SectionHero,
				Eyebrow:       "Black Tie Gala",
				Vibe:          "An evening of elegance",
				Badge:         "Black Tie Only",
				ShowCountdown: true,
			},
			&domain.DetailsSection{
				SectionType:        domain.SectionDetails,
				ShowMapLink:        true,
				ShowCalendarButton: true,
			},
			&domain.TimelineSection{
				SectionType: domain.SectionTimeline,
				Items: []domain.TimelineItem{
					{Time: "7:00 PM", Title: "Reception"},
				},
			},
			&domain.RSVPSection{
				SectionType:  domain.SectionRSVP,
				Enabled:      true,
				MaxPartySize: 2,
			},
		},
	}

	t.Run("RenderInvitation outputs valid HTML with partials", func(t *testing.T) {
		var buf bytes.Buffer
		if err := r.RenderInvitation(&buf, inv); err != nil {
			t.Fatalf("RenderInvitation error: %v", err)
		}

		html := buf.String()
		expectedKeywords := []string{
			"Annual Charity Gala",
			"The Observatory",
			"theme-midnight-soiree",
			"Black Tie Gala",
			"RSVP",
			"inv-hero-content",
		}

		for _, kw := range expectedKeywords {
			if !strings.Contains(html, kw) {
				t.Errorf("expected HTML to contain %q, but missing", kw)
			}
		}
	})

	t.Run("RenderInvitation with Hero Cover Image and Carousel", func(t *testing.T) {
		weddingInv := *inv
		weddingInv.Sections = []domain.Section{
			&domain.HeroSection{
				SectionType:   domain.SectionHero,
				Badge:         "Wedding Celebration",
				CoverImageURL: "/static/img/demo-hero.webp",
			},
			&domain.CarouselSection{
				SectionType: domain.SectionCarousel,
				Title:       "Photo Moments",
				Images: []domain.CarouselImage{
					{URL: "/static/img/demo-carousel-1.webp", Caption: "Engagement at Big Sur"},
					{URL: "/static/img/demo-carousel-2.webp", Caption: "Tuscany Trip"},
				},
			},
		}

		var buf bytes.Buffer
		if err := r.RenderInvitation(&buf, &weddingInv); err != nil {
			t.Fatalf("RenderInvitation with image/carousel error: %v", err)
		}

		html := buf.String()
		expectedImgKeywords := []string{
			"has-cover-image",
			"/static/img/demo-hero.webp",
			"section-gallery",
			"carousel-track",
			"carousel-dots",
			"/static/img/demo-carousel-1.webp",
			"inv-carousel-backdrop",
			"inv-carousel-caption",
			"Engagement at Big Sur",
		}

		for _, kw := range expectedImgKeywords {
			if !strings.Contains(html, kw) {
				t.Errorf("expected rendered HTML to contain %q, but missing", kw)
			}
		}
	})

	t.Run("RenderInvitation with Carousel AspectRatio and Cover Fit", func(t *testing.T) {
		customCarouselInv := *inv
		customCarouselInv.Sections = []domain.Section{
			&domain.CarouselSection{
				SectionType: domain.SectionCarousel,
				Title:       "Custom Frame Gallery",
				AspectRatio: "16:9",
				Fit:         "cover",
				Images: []domain.CarouselImage{
					{URL: "/static/img/demo-carousel-1.webp", Caption: "Wide Photo"},
				},
			},
		}

		var buf bytes.Buffer
		if err := r.RenderInvitation(&buf, &customCarouselInv); err != nil {
			t.Fatalf("RenderInvitation error: %v", err)
		}

		html := buf.String()
		if !strings.Contains(html, "inv-carousel-aspect-16-9") {
			t.Errorf("expected HTML to contain 'inv-carousel-aspect-16-9', got:\n%s", html)
		}
		if !strings.Contains(html, "inv-carousel-fit-cover") {
			t.Errorf("expected HTML to contain 'inv-carousel-fit-cover', got:\n%s", html)
		}
		// In cover fit mode, backdrop layer should be skipped
		if strings.Contains(html, "inv-carousel-backdrop-wrap") {
			t.Errorf("expected cover fit carousel not to contain backdrop wrap")
		}
	})

	t.Run("RenderInvitation with Alternative Hero Banner Layout", func(t *testing.T) {
		bannerInv := *inv
		bannerInv.Sections = []domain.Section{
			&domain.HeroSection{
				SectionType:   domain.SectionHero,
				Layout:        "banner",
				Eyebrow:       "Wedding Celebration",
				Vibe:          "An intimate evening under the stars",
				Badge:         "Sep 19, 2026",
				CoverImageURL: "/static/img/demo-hero.webp",
				ShowCountdown: true,
			},
		}

		var buf bytes.Buffer
		if err := r.RenderInvitation(&buf, &bannerInv); err != nil {
			t.Fatalf("RenderInvitation with hero banner error: %v", err)
		}

		html := buf.String()
		expectedKeywords := []string{
			"inv-hero-banner",
			"inv-hero-banner-media",
			"inv-hero-banner-img",
			"inv-hero-banner-content",
			"Wedding Celebration",
		}

		for _, kw := range expectedKeywords {
			if !strings.Contains(html, kw) {
				t.Errorf("expected hero banner HTML to contain %q, but missing", kw)
			}
		}
	})

	t.Run("RenderInvitation with Quote, Text, and Multiple Images in Custom Order", func(t *testing.T) {
		fullInv := *inv
		fullInv.Sections = []domain.Section{
			&domain.HeroSection{
				SectionType: domain.SectionHero,
				Badge:       "Welcome",
			},
			&domain.QuoteSection{
				SectionType: domain.SectionQuote,
				Text:        "First quote: Two lives, one shared journey.",
				Author:      "Rumi",
			},
			&domain.ImageSection{
				SectionType: domain.SectionImage,
				URL:         "/static/img/venue.webp",
				Caption:     "The Botanical Glasshouse",
			},
			&domain.TextSection{
				SectionType: domain.SectionText,
				Title:       "Guest Notice",
				Text:        "Second message: Complimentary shuttle departs from hotel lobby.",
			},
			&domain.ImageSection{
				SectionType: domain.SectionImage,
				URL:         "/static/img/swatches.webp",
				Caption:     "Attire Inspiration Swatches",
			},
			&domain.ClosingSection{
				SectionType: domain.SectionClosing,
				Message:     "We can't wait to celebrate with you!",
				Signoff:     "With love,",
				Hosts:       "Sarah & Alex",
			},
		}

		var buf bytes.Buffer
		if err := r.RenderInvitation(&buf, &fullInv); err != nil {
			t.Fatalf("RenderInvitation error: %v", err)
		}

		html := buf.String()
		expectedKeywords := []string{
			"section-quote",
			"First quote: Two lives, one shared journey.",
			"Rumi",
			"section-image",
			"/static/img/venue.webp",
			"The Botanical Glasshouse",
			"section-text",
			"Guest Notice",
			"Second message: Complimentary shuttle departs from hotel lobby.",
			"/static/img/swatches.webp",
			"Attire Inspiration Swatches",
			"section-closing",
			"We can&#39;t wait to celebrate with you!",
			"With love,",
			"Sarah &amp; Alex",
		}

		for _, kw := range expectedKeywords {
			if !strings.Contains(html, kw) {
				t.Errorf("expected rendered HTML to contain %q, but missing", kw)
			}
		}

		// Verify ordering in output: First quote before Second message
		idxQuote := strings.Index(html, "First quote")
		idxFirstImg := strings.Index(html, "/static/img/venue.webp")
		idxText := strings.Index(html, "Second message")
		idxSecondImg := strings.Index(html, "/static/img/swatches.webp")
		idxClosing := strings.Index(html, "section-closing")

		if idxQuote == -1 || idxFirstImg == -1 || idxText == -1 || idxSecondImg == -1 || idxClosing == -1 {
			t.Fatalf("missing expected elements in rendered output")
		}

		if !(idxQuote < idxFirstImg && idxFirstImg < idxText && idxText < idxSecondImg && idxSecondImg < idxClosing) {
			t.Errorf("expected blocks to render in exact configured sequence: quote < first img < text < second img < closing")
		}
	})

	t.Run("RenderToHTMLString produces static SSG bundle", func(t *testing.T) {
		htmlStr, err := r.RenderToHTMLString(inv)
		if err != nil {
			t.Fatalf("RenderToHTMLString error: %v", err)
		}

		if len(htmlStr) < 100 || !strings.Contains(htmlStr, "<!DOCTYPE html>") {
			t.Errorf("invalid static HTML generated: %s", htmlStr)
		}
	})

	t.Run("Icon helper inlines embedded SVG files into details section", func(t *testing.T) {
		var buf bytes.Buffer
		if err := r.RenderInvitation(&buf, inv); err != nil {
			t.Fatalf("RenderInvitation error: %v", err)
		}

		html := buf.String()
		expectedIconStrings := []string{
			"icon-tabler-calendar-week",
			"icon-tabler-map-2",
			"icon-tabler-arrow-up-right",
			"<svg",
			"viewBox=\"0 0 24 24\"",
			"section-when",
			"section-where",
			"btn-view-map",
		}

		for _, s := range expectedIconStrings {
			if !strings.Contains(html, s) {
				t.Errorf("expected rendered HTML to contain icon markup %q, but was missing", s)
			}
		}
	})

	t.Run("RenderInvitation with RSVPExternalSection in open state", func(t *testing.T) {
		extInv := *inv
		extInv.Sections = []domain.Section{
			&domain.RSVPExternalSection{
				SectionType:      domain.SectionRSVPExternal,
				Enabled:          true,
				Title:            "Confirmación de Asistencia",
				Prompt:           "Por favor confirmá tu asistencia:",
				FormURL:          "https://tally.so/r/demo-form",
				ButtonLabel:      "Abrir Formulario",
				ContributionNote: "Seña de $10.000 para la reserva (Alias: cumple.fiesta)",
				ReceiptNote:      "Enviá el comprobante de transferencia al anfitrión por WhatsApp",
				CustomNote:       "Fecha límite: 15 de Agosto",
			},
		}

		var buf bytes.Buffer
		if err := r.RenderInvitation(&buf, &extInv); err != nil {
			t.Fatalf("RenderInvitation with external RSVP error: %v", err)
		}

		html := buf.String()
		expectedKeywords := []string{
			"section-rsvp-external",
			"Confirmación de Asistencia",
			"Por favor confirmá tu asistencia:",
			"https://tally.so/r/demo-form",
			"Abrir Formulario",
			"btn-rsvp-external",
			"rsvp-notes-group",
			"rsvp-note-row",
			"Reserva:",
			"Seña de $10.000 para la reserva (Alias: cumple.fiesta)",
			"Comprobante:",
			"Enviá el comprobante de transferencia al anfitrión por WhatsApp",
			"Fecha límite: 15 de Agosto",
		}

		for _, kw := range expectedKeywords {
			if !strings.Contains(html, kw) {
				t.Errorf("expected rendered HTML to contain %q, but was missing", kw)
			}
		}
	})

	t.Run("RenderInvitation with RSVPExternalSection in closed state", func(t *testing.T) {
		past := time.Now().Add(-48 * time.Hour)
		closedInv := *inv
		closedInv.Sections = []domain.Section{
			&domain.RSVPExternalSection{
				SectionType: domain.SectionRSVPExternal,
				Enabled:     true,
				Title:       "Confirmación de Asistencia",
				FormURL:     "https://tally.so/r/demo-form",
				Deadline:    &past,
				CustomNote:  "El evento ya no recibe nuevas confirmaciones.",
			},
		}

		var buf bytes.Buffer
		if err := r.RenderInvitation(&buf, &closedInv); err != nil {
			t.Fatalf("RenderInvitation with closed external RSVP error: %v", err)
		}

		html := buf.String()
		if !strings.Contains(html, "rsvp-closed-badge") || !strings.Contains(html, "Plazo Finalizado") {
			t.Errorf("expected closed badge in output, got:\n%s", html)
		}
		if strings.Contains(html, "btn-rsvp-external") {
			t.Errorf("expected closed RSVP not to display active CTA button")
		}
	})

	t.Run("RenderInvitation with SplashScreen and Music config", func(t *testing.T) {
		musicAndSplashInv := *inv
		musicAndSplashInv.Music = &domain.MusicConfig{
			URL:      "/uploads/music/romantic-piano.mp3",
			Title:    "Romantic Piano Serenade",
			Autoplay: true,
			Loop:     true,
		}
		musicAndSplashInv.SplashScreen = &domain.SplashScreenConfig{
			Enabled:            true,
			Title:              "Welcome to Our Celebration",
			Message:            "Please tap below to enter the invitation and enjoy background music.",
			ButtonText:         "Enter Celebration",
			BackgroundImageURL: "/uploads/images/splash-bg.webp",
		}

		var buf bytes.Buffer
		if err := r.RenderInvitation(&buf, &musicAndSplashInv); err != nil {
			t.Fatalf("RenderInvitation with splash & music error: %v", err)
		}

		html := buf.String()

		// Verify Splash Screen elements
		splashExpected := []string{
			`id="inv-splash-overlay"`,
			`class="inv-splash-overlay"`,
			`style="background-image: url('/uploads/images/splash-bg.webp');"`,
			`class="inv-splash-backdrop"`,
			`class="inv-splash-card"`,
			`class="inv-splash-seal"`,
			`Welcome to Our Celebration`,
			`Please tap below to enter the invitation and enjoy background music.`,
			`id="btn-splash-open"`,
			`class="inv-splash-cta"`,
			`Enter Celebration`,
			`class="inv-splash-arrow"`,
		}
		for _, exp := range splashExpected {
			if !strings.Contains(html, exp) {
				t.Errorf("expected HTML to contain splash screen element %q, but missing", exp)
			}
		}

		// Verify Music Audio and FAB elements
		musicExpected := []string{
			`id="inv-bg-audio"`,
			`preload="auto"`,
			`loop`,
			`<source src="/uploads/music/romantic-piano.mp3" type="audio/mpeg">`,
			`id="inv-music-toggle"`,
			`class="inv-music-fab"`,
			`data-autoplay="true"`,
			`title="Romantic Piano Serenade"`,
			`class="music-disc-icon"`,
			`class="icon-disc"`,
			`class="music-bars"`,
			`class="music-tooltip"`,
			`Romantic Piano Serenade`,
		}
		for _, exp := range musicExpected {
			if !strings.Contains(html, exp) {
				t.Errorf("expected HTML to contain music player element %q, but missing", exp)
			}
		}
	})

	t.Run("RenderInvitation with SplashScreen fallback defaults", func(t *testing.T) {
		fallbackInv := *inv
		fallbackInv.SplashScreen = &domain.SplashScreenConfig{
			Enabled: true,
		}

		var buf bytes.Buffer
		if err := r.RenderInvitation(&buf, &fallbackInv); err != nil {
			t.Fatalf("RenderInvitation error: %v", err)
		}

		html := buf.String()

		if !strings.Contains(html, `id="inv-splash-overlay"`) {
			t.Errorf("expected splash overlay to be rendered")
		}
		// Fallback to invitation title
		if !strings.Contains(html, `<h2 class="inv-splash-title">`) || !strings.Contains(html, fallbackInv.Title) {
			t.Errorf("expected splash title fallback to invitation Title %q", fallbackInv.Title)
		}
		// Fallback button text
		if !strings.Contains(html, "Open Invitation") {
			t.Errorf("expected fallback CTAButtonText 'Open Invitation'")
		}
		// No background-image style
		if strings.Contains(html, "background-image:") {
			t.Errorf("expected no background-image style when BackgroundImageURL is empty")
		}
		// No message paragraph
		if strings.Contains(html, "inv-splash-message") {
			t.Errorf("expected no message element when Message is empty")
		}
	})

	t.Run("RenderInvitation with disabled or omitted SplashScreen and Music", func(t *testing.T) {
		// Test legacy / omitted config
		var buf bytes.Buffer
		if err := r.RenderInvitation(&buf, inv); err != nil {
			t.Fatalf("RenderInvitation error: %v", err)
		}
		html := buf.String()

		unexpected := []string{
			"inv-splash-overlay",
			"btn-splash-open",
			"inv-bg-audio",
			"inv-music-toggle",
			"music-disc-icon",
		}
		for _, unexp := range unexpected {
			if strings.Contains(html, unexp) {
				t.Errorf("expected HTML without splash/music NOT to contain %q", unexp)
			}
		}

		// Test explicitly disabled splash screen
		disabledInv := *inv
		disabledInv.SplashScreen = &domain.SplashScreenConfig{
			Enabled: false,
			Title:   "Should not appear",
		}
		buf.Reset()
		if err := r.RenderInvitation(&buf, &disabledInv); err != nil {
			t.Fatalf("RenderInvitation error: %v", err)
		}
		htmlDisabled := buf.String()
		if strings.Contains(htmlDisabled, "inv-splash-overlay") || strings.Contains(htmlDisabled, "Should not appear") {
			t.Errorf("expected disabled splash screen NOT to render overlay")
		}
	})

	t.Run("RenderAdmin views output complete HTML documents with header and footer partials", func(t *testing.T) {
		adminViews := []struct {
			name   string
			render func(w *bytes.Buffer) error
		}{
			{
				name: "RenderAdminIndex",
				render: func(w *bytes.Buffer) error {
					return r.RenderAdminIndex(w, map[string]any{
						"Title":                "Overview",
						"CurrentYear":          2026,
						"TotalEvents":          0,
						"TotalConfirmedGuests": 0,
						"Events":               []any{},
					})
				},
			},
			{
				name: "RenderAdminRSVPs",
				render: func(w *bytes.Buffer) error {
					return r.RenderAdminRSVPs(w, map[string]any{
						"Title":       "RSVP Dashboard",
						"CurrentYear": 2026,
						"Invitation":  inv,
						"RSVPs":       []domain.RSVPSubmission{},
						"Stats":       storage.RSVPStats{},
					})
				},
			},
			{
				name: "RenderAdminEditor",
				render: func(w *bytes.Buffer) error {
					return r.RenderAdminEditor(w, map[string]any{
						"Title":          "Editor",
						"CurrentYear":    2026,
						"IsNew":          false,
						"Invitation":     inv,
						"InvitationJSON": `{"slug":"test-gala"}`,
					})
				},
			},
		}

		for _, tc := range adminViews {
			t.Run(tc.name, func(t *testing.T) {
				var buf bytes.Buffer
				if err := tc.render(&buf); err != nil {
					t.Fatalf("failed to render %s: %v", tc.name, err)
				}
				output := strings.TrimSpace(buf.String())

				if !strings.HasPrefix(output, "<!DOCTYPE html>") {
					t.Errorf("expected %s output to start with <!DOCTYPE html>", tc.name)
				}
				if !strings.Contains(output, `<header class="admin-header">`) {
					t.Errorf("expected %s to contain composed header component", tc.name)
				}
				if !strings.Contains(output, `<main class="admin-main">`) {
					t.Errorf("expected %s to contain admin-main container", tc.name)
				}
				if !strings.Contains(output, `<footer class="admin-footer">`) {
					t.Errorf("expected %s to contain composed footer component", tc.name)
				}
				if !strings.HasSuffix(output, "</html>") {
					t.Errorf("expected %s output to end with </html>", tc.name)
				}

				if tc.name == "RenderAdminEditor" {
					expectedStrings := []string{
						"4. Welcome Splash & Background Music",
						"5. Modular Section Canvas",
						"ensureSplashScreen().enabled",
						"invitation.splash_screen.title",
						"invitation.splash_screen.button_text",
						"invitation.splash_screen.message",
						"hasMusic()",
						"toggleMusic($event.target.checked)",
						"uploadAudioFile($event)",
						"invitation.music.url",
						"invitation.music.title",
						"invitation.music.autoplay",
						"invitation.music.loop",
						"🎵 Upload Audio",
						"sec.surface",
						"Fondo:",
						"Automático",
					}
					for _, exp := range expectedStrings {
						if !strings.Contains(output, exp) {
							t.Errorf("expected RenderAdminEditor to contain %q", exp)
						}
					}
				}
			})
		}
	})
}

func TestRendererAlternatingSurfaces(t *testing.T) {
	r, err := New()
	if err != nil {
		t.Fatalf("failed to initialize renderer: %v", err)
	}

	date := time.Date(2026, time.September, 19, 16, 0, 0, 0, time.UTC)

	t.Run("Alternates tones automatically and injects surface classes", func(t *testing.T) {
		inv := &domain.Invitation{
			Version:   "1.0",
			Slug:      "test-surfaces",
			Title:     "Surface Test Party",
			DateStart: date,
			Location: domain.Location{
				Name:    "Venue",
				Address: "123 Main St",
			},
			Theme: domain.ThemeConfig{
				ID: domain.ThemeBotanicalElegance,
			},
			Sections: []domain.Section{
				&domain.HeroSection{
					SectionType: domain.SectionHero,
				},
				&domain.DetailsSection{
					SectionType: domain.SectionDetails,
				},
				&domain.QuoteSection{
					SectionType: domain.SectionQuote,
					Text:        "A wonderful quotation about life",
					Author:      "Author Name",
				},
				&domain.CarouselSection{
					SectionType: domain.SectionCarousel,
					Title:       "Memories",
					Images: []domain.CarouselImage{
						{URL: "/img/1.webp", Caption: "First image"},
					},
				},
				&domain.TimelineSection{
					SectionType: domain.SectionTimeline,
					Title:       "Program",
					Items: []domain.TimelineItem{
						{Time: "19:00", Title: "Arrival"},
					},
				},
			},
		}

		var buf bytes.Buffer
		if err := r.RenderInvitation(&buf, inv); err != nil {
			t.Fatalf("RenderInvitation error: %v", err)
		}

		html := buf.String()

		// Quote follows Details (where-strip ends on contrast), so quote should be light
		if !strings.Contains(html, `inv-quote-section inv-block inv-block--light`) {
			t.Errorf("expected quote section to have inv-block--light class, got: %s", html)
		}

		// Carousel follows quote (light), so it alternates to contrast
		if !strings.Contains(html, `inv-carousel-section section-gallery inv-block inv-block--contrast`) {
			t.Errorf("expected carousel section to have inv-block--contrast class, got: %s", html)
		}

		// Timeline follows carousel (contrast), so it alternates to light
		if !strings.Contains(html, `inv-timeline-section inv-block inv-block--light`) {
			t.Errorf("expected timeline section to have inv-block--light class, got: %s", html)
		}
	})

	t.Run("Explicit user surface override is rendered in HTML", func(t *testing.T) {
		inv := &domain.Invitation{
			Version:   "1.0",
			Slug:      "test-surfaces-override",
			Title:     "Surface Override Party",
			DateStart: date,
			Location: domain.Location{
				Name:    "Venue",
				Address: "123 Main St",
			},
			Theme: domain.ThemeConfig{
				ID: domain.ThemeBotanicalElegance,
			},
			Sections: []domain.Section{
				&domain.QuoteSection{
					SectionType: domain.SectionQuote,
					Surface:     "contrast", // forced contrast
					Text:        "Dark moody quote",
				},
				&domain.CarouselSection{
					SectionType: domain.SectionCarousel,
					Surface:     "contrast", // two consecutive contrast blocks
					Title:       "Dark moody gallery",
					Images: []domain.CarouselImage{
						{URL: "/img/photo.webp"},
					},
				},
			},
		}

		var buf bytes.Buffer
		if err := r.RenderInvitation(&buf, inv); err != nil {
			t.Fatalf("RenderInvitation error: %v", err)
		}

		html := buf.String()

		if !strings.Contains(html, `inv-quote-section inv-block inv-block--contrast`) {
			t.Errorf("expected overridden quote to have inv-block--contrast")
		}
		if !strings.Contains(html, `inv-carousel-section section-gallery inv-block inv-block--contrast`) {
			t.Errorf("expected overridden carousel to have inv-block--contrast")
		}
	})
}

