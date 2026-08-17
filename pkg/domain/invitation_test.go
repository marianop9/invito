package domain

import (
	"encoding/json"
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
		Sections: []Section{
			&HeroSection{SectionType: SectionHero, Badge: "Save The Date"},
			&RSVPSection{SectionType: SectionRSVP, Enabled: true, MaxPartySize: 2},
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
	invalidCarouselInv.Sections = []Section{
		&CarouselSection{
			SectionType: SectionCarousel,
			Title:       "Our Moments",
			Images: []CarouselImage{
				{URL: "/static/img/hero.webp", Caption: "Valid photo"},
				{URL: "", Caption: "Missing URL photo"},
			},
		},
	}
	if err := invalidCarouselInv.Validate(); err == nil {
		t.Errorf("expected error when carousel image url is empty, got nil")
	}

	// Test invalid quote with empty text
	invalidQuoteInv := validInv
	invalidQuoteInv.Sections = []Section{
		&QuoteSection{SectionType: SectionQuote, Text: "   "},
	}
	if err := invalidQuoteInv.Validate(); err == nil {
		t.Errorf("expected error when quote text is empty, got nil")
	}

	// Test invalid text with empty text
	invalidTextInv := validInv
	invalidTextInv.Sections = []Section{
		&TextSection{SectionType: SectionText, Text: "   "},
	}
	if err := invalidTextInv.Validate(); err == nil {
		t.Errorf("expected error when text content is empty, got nil")
	}

	// Test invalid image with empty url
	invalidImgInv := validInv
	invalidImgInv.Sections = []Section{
		&ImageSection{SectionType: SectionImage, URL: "   "},
	}
	if err := invalidImgInv.Validate(); err == nil {
		t.Errorf("expected error when image url is empty, got nil")
	}
}

func TestInvitationJSONUnmarshaling(t *testing.T) {
	t.Run("Polymorphic Array Unmarshaling with Quote, Text, and Images", func(t *testing.T) {
		jsonBlob := `{
			"version": "1.0",
			"slug": "custom-blocks-event",
			"title": "Custom Blocks Wedding",
			"date_start": "2026-09-19T16:00:00Z",
			"location": {
				"name": "Botanical Estate",
				"address": "450 Magnolia Lane"
			},
			"theme": { "id": "botanical-elegance" },
			"sections": [
				{ "type": "hero", "badge": "Special Announcement", "cover_image_url": "/static/img/demo-hero.webp" },
				{ "type": "quote", "text": "Two lives, one path.", "author": "Poet" },
				{ "type": "carousel", "title": "Photo Moments", "images": [{ "url": "/static/img/c1.webp" }, { "url": "/static/img/c2.webp" }] },
				{ "type": "details", "show_map_link": true },
				{ "type": "timeline", "title": "Schedule", "items": [{ "time": "4:00 PM", "title": "Ceremony" }] },
				{ "type": "image", "url": "/static/img/venue.webp", "caption": "The Glasshouse Conservatory" },
				{ "type": "text", "title": "Guest Logistics", "text": "Shuttle departs at 3:15 PM sharp." },
				{ "type": "dress_code", "title": "Garden Formal", "palette_hints": ["#2A4738", "#D4AF37"] },
				{ "type": "image", "url": "/static/img/swatches.webp", "caption": "Color Swatches" },
				{ "type": "rsvp", "enabled": true, "max_party_size": 2 },
				{ "type": "faqs", "title": "Questions", "items": [{ "question": "Parking?", "answer": "Valet at gate" }] },
				{ "type": "gift_registry", "message": "Gifts welcome", "links": [{ "label": "Store", "url": "https://store.com" }] },
				{ "type": "closing", "message": "See you there!", "signoff": "Love,", "hosts": "Sarah & Alex" }
			]
		}`

		var inv Invitation
		if err := json.Unmarshal([]byte(jsonBlob), &inv); err != nil {
			t.Fatalf("failed to unmarshal polymorphic blocks: %v", err)
		}

		if len(inv.Sections) != 13 {
			t.Fatalf("expected 13 sections, got %d", len(inv.Sections))
		}

		// Verify ordering and types
		expectedTypes := []SectionType{
			SectionHero,
			SectionQuote,
			SectionCarousel,
			SectionDetails,
			SectionTimeline,
			SectionImage,
			SectionText,
			SectionDressCode,
			SectionImage,
			SectionRSVP,
			SectionFAQs,
			SectionGiftRegistry,
			SectionClosing,
		}

		for i, expected := range expectedTypes {
			if inv.Sections[i].Type() != expected {
				t.Errorf("section #%d: expected type %s, got %s", i+1, expected, inv.Sections[i].Type())
			}
		}

		// Check quote contents
		quote, ok := inv.Sections[1].(*QuoteSection)
		if !ok || !strings.Contains(quote.Text, "Two lives") || quote.Author != "Poet" {
			t.Errorf("unexpected quote: %+v", inv.Sections[1])
		}

		// Check text announcement contents
		txt, ok := inv.Sections[6].(*TextSection)
		if !ok || !strings.Contains(txt.Text, "Shuttle departs") || txt.Title != "Guest Logistics" {
			t.Errorf("unexpected txt: %+v", inv.Sections[6])
		}

		// Check multiple image contents
		img1, ok := inv.Sections[5].(*ImageSection)
		if !ok || img1.URL != "/static/img/venue.webp" {
			t.Errorf("unexpected img1: %+v", inv.Sections[5])
		}
		img2, ok := inv.Sections[8].(*ImageSection)
		if !ok || img2.URL != "/static/img/swatches.webp" {
			t.Errorf("unexpected img2: %+v", inv.Sections[8])
		}

		if err := inv.Validate(); err != nil {
			t.Fatalf("unmarshaled invitation failed validation: %v", err)
		}
	})

	t.Run("Legacy Object Map Fallback Unmarshaling", func(t *testing.T) {
		legacyJSON := `{
			"version": "1.0",
			"slug": "legacy-party",
			"title": "Legacy Party",
			"date_start": "2026-09-19T16:00:00Z",
			"location": { "name": "Hall", "address": "123 Main" },
			"theme": { "id": "golden-sunset" },
			"sections": {
				"hero": { "badge": "Legacy Eyebrow" },
				"quote": { "text": "Legacy statement", "author": "Host" },
				"text": { "text": "Legacy announcement" },
				"rsvp": { "enabled": true, "max_party_size": 2 }
			}
		}`

		var inv Invitation
		if err := json.Unmarshal([]byte(legacyJSON), &inv); err != nil {
			t.Fatalf("failed to unmarshal legacy map: %v", err)
		}

		if len(inv.Sections) == 0 {
			t.Fatalf("expected legacy sections to convert to slice, got 0")
		}

		if inv.HeroSection() == nil || inv.HeroSection().Badge != "Legacy Eyebrow" {
			t.Errorf("expected hero section to be parsed from legacy map")
		}
		if inv.RSVPSection() == nil || !inv.RSVPSection().Enabled {
			t.Errorf("expected rsvp section to be parsed from legacy map")
		}
	})
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
		Sections: []Section{
			&HeroSection{
				SectionType:    SectionHero,
				Badge:          "Save The Date",
				CoverImageURL:  "/static/img/hero.webp",
				BannerImageURL: "/static/img/banner.webp",
			},
			&QuoteSection{
				SectionType: SectionQuote,
				Text:        "Two lives, one shared journey.",
				Author:      "Poet",
			},
			&CarouselSection{
				SectionType: SectionCarousel,
				Title:       "Our Journey",
				Images: []CarouselImage{
					{URL: "/static/img/carousel-1.webp", Caption: "Engagement Day", Alt: "Sarah and Alex engagement"},
					{URL: "/static/img/carousel-2.webp", Caption: "Summer in Italy"},
				},
			},
			&ImageSection{
				SectionType: SectionImage,
				URL:         "/static/img/venue.webp",
				Caption:     "The conservatory",
				Alt:         "Glasshouse venue",
			},
			&ClosingSection{
				SectionType: SectionClosing,
				Message:     "See you soon!",
				Signoff:     "With love,",
				Hosts:       "Sarah & Alex",
			},
		},
	}

	if inv.HostsDisplay() != "Sarah & Alex" {
		t.Errorf("expected 'Sarah & Alex', got %q", inv.HostsDisplay())
	}

	if inv.FormattedDateShort() != "19/09/26" {
		t.Errorf("expected '19/09/26', got %q", inv.FormattedDateShort())
	}

	if !inv.HasCoverImage() {
		t.Errorf("expected HasCoverImage() to be true")
	}
	if inv.CoverImage() != "/static/img/hero.webp" {
		t.Errorf("expected CoverImage() to return '/static/img/hero.webp', got %q", inv.CoverImage())
	}
	hero := inv.HeroSection()
	if !hero.HasCoverImage() || !hero.HasBannerImage() {
		t.Errorf("expected HeroSection HasCoverImage and HasBannerImage to be true")
	}
	if hero.TemplateName() != "partial_hero" {
		t.Errorf("expected default hero TemplateName to be 'partial_hero', got %q", hero.TemplateName())
	}

	bannerHero := HeroSection{Layout: "banner"}
	if bannerHero.TemplateName() != "partial_hero_banner" {
		t.Errorf("expected banner hero TemplateName to be 'partial_hero_banner', got %q", bannerHero.TemplateName())
	}

	imgSec := inv.Sections[3].(*ImageSection)
	if imgSec.AltText("def") != "Glasshouse venue" {
		t.Errorf("expected image alt text 'Glasshouse venue', got %q", imgSec.AltText("def"))
	}

	closingSec := inv.Sections[4].(*ClosingSection)
	if closingSec.DisplayHosts("fallback") != "Sarah & Alex" {
		t.Errorf("expected 'Sarah & Alex', got %q", closingSec.DisplayHosts("fallback"))
	}

	if !inv.HasCarousel() {
		t.Errorf("expected HasCarousel() to be true")
	}
	carousel := inv.CarouselSection()
	if carousel.ImageCount() != 2 {
		t.Errorf("expected ImageCount() == 2, got %d", carousel.ImageCount())
	}
	if !carousel.HasImages() {
		t.Errorf("expected HasImages() to be true")
	}

	img1 := carousel.Images[0]
	if img1.AltText("default") != "Sarah and Alex engagement" {
		t.Errorf("expected alt text 'Sarah and Alex engagement', got %q", img1.AltText("default"))
	}
	img2 := carousel.Images[1]
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

func TestCarouselSectionHelperMethods(t *testing.T) {
	c := &CarouselSection{
		SectionType: SectionCarousel,
		Title:       "Test Gallery",
		Images: []CarouselImage{
			{URL: "/static/img/photo1.webp"},
		},
	}

	// Default aspect ratio
	if c.AspectRatioClass() != "4-3" {
		t.Errorf("expected default aspect ratio class '4-3', got %q", c.AspectRatioClass())
	}
	if c.IsCoverFit() {
		t.Errorf("expected default IsCoverFit to be false, got true")
	}

	// 16:9
	c.AspectRatio = "16:9"
	if c.AspectRatioClass() != "16-9" {
		t.Errorf("expected aspect ratio class '16-9', got %q", c.AspectRatioClass())
	}

	// Square 1:1
	c.AspectRatio = "1:1"
	if c.AspectRatioClass() != "1-1" {
		t.Errorf("expected aspect ratio class '1-1', got %q", c.AspectRatioClass())
	}

	// 4:5 Portrait
	c.AspectRatio = "4:5"
	if c.AspectRatioClass() != "4-5" {
		t.Errorf("expected aspect ratio class '4-5', got %q", c.AspectRatioClass())
	}

	// Fit cover
	c.Fit = "cover"
	if !c.IsCoverFit() {
		t.Errorf("expected IsCoverFit to be true when Fit is 'cover', got false")
	}

	// Nil safety
	var nilCarousel *CarouselSection
	if nilCarousel.AspectRatioClass() != "4-3" {
		t.Errorf("expected nil carousel AspectRatioClass to return '4-3', got %q", nilCarousel.AspectRatioClass())
	}
	if nilCarousel.IsCoverFit() {
		t.Errorf("expected nil carousel IsCoverFit to return false, got true")
	}
}

func TestRSVPExternalSection(t *testing.T) {
	future := time.Now().Add(24 * time.Hour)
	past := time.Now().Add(-24 * time.Hour)

	ext := &RSVPExternalSection{
		SectionType:      SectionRSVPExternal,
		Enabled:          true,
		Title:            "Confirmación de Asistencia",
		Prompt:           "Completá el formulario para confirmar tu lugar:",
		FormURL:          "https://forms.gle/demo-link",
		ButtonLabel:      "Completar Formulario",
		ContributionNote: "Seña de $15.000 (Alias: cari.cumple45)",
		ReceiptNote:      "Enviá el comprobante por WhatsApp al anfitrión",
		CustomNote:       "Hasta el 20 de Agosto",
		Deadline:         &future,
	}

	if ext.Type() != SectionRSVPExternal {
		t.Errorf("expected type 'rsvp_external', got %s", ext.Type())
	}

	if ext.TemplateName() != "partial_rsvp_external" {
		t.Errorf("expected TemplateName 'partial_rsvp_external', got %s", ext.TemplateName())
	}

	if err := ext.Validate(); err != nil {
		t.Fatalf("expected valid rsvp_external section, got error: %v", err)
	}

	if !ext.IsOpen() {
		t.Errorf("expected rsvp_external to be open with future deadline")
	}

	if !ext.HasContributionNote() {
		t.Errorf("expected HasContributionNote to be true")
	}

	if !ext.HasReceiptNote() {
		t.Errorf("expected HasReceiptNote to be true")
	}

	if !ext.HasCustomNote() {
		t.Errorf("expected HasCustomNote to be true")
	}

	// Missing form_url validation error
	invalidExt := &RSVPExternalSection{
		SectionType: SectionRSVPExternal,
		Enabled:     true,
	}
	if err := invalidExt.Validate(); err == nil {
		t.Errorf("expected error when form_url is empty, got nil")
	}

	// Past deadline
	ext.Deadline = &past
	if ext.IsOpen() {
		t.Errorf("expected IsOpen to be false with past deadline")
	}

	// Disabled
	ext.Enabled = false
	ext.Deadline = nil
	if ext.IsOpen() {
		t.Errorf("expected IsOpen to be false when disabled")
	}

	// Invitation helper methods
	inv := &Invitation{
		Slug: "test-rsvp-ext",
		Sections: []Section{
			ext,
		},
	}
	if inv.RSVPExternalSection() != ext {
		t.Errorf("expected inv.RSVPExternalSection() to return ext")
	}
}
