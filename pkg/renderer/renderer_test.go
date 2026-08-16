package renderer

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"invitation/pkg/domain"
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
			"An evening of elegance",
			"Black Tie Only",
			"RSVP",
			"inv-hero-meta",
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
			"Engagement at Big Sur",
		}

		for _, kw := range expectedImgKeywords {
			if !strings.Contains(html, kw) {
				t.Errorf("expected rendered HTML to contain %q, but missing", kw)
			}
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
			"An intimate evening under the stars",
			"Sep 19, 2026",
			"inv-hero-banner-scroll-cue",
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
}
