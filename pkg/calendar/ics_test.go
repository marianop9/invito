package calendar

import (
	"strings"
	"testing"
	"time"

	"invitation/pkg/domain"
)

func TestGenerateICS(t *testing.T) {
	date := time.Date(2026, time.September, 19, 16, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.September, 19, 23, 30, 0, 0, time.UTC)

	inv := &domain.Invitation{
		Slug:        "sarah-and-alex-wedding",
		Title:       "Sarah & Alex's Wedding Celebration",
		Description: "Join us for an evening of love, dinner, and dancing.",
		DateStart:   date,
		DateEnd:     &end,
		Location: domain.Location{
			Name:    "Willowbrook Estate",
			Address: "450 Magnolia Lane",
		},
	}

	ics := GenerateICS(inv)

	expectedSnippets := []string{
		"BEGIN:VCALENDAR",
		"VERSION:2.0",
		"BEGIN:VEVENT",
		"SUMMARY:Sarah & Alex's Wedding Celebration",
		"LOCATION:Willowbrook Estate\\, 450 Magnolia Lane",
		"DTSTART:20260919T160000Z",
		"DTEND:20260919T233000Z",
		"END:VEVENT",
		"END:VCALENDAR",
	}

	for _, snippet := range expectedSnippets {
		if !strings.Contains(ics, snippet) {
			t.Errorf("expected ICS output to contain %q, but got:\n%s", snippet, ics)
		}
	}
}
