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

	// Test invalid carousel with empty image URL
	invalidCarouselInv := validInv
	invalidCarouselInv.Sections.Carousel = &CarouselSection{
		Title: "Our Moments",
		Images: []CarouselImage{
			{URL: "/static/img/hero.webp", Caption: "Valid photo"},
			{URL: "", Caption: "Missing URL photo"},
		},
	}
	if err := invalidCarouselInv.Validate(); err == nil {
		t.Errorf("expected error when carousel image url is empty, got nil")
	}

	// Test invalid message with empty text
	invalidMsgInv := validInv
	invalidMsgInv.Sections.Message = &MessageSection{Text: "   "}
	if err := invalidMsgInv.Validate(); err == nil {
		t.Errorf("expected error when message text is empty, got nil")
	}

	// Test invalid image with empty url
	invalidImgInv := validInv
	invalidImgInv.Sections.Image = &ImageSection{URL: "   "}
	if err := invalidImgInv.Validate(); err == nil {
		t.Errorf("expected error when image url is empty, got nil")
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
		Sections: Sections{
			Hero: &HeroSection{
				Badge:          "Save The Date",
				CoverImageURL:  "/static/img/hero.webp",
				BannerImageURL: "/static/img/banner.webp",
			},
			Message: &MessageSection{
				Text:   "Two lives, one shared journey.",
				Author: "Poet",
			},
			Carousel: &CarouselSection{
				Title: "Our Journey",
				Images: []CarouselImage{
					{URL: "/static/img/carousel-1.webp", Caption: "Engagement Day", Alt: "Sarah and Alex engagement"},
					{URL: "/static/img/carousel-2.webp", Caption: "Summer in Italy"},
				},
			},
			Image: &ImageSection{
				URL:     "/static/img/venue.webp",
				Caption: "The conservatory",
				Alt:     "Glasshouse venue",
			},
			Closing: &ClosingSection{
				Message: "See you soon!",
				Signoff: "With love,",
				Hosts:   "Sarah & Alex",
			},
		},
	}

	if inv.HostsDisplay() != "Sarah & Alex" {
		t.Errorf("expected 'Sarah & Alex', got %q", inv.HostsDisplay())
	}

	if inv.FormattedDateShort() != "Sep 19, 2026" {
		t.Errorf("expected 'Sep 19, 2026', got %q", inv.FormattedDateShort())
	}

	if !inv.HasCoverImage() {
		t.Errorf("expected HasCoverImage() to be true")
	}
	if inv.CoverImage() != "/static/img/hero.webp" {
		t.Errorf("expected CoverImage() to return '/static/img/hero.webp', got %q", inv.CoverImage())
	}
	if !inv.Sections.Hero.HasCoverImage() || !inv.Sections.Hero.HasBannerImage() {
		t.Errorf("expected HeroSection HasCoverImage and HasBannerImage to be true")
	}

	if !inv.HasMessage() {
		t.Errorf("expected HasMessage() to be true")
	}
	if !inv.HasImage() {
		t.Errorf("expected HasImage() to be true")
	}
	if inv.Sections.Image.AltText("def") != "Glasshouse venue" {
		t.Errorf("expected image alt text 'Glasshouse venue', got %q", inv.Sections.Image.AltText("def"))
	}
	if !inv.HasClosing() {
		t.Errorf("expected HasClosing() to be true")
	}
	if inv.Sections.Closing.DisplayHosts("fallback") != "Sarah & Alex" {
		t.Errorf("expected 'Sarah & Alex', got %q", inv.Sections.Closing.DisplayHosts("fallback"))
	}

	if !inv.HasCarousel() {
		t.Errorf("expected HasCarousel() to be true")
	}
	if inv.Sections.Carousel.ImageCount() != 2 {
		t.Errorf("expected ImageCount() == 2, got %d", inv.Sections.Carousel.ImageCount())
	}
	if !inv.Sections.Carousel.HasImages() {
		t.Errorf("expected HasImages() to be true")
	}

	img1 := inv.Sections.Carousel.Images[0]
	if img1.AltText("default") != "Sarah and Alex engagement" {
		t.Errorf("expected alt text 'Sarah and Alex engagement', got %q", img1.AltText("default"))
	}
	img2 := inv.Sections.Carousel.Images[1]
	if img2.AltText("default") != "Summer in Italy" {
		t.Errorf("expected caption fallback for alt text 'Summer in Italy', got %q", img2.AltText("default"))
	}
	emptyImg := CarouselImage{}
	if emptyImg.AltText("fallback") != "fallback" {
		t.Errorf("expected fallback 'fallback', got %q", emptyImg.AltText("fallback"))
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
	if !strings.Contains(jsonld, "/static/img/hero.webp") {
		t.Errorf("expected JSON-LD to contain cover image, got %s", jsonld)
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
