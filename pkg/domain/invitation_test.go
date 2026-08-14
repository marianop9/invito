package domain

import (
	"strings"
	"testing"
	"time"
)

func TestInvitationValidation(t *testing.T) {
	now := time.Now().Add(24 * time.Hour)
	end := now.Add(4 * time.Hour)

	validInv := Invitation{
		Version:   "1.0",
		Slug:      "test-party",
		Title:     "Test Party Celebration",
		DateStart: now,
		DateEnd:   &end,
		Location: Location{
			Name:    "Grand Ballroom",
			Address: "123 Main St, City, Country",
		},
		Theme: ThemeConfig{
			ID: ThemeBotanicalElegance,
		},
		Sections: Sections{
			RSVP: &RSVPSection{
				Enabled:      true,
				MaxPartySize: 2,
			},
		},
	}

	if err := validInv.Validate(); err != nil {
		t.Fatalf("expected valid invitation, got error: %v", err)
	}

	// Test invalid slug with spaces
	invalidSlug := validInv
	invalidSlug.Slug = "invalid slug with spaces"
	if err := invalidSlug.Validate(); err == nil {
		t.Errorf("expected error for invalid slug, got nil")
	}

	// Test invalid date_end before date_start
	invalidDates := validInv
	badEnd := now.Add(-1 * time.Hour)
	invalidDates.DateEnd = &badEnd
	if err := invalidDates.Validate(); err == nil {
		t.Errorf("expected error when date_end is before date_start, got nil")
	}
}

func TestInvitationHelpers(t *testing.T) {
	date := time.Date(2026, time.September, 19, 16, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.September, 19, 23, 30, 0, 0, time.UTC)

	inv := Invitation{
		Slug:      "sarah-and-alex",
		Title:     "Sarah & Alex Wedding",
		Hosts:     []string{"Sarah", "Alex"},
		DateStart: date,
		DateEnd:   &end,
		Location: Location{
			Name:    "Botanical Gardens",
			Address: "123 Flower Ave",
		},
	}

	if inv.HostsDisplay() != "Sarah & Alex" {
		t.Errorf("expected 'Sarah & Alex', got %q", inv.HostsDisplay())
	}

	if inv.FormattedDateShort() != "Sep 19, 2026" {
		t.Errorf("expected 'Sep 19, 2026', got %q", inv.FormattedDateShort())
	}

	calURL := inv.GoogleCalendarURL()
	if !strings.Contains(calURL, "calendar.google.com") {
		t.Errorf("expected google calendar link, got %q", calURL)
	}

	jsonld, err := inv.JSONLD()
	if err != nil {
		t.Fatalf("failed to generate JSON-LD: %v", err)
	}
	if !strings.Contains(jsonld, "EventScheduled") {
		t.Errorf("expected JSON-LD to contain EventScheduled, got %s", jsonld)
	}
}

func TestRSVPValidation(t *testing.T) {
	sub := RSVPSubmission{
		Name:       "Jane Doe",
		Email:      "jane@example.com",
		Attending:  true,
		GuestCount: 2,
	}

	if err := sub.Validate(2); err != nil {
		t.Fatalf("expected valid RSVP, got %v", err)
	}

	// Exceed party size
	if err := sub.Validate(1); err == nil {
		t.Errorf("expected error when guest count exceeds max party size, got nil")
	}

	// Invalid email
	sub.Email = "invalid-email"
	if err := sub.Validate(2); err == nil {
		t.Errorf("expected error for invalid email, got nil")
	}
}
