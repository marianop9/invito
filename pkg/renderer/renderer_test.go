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
		Sections: domain.Sections{
			Hero: &domain.HeroSection{
				Badge:         "Black Tie Only",
				ShowCountdown: true,
			},
			Timeline: []domain.TimelineItem{
				{Time: "7:00 PM", Title: "Reception"},
			},
			RSVP: &domain.RSVPSection{
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
			"Black Tie Only",
			"RSVP",
			"inv-countdown",
		}

		for _, kw := range expectedKeywords {
			if !strings.Contains(html, kw) {
				t.Errorf("expected HTML to contain %q, but missing", kw)
			}
		}
	})

	t.Run("RenderInvitation with Hero Cover Image and Carousel", func(t *testing.T) {
		weddingInv := *inv
		weddingInv.Sections.Hero = &domain.HeroSection{
			Badge:         "Wedding Celebration",
			CoverImageURL: "/static/img/demo-hero.webp",
		}
		weddingInv.Sections.Carousel = &domain.CarouselSection{
			Title: "Photo Moments",
			Images: []domain.CarouselImage{
				{URL: "/static/img/demo-carousel-1.webp", Caption: "Engagement at Big Sur"},
				{URL: "/static/img/demo-carousel-2.webp", Caption: "Tuscany Trip"},
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

	t.Run("RenderInvitation with Message, Static Image, and Closing", func(t *testing.T) {
		fullInv := *inv
		fullInv.Sections.Message = &domain.MessageSection{
			Text:   "Two lives, one shared journey.",
			Author: "Rumi",
		}
		fullInv.Sections.Image = &domain.ImageSection{
			URL:     "/static/img/venue.webp",
			Caption: "The Botanical Glasshouse",
		}
		fullInv.Sections.Closing = &domain.ClosingSection{
			Message: "We can't wait to celebrate with you!",
			Signoff: "With love,",
			Hosts:   "Sarah & Alex",
		}

		var buf bytes.Buffer
		if err := r.RenderInvitation(&buf, &fullInv); err != nil {
			t.Fatalf("RenderInvitation error: %v", err)
		}

		html := buf.String()
		expectedKeywords := []string{
			"section-message",
			"Two lives, one shared journey.",
			"Rumi",
			"section-image",
			"/static/img/venue.webp",
			"The Botanical Glasshouse",
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
