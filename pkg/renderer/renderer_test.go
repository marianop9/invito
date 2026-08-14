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
